// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package httpz

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	HkJSON = "application/json"
	HkFORM = "application/x-www-form-urlencoded"
)

var ErrInsecureTransport = errors.New("insecure transport: BaseURL must use https")

var allowInsecure bool

func SetAllowInsecure(v bool) { allowInsecure = v }

// AllowInsecure reports whether non-TLS transport is permitted.
func AllowInsecure() bool { return allowInsecure }

// Opt is a function type that takes an *http.Request as input and returns an error.
// It's used to customize the HTTP request before sending it.
type Opt func(*http.Request) error

// File represents a form field with a name, field name, and reader for the field's value.
type File struct {
	Field string
	Name  string
	R     io.Reader
}

// pipeBody is a struct that holds an io.Reader and a cancel function.
// It's used to manage the body of a request when using the Multipart function.
type pipeBody struct {
	io.Reader
	cancel func(error) error
}

// New creates a new HTTP request with the provided context, method, and URL.
// It applies the given options to the request.
func New(ctx context.Context, method, urlStr string, opts ...Opt) (*http.Request, error) {
	if urlStr == "" {
		return nil, ErrInvalidURL
	}
	if _, err := url.ParseRequestURI(urlStr); err != nil {
		return nil, ErrInvalidURL
	}
	req, err := http.NewRequestWithContext(ctx, method, urlStr, http.NoBody)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		if err = opt(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// header is a function that sets an HTTP request header with the given key and value.
// It validates the key and value before setting the header.
func header(key, val string) Opt {
	return func(r *http.Request) error {
		if key == "" || strings.ContainsAny(key, "\r\n") {
			return ErrInvalidParam
		}
		cleanVal := strings.ReplaceAll(val, "\r", "")
		cleanVal = strings.ReplaceAll(cleanVal, "\n", "")
		r.Header.Set(key, cleanVal)
		return nil
	}
}

// query is a function that sets the query parameters of an HTTP request.
// It validates the query parameters before setting them.
func query(v url.Values) Opt {
	return func(r *http.Request) error {
		if r.URL == nil {
			return ErrInvalidURL
		}
		q := r.URL.Query()
		for k, vs := range v {
			if k == "" {
				return ErrInvalidParam
			}
			for _, s := range vs {
				q.Add(k, s)
			}
		}
		r.URL.RawQuery = q.Encode()
		return nil
	}
}

// body is a function that sets the body of an HTTP request.
// It takes a content type and an encoding function as arguments.
func body(contentType string, encode func(io.Writer) error) Opt {
	return func(r *http.Request) error {
		var buf bytes.Buffer
		if err := encode(&buf); err != nil {
			return err
		}

		data := buf.Bytes()
		replay := func() io.ReadCloser { return io.NopCloser(bytes.NewReader(data)) }

		r.Body = replay()
		r.GetBody = func() (io.ReadCloser, error) { return replay(), nil }
		r.ContentLength = int64(len(data))

		if contentType != "" {
			r.Header.Set("Content-Type", contentType)
		}
		return nil
	}
}

// Header is a convenience function that create Opt function for setting HTTP headers, query parameters, and request bodies.
func Header(k, v string) Opt {
	return header(k, v)
}

// Query is a convenience function that create Opt function for setting HTTP headers, query parameters, and request bodies.
func Query(v url.Values) Opt {
	return query(v)
}

// QueryMap is a convenience function that create Opt function for setting HTTP headers, query parameters, and request bodies.
func QueryMap(m map[string]string) Opt {
	v := make(url.Values, len(m))
	for k, s := range m {
		v.Set(k, s)
	}
	return query(v)
}

// JSON is a convenience function that create Opt function for setting HTTP headers, query parameters, and request bodies.
func JSON(v any) Opt {
	return body(
		HkJSON, func(w io.Writer) error {
			enc := json.NewEncoder(w)
			enc.SetEscapeHTML(false)
			return enc.Encode(v)
		},
	)
}

// Form is a convenience function that create Opt function for setting HTTP headers, query parameters, and request bodies.
func Form(fields map[string]string) Opt {
	return body(
		HkFORM, func(w io.Writer) error {
			u := url.Values{}
			for k, v := range fields {
				if k == "" {
					return ErrInvalidParam
				}
				u.Set(k, v)
			}
			_, err := io.WriteString(w, u.Encode())
			return err
		},
	)
}

// Close is a method of the pipeBody struct that closes the underlying reader and writer.
// It also calls the cancel function if it's not nil.
func (p *pipeBody) Close() error {
	if p.cancel != nil {
		err := p.cancel(io.EOF)
		if err != nil {
			return err
		}
		p.cancel = nil
	}
	if c, ok := p.Reader.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// Multipart is a function that creates an Opt function for sending multipart/form-data requests.
// It takes a map of form fields and a slice of File structs as arguments.
// The function sets the request body to a multipart/form-data format,
// encoding the form fields and files according to their respective content types.
func Multipart(fields map[string]string, files []File) Opt {
	return func(r *http.Request) error {
		ctx := r.Context()

		pr, pw := io.Pipe()
		mw := multipart.NewWriter(pw)

		r.Body = &pipeBody{Reader: pr, cancel: pw.CloseWithError}
		r.Header.Set("Content-Type", mw.FormDataContentType())

		go func() {
			defer mw.Close()
			defer pw.Close()

			fail := func(err error) {
				_ = pw.CloseWithError(err)
			}

			for k, v := range fields {
				select {
				case <-ctx.Done():
					fail(ctx.Err())
					return
				default:
				}
				if err := mw.WriteField(k, v); err != nil {
					fail(err)
					return
				}
			}

			for _, f := range files {
				if f.Field == "" || f.R == nil {
					fail(ErrInvalidParam)
					return
				}
				select {
				case <-ctx.Done():
					fail(ctx.Err())
					return
				default:
				}
				part, err := mw.CreateFormFile(f.Field, filepath.Base(f.Name))
				if err != nil {
					fail(err)
					return
				}
				if _, err := io.Copy(part, f.R); err != nil {
					fail(err)
					return
				}
				if c, ok := f.R.(io.Closer); ok {
					_ = c.Close()
				}
			}
		}()

		return nil
	}
}

// URLJoin  combines api base and api path
func URLJoin(baseURL, rel string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, rel)
	return u.String(), nil
}

// Do issues an HTTP request and unmarshal the JSON body into T.
func Do[T any](ctx context.Context, cli APIClient, method, relPath string, opts ...Opt) (T, error) {
	var zero T

	if u := cli.BaseURL(); !IsTLS(u) && !AllowInsecure() {
		return zero, ErrInsecureTransport
	}

	full, err := URLJoin(cli.BaseURL(), relPath)
	if err != nil {
		return zero, err
	}

	ctx, cancel := WithDefaultTimeout(ctx, 60*time.Second)
	defer cancel()

	req, err := New(ctx, method, full, opts...)
	if err != nil {
		return zero, err
	}

	resp, err := cli.HTTPClient().Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	err = DecodeResponse(resp, &zero)
	return zero, err
}

// IsTLS returns true when the URL is https. This is for tests or local development
func IsTLS(u string) bool {
	p, err := url.Parse(u)
	if err != nil {
		return false
	}
	return strings.EqualFold(p.Scheme, "https")
}

// WithDefaultTimeout returns ctx unchanged if it already has a deadline
func WithDefaultTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {} // Return a no-op cancel function if ctx is unchanged
	}
	return context.WithTimeout(ctx, d)
}
