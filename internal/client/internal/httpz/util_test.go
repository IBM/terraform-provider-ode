// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package httpz

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type M struct{ X int }
type N struct{ V string }
type errorReader struct{}

func (errorReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestEncodeMarshalDecodeRoundtrip(t *testing.T) {
	var buf bytes.Buffer
	if err := Encode(&buf, M{X: 7}); err != nil {
		t.Fatalf("encode: %v", err)
	}
	b := buf.Bytes()
	got, err := Marshal(M{X: 9})
	if err != nil || !bytes.Contains(got, []byte(`"X":9`)) {
		t.Fatalf("want: marshal: bad %v %s", err, got)
	}
	var m M
	if err = DecodeInto(bytes.NewReader(b), &m); err != nil || m.X != 7 {
		t.Fatalf("decodeinto: %v %+v", err, m)
	}
}

func TestNilErrors(t *testing.T) {
	if _, err := Marshal(struct{ A chan int }{}); err == nil {
		t.Fatal("expected marshal error due to unsupported type")
	}
	if err := Encode(nil, 1); !errors.Is(err, ErrNilWriter) {
		t.Fatalf("expect ErrNilWriter")
	}
	if _, err := Decode[int](nil); !errors.Is(err, ErrNilReader) {
		t.Fatalf("expect ErrNilReader")
	}
}

func TestDecodeResponse(t *testing.T) {
	testBody := `{"ok":true}`
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(testBody)),
	}
	var dst struct {
		Ok bool `json:"ok"`
	}
	if err := DecodeResponse(resp, &dst); err != nil || !dst.Ok {
		t.Fatalf("decode resp: %v %+v", err, dst)
	}

	bad := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(strings.NewReader("fail")),
	}
	if err := DecodeResponse(bad, &dst); err == nil {
		t.Fatal("want error on 500")
	}
	if err := DecodeResponse(nil, &dst); !errors.Is(err, ErrNilResponse) {
		t.Fatal("expect nil response error")
	}
}

func TestDecodeGeneric(t *testing.T) {
	v, err := Decode[N](strings.NewReader(`{"V":"x"}`))
	if err != nil || v.V != "x" {
		t.Fatalf("generic decode failed: %v %#v", err, v)
	}
}

func TestMarshalErrorPath(t *testing.T) {
	ch := make(chan int)
	if _, err := Marshal(ch); err == nil {
		t.Fatal("expected error on marshal unsupported type")
	}
}

func FuzzEncodeDecodeRoundTrip(f *testing.F) {
	f.Add(`{"a":1}`)
	f.Fuzz(
		func(t *testing.T, input string) {
			var dst any
			if err := DecodeInto(strings.NewReader(input), &dst); err == nil {
				var buf bytes.Buffer
				if err = Encode(&buf, dst); err != nil {
					t.Fatalf("encode after decode: %v", err)
				}
			}
		},
	)
}

func TestDecodeIntoReaderError(t *testing.T) {
	var out struct{}
	err := DecodeInto(errorReader{}, &out)
	if err == nil {
		t.Fatalf("want: error, got nil")
	}
	if !strings.Contains(err.Error(), "httpz decode:") {
		t.Errorf("want: wrapped error, got %v", err)
	}
}

func TestDecodeIntoReaderErrorNil(t *testing.T) {
	var out struct{}
	err := DecodeInto(nil, &out)
	if !errors.Is(err, ErrNilReader) {
		t.Fatalf("want: ErrNilReader, got %v", err)
	}
}

func TestDecodeReaderError(t *testing.T) {
	v, err := Decode[struct{ X int }](errorReader{})
	if err == nil {
		t.Fatalf("want: error, got nil")
	}
	if !strings.Contains(err.Error(), "httpz decode:") {
		t.Errorf("want: wrapped error, got %v", err)
	}
	if v.X != 0 {
		t.Errorf("expewant:cted zero value, got %+v", v)
	}
}
