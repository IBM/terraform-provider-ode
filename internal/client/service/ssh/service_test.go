// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package ssh

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/httpz"
)

func init() { httpz.SetAllowInsecure(true) }

type stub struct {
	base string
	hc   *http.Client
}

func (s *stub) BaseURL() string          { return s.base }
func (s *stub) HTTPClient() *http.Client { return s.hc }

func newTestServer(t *testing.T) (*httptest.Server, *bytes.Buffer, *string, *string, *string) {
	var body bytes.Buffer
	var path, ct, accept string
	mux := http.NewServeMux()
	mux.HandleFunc(
		"/", func(w http.ResponseWriter, r *http.Request) {
			body.Reset()
			path, ct, accept = r.URL.Path, r.Header.Get("Content-Type"), r.Header.Get("Accept")
			io.Copy(&body, r.Body)
			switch path {
			case v2Path:
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"token":"mockToken"}`))
			case v1Path:
				http.Error(w, "forbidden", http.StatusForbidden)
			case keyPath:
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"token":"fakeToken"}`))
			case v2keyPath:
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"token":"fakeToken"}`))
			default:
				http.NotFound(w, r)
			}
		},
	)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &body, &path, &ct, &accept
}

func TestService(t *testing.T) {
	srv, body, path, ct, accept := newTestServer(t)
	svc := New(&stub{srv.URL, srv.Client()})

	t.Run(
		"TargetOK", func(t *testing.T) {
			body.Reset()
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			tok, err := svc.AuthenticateTarget(ctx, "h", "u", "p", 2222)
			if err != nil || tok.Value != "mockToken" {
				t.Errorf("got: %v, %q, want nil, %q", err, tok.Value, "xyz")
			}
			if *path != v2Path {
				t.Errorf("path %q, want %q", *path, v2Path)
			}
			var m struct {
				Hostname string
				Port     int
			}
			if err := json.Unmarshal(body.Bytes(), &m); err != nil || m.Hostname != "h" || m.Port != 2222 {
				t.Errorf("body hostname=%q, port=%d, want h, 2222", m.Hostname, m.Port)
			}
		},
	)

	t.Run(
		"Instance403", func(t *testing.T) {
			body.Reset()
			_, err := svc.AuthenticateInstance(context.Background(), "uuid", "u", "p")
			if err == nil {
				t.Errorf("got: nil, want: error")
			}
			if *path != v1Path {
				t.Errorf("path %q, want: %q", *path, v1Path)
			}
		},
	)

	t.Run(
		"KeyMultipartOK", func(t *testing.T) {
			body.Reset()
			const keyData = "dummy-key"
			tok, err := svc.AuthenticateKey(context.Background(), "uuid", "u", "p", []byte(keyData))
			if err != nil || tok.Value != "fakeToken" {
				t.Errorf("got: %v, %q, want: nil, %q", err, tok.Value, "k")
			}
			if *path != keyPath {
				t.Errorf("path %q, want: %q", *path, keyPath)
			}
			if !strings.HasPrefix(*ct, "multipart/") {
				t.Errorf("content-type %q, want: multipart/*", *ct)
			}
			if *accept != "application/json" {
				t.Errorf("accept %q, want: application/json", *accept)
			}
			bodyStr := body.String()
			if !strings.Contains(bodyStr, `"system-uuid":"uuid"`) || !strings.Contains(bodyStr, keyData) {
				t.Errorf("body missing creds or key: %s", bodyStr)
			}
		},
	)
	t.Run(
		"KeyMultipartOK", func(t *testing.T) {
			body.Reset()
			const keyData = "dummy-key"

			tok, err := svc.AuthenticateTargetWithKey(context.Background(), "h", "u", "pass", 22, []byte(keyData))
			if err != nil || tok.Value != "fakeToken" {
				t.Errorf("got: %v, %q, want: nil, %q", err, tok.Value, "k")
			}
			if *path != v2keyPath {
				t.Errorf("path %q, want: %q", *path, keyPath)
			}
			if !strings.HasPrefix(*ct, "multipart/") {
				t.Errorf("content-type %q, want: multipart/*", *ct)
			}
			if *accept != "application/json" {
				t.Errorf("accept %q, want: application/json", *accept)
			}
			bodyStr := body.String()
			if !strings.Contains(bodyStr, `"username":"u"`) || !strings.Contains(bodyStr, keyData) {
				t.Errorf("body missing creds or key: %s", bodyStr)
			}
		},
	)
}
