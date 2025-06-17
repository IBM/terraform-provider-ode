// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package httpz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// maxBodySize is the maximum allowed body size for decoding JSON responses.
const maxBodySize = 10 << 20

// Encode writes the provided value to the given writer in JSON format.
// It returns an error if the writer is nil.
func Encode(w io.Writer, v any) error {
	if w == nil {
		return ErrNilWriter
	}
	return json.NewEncoder(w).Encode(v)
}

// Marshal encodes the provided value to JSON and returns the resulting byte slice.
// It returns an error if encoding fails.
func Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := Encode(&buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode reads JSON data from the given reader and decodes it into the provided value.
// It returns the decoded value and an error if decoding fails.
// If the reader is nil, it returns the zero value of the provided type and an error.
func Decode[T any](r io.Reader) (T, error) {
	var dst T
	if r == nil {
		return dst, ErrNilReader
	}
	if err := json.NewDecoder(r).Decode(&dst); err != nil {
		return dst, fmt.Errorf("httpz decode: %w", err)
	}
	return dst, nil
}

// DecodeInto reads JSON data from the given reader and decodes it into the provided output value.
// It returns an error if decoding fails.
// If the reader is nil, it returns an error.
func DecodeInto(r io.Reader, out any) error {
	if r == nil {
		return ErrNilReader
	}
	if err := json.NewDecoder(r).Decode(out); err != nil {
		return fmt.Errorf("httpz decode: %w", err)
	}
	return nil
}

// DecodeResponse decodes the JSON response body into the provided output value.
// It returns an error if decoding fails or if the response status code indicates an error.
// If the response is nil, it returns an error.
func DecodeResponse(resp *http.Response, out any) error {
	if resp == nil {
		return ErrNilResponse
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, maxBodySize)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		b, _ := io.ReadAll(limited)
		return fmt.Errorf("httpz: http %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return DecodeInto(limited, out)
}
