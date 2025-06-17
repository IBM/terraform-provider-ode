// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package httpz

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

type P struct{ A string }

func TestNewAndHeader(t *testing.T) {
	req, err := New(
		t.Context(),
		"GET",
		"https://fake.com",
		Header("Header-Test", "ok"),
	)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got := req.Header.Get("Header-Test"); got != "ok" {
		t.Fatalf("header not set: %q", got)
	}
	if _, err = New(t.Context(), "GET", ":", nil); !errors.Is(err, ErrInvalidURL) {
		t.Fatalf("want: ErrInvalidURL, got %v", err)
	}
	if _, err = New(t.Context(), "GET", "https://ex", Header("\n", "")); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("expected ErrInvalidParam for bad header")
	}
}

func TestQueryAndQueryMap(t *testing.T) {
	req, _ := New(t.Context(), "GET", "https://ex.com", QueryMap(map[string]string{"key": "value"}))
	if got := req.URL.Query().Get("key"); got != "value" {
		t.Fatalf("want query not encoded, got %q", got)
	}

	_, err := New(t.Context(), "GET", "https://fake.com", Query(url.Values{"": {"fake"}}))
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("want: invalid param, got %v", err)
	}
}

func TestJSONOpt(t *testing.T) {
	req, _ := New(
		t.Context(), "POST", "https://fake",
		JSON(P{A: "body"}),
	)
	if content := req.Header.Get("Content-Type"); content != HkJSON {
		t.Fatalf("wrong content %s", content)
	}
	var p P
	if err := json.NewDecoder(req.Body).Decode(&p); err != nil || p.A != "body" {
		t.Fatalf("decode failed: %v / %+v", err, p)
	}
}

func TestFormOpt(t *testing.T) {
	req, _ := New(t.Context(), "POST", "https://fake", Form(map[string]string{"foo": "bar"}))
	if req.Header.Get("Content-Type") != HkFORM {
		t.Fatalf("ct wrong")
	}
	b, _ := io.ReadAll(req.Body)
	if string(b) != "foo=bar" {
		t.Fatalf("bad body: %s", b)
	}
}

func TestMultipartSuccess(t *testing.T) {
	req, err := New(
		t.Context(), "POST", "http://fake",
		Multipart(
			map[string]string{"formField1": "v1"},
			[]File{{Field: "file", Name: "name.txt", R: strings.NewReader("hello")}},
		),
	)
	if err != nil {
		t.Fatalf("build err: %v", err)
	}
	mt, params, _ := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if !strings.HasPrefix(mt, "multipart/") {
		t.Fatalf("not multipart")
	}
	r := multipart.NewReader(req.Body, params["boundary"])

	p1, _ := r.NextPart()
	if p1.FormName() != "formField1" {
		t.Fatalf("want field f1")
	}
	v1, _ := io.ReadAll(p1)
	if string(v1) != "v1" {
		t.Fatalf("bad value")
	}

	p2, _ := r.NextPart()
	data, _ := io.ReadAll(p2)
	if string(data) != "hello" {
		t.Fatalf("file contents wrong")
	}
}

func TestMultipartContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	req, err := New(
		ctx, "POST", "http://faek",
		Multipart(nil, []File{{Field: "file", Name: "x", R: strings.NewReader("zzz")}}),
	)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	cancel()
	_, err = io.Copy(io.Discard, req.Body)
	if err == nil {
		t.Fatalf("expected error from cancelled ctx")
	}
}

func TestMultipartBodyCloseUnblocks(t *testing.T) {
	req, _ := New(
		t.Context(), "POST", "http://fake",
		Multipart(nil, []File{{Field: "file", Name: "x", R: strings.NewReader(strings.Repeat("x", 1024))}}),
	)
	done := make(chan struct{})
	go func() {
		_, err := io.Copy(io.Discard, req.Body)
		if err != nil {
			t.Errorf("failure to copy")
		}
		close(done)
	}()
	time.Sleep(10 * time.Millisecond)
	_ = req.Body.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("body close did not unblock copy")
	}
}

func TestPipeBodyCloseUnblocksWriter(t *testing.T) {
	pr, pw := io.Pipe()
	pb := &pipeBody{Reader: pr, cancel: pw.CloseWithError}

	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(io.Discard, pb)
		close(done)
	}()

	_ = pb.Close()
	<-done
}

func TestPipeBodyDoubleCloseIdempotent(t *testing.T) {
	pr, pw := io.Pipe()
	pb := &pipeBody{Reader: pr, cancel: pw.CloseWithError}

	_ = pb.Close()
	if err := pb.Close(); err != nil {
		t.Fatalf("second close returned err: %v", err)
	}
}

func TestHeaderSanitisesCRLF(t *testing.T) {
	req, _ := New(
		t.Context(), "GET", "https://fake.com",
		Header("X-Kyle", "good\r\nnbad"),
	)
	if got := req.Header.Get("X-Kyle"); got != "goodnbad" {
		t.Fatalf("CR/LF not stripped: %q", got)
	}
}

func TestContentLengthComputed(t *testing.T) {
	payload := map[string]int{"n": 1}
	req, err := New(t.Context(), "POST", "https://fake", JSON(payload))
	if err != nil {
		t.Fatalf("failure to create new context")
	}
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(payload)
	if err != nil {
		t.Fatalf("failure to create json encoder")
	}
	want := int64(buf.Len())

	if req.ContentLength != want {
		t.Fatalf("content-length, want %d got %d", want, req.ContentLength)
	}
}

func TestFormInvalidKey(t *testing.T) {
	_, err := New(
		t.Context(), "POST", "https://fake",
		Form(map[string]string{"": "v"}),
	)
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("want: invalid param, got %v", err)
	}
}

func TestQueryHelperInternalError(t *testing.T) {
	v := url.Values{"": {"a"}}
	_, err := New(t.Context(), "GET", "https://fake", Query(v))
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("invalid param expected")
	}
}

func TestMultipartCopyError(t *testing.T) {
	rdr := &errReader{n: 1, err: errors.New("bad")}
	req, _ := New(
		t.Context(), "POST", "http://fake",
		Multipart(nil, []File{{Field: "f", Name: "n", R: rdr}}),
	)

	_, err := io.Copy(io.Discard, req.Body)
	if err == nil || !strings.Contains(err.Error(), "bad") {
		t.Fatalf("want: propagated copy error, got %v", err)
	}
}

func TestMultipartInvalidFile(t *testing.T) {
	req, _ := New(
		t.Context(), "POST", "http://fake",
		Multipart(nil, []File{{Field: "", Name: "n", R: io.NopCloser(strings.NewReader("a"))}}),
	)

	_, err := io.Copy(io.Discard, req.Body)
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("want ErrInvalidParam, got: %v", err)
	}
}

func TestMultipartCtxCancelDuringFieldsLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	cancel()

	req, _ := New(
		ctx, "POST", "http://fake",
		Multipart(map[string]string{"k": "v"}, nil),
	)

	_, err := io.Copy(io.Discard, req.Body)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want: context.Canceled, got %v", err)
	}
}

func TestNewInvalidMethod(t *testing.T) {
	_, err := New(t.Context(), "not a method", "https://fake", nil)
	if err == nil {
		t.Fatal("expected error from invalid HTTP method")
	}
}

func TestQueryNilURL(t *testing.T) {
	req := &http.Request{} // URL is nil
	if err := QueryMap(map[string]string{"k": "v"})(req); !errors.Is(err, ErrInvalidURL) {
		t.Fatalf("want: ErrInvalidURL, got: %v", err)
	}
}

func TestMultipartClosesFileReader(t *testing.T) {
	buf := &closableBuf{Reader: strings.NewReader("data")}
	req, _ := New(
		t.Context(), "POST", "http://fake",
		Multipart(nil, []File{{Field: "f", Name: "n", R: buf}}),
	)

	_, _ = io.Copy(io.Discard, req.Body)
	_ = req.Body.Close()

	time.Sleep(10 * time.Millisecond)
	if !buf.closed {
		t.Fatal("reader was not closed by Multipart helper")
	}
}

type errReader struct {
	n   int
	err error
}

func (r *errReader) Read(p []byte) (int, error) {
	return 0, r.err
}

type closableBuf struct {
	io.Reader
	closed bool
}

func (c *closableBuf) Close() error { c.closed = true; return nil }
