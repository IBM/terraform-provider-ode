// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"crypto/tls"
	"encoding/base64"
	"io"
	"strings"
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/client"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/config"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/logger"
	testlog "github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/transport/testutil"
)

func TestNew_ValidationErrors(t *testing.T) {
	if _, err := client.New(nil); err == nil || !strings.Contains(err.Error(), "config pointer is nil") {
		t.Fatalf("want: nil‑config error, got: %v", err)
	}

	if _, err := client.New(&config.Config{}); err == nil || !strings.Contains(err.Error(), config.UrlRequired) {
		t.Fatalf("want: base URL error, got: %v", err)
	}

	if _, err := client.New(&config.Config{BaseURL: "https://fake.test"}); err == nil || !strings.Contains(
		err.Error(), config.UserRequired,
	) {
		t.Fatalf("want: user required error, got %v", err)
	}

	if _, err := client.New(&config.Config{BaseURL: "https://api.test", User: "Aman"}); err == nil || !strings.Contains(
		err.Error(), config.PasswordRequired,
	) {
		t.Fatalf("want: pass required error, got %v", err)
	}
}

func TestNew_Success(t *testing.T) {
	cfg := &config.Config{BaseURL: "https://api.test", User: "Josh", Pass: "secret"}
	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.BaseURL() != cfg.BaseURL {
		t.Errorf("got: BaseURL = %q; want BaseURL %q", c.BaseURL(), cfg.BaseURL)
	}

	if c.HTTPClient().Timeout != config.DefaultRequestTimeout {
		t.Errorf("got: default timeout = %v; want: %v", c.HTTPClient().Timeout, config.DefaultRequestTimeout)
	}

	if c.Target == nil || c.Image == nil || c.Instance == nil {
		t.Error("want: services initialised, got: nil")
	}
}

func TestOption_HTTPClient(t *testing.T) {
	want := &http.Client{Timeout: 5 * time.Second}
	cfg := &config.Config{BaseURL: "https://api.test", User: "u", Pass: "p"}
	c, err := client.New(cfg, client.WithHTTPClient(want))
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	if got := c.HTTPClient(); got != want {
		t.Errorf("got: HTTPClient ptr = %p; want %p", got, want)
	}
}

func TestBasicAuth_Header(t *testing.T) {
	var got string
	srv := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				got = r.Header.Get("Authorization")
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer srv.Close()

	cfg := &config.Config{BaseURL: srv.URL, User: "kyle", Pass: "foobar"}
	cli, _ := client.New(cfg, client.WithLogger(logger.New(io.Discard)))
	resp, err := cli.HTTPClient().Get(srv.URL)
	if err != nil {
		t.Fatalf("GET error: %v", err)
	}
	resp.Body.Close()

	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("kyle:foobar"))
	if got != want {
		t.Errorf("got: Authorization = %q; want: %q", got, want)
	}
}

func TestRetry_SucceedsAfterRetries(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				calls++
				if calls <= 2 {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer srv.Close()

	cfg := &config.Config{
		BaseURL:      srv.URL,
		User:         "u",
		Pass:         "p",
		RetryMax:     2,
		RetryMin:     1 * time.Millisecond,
		RetryMaxWait: 1 * time.Millisecond,
	}

	cli, _ := client.New(cfg, client.WithLogger(logger.New(io.Discard)))
	resp, err := cli.HTTPClient().Get(srv.URL)
	if err != nil {
		t.Fatalf("GET error: %v", err)
	}
	resp.Body.Close()

	if calls != 3 {
		t.Errorf("want: 3 attempts, got %d", calls)
	}
}

func TestRetry_Exhaustion(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.WriteHeader(http.StatusInternalServerError)
			},
		),
	)
	defer srv.Close()

	cfg := &config.Config{
		BaseURL:      srv.URL,
		User:         "u",
		Pass:         "p",
		RetryMax:     1,
		RetryMin:     1 * time.Millisecond,
		RetryMaxWait: 1 * time.Millisecond,
	}

	cli, _ := client.New(cfg, client.WithLogger(logger.New(io.Discard)))
	resp, err := cli.HTTPClient().Get(srv.URL)
	if err != nil {
		t.Fatalf("GET error: %v", err)
	}
	resp.Body.Close()

	if calls != 2 {
		t.Errorf("want: 2 attempts (1+1 retry), got: %d", calls)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("got: final status = %d; want 500", resp.StatusCode)
	}
}

func TestLogger_RequestID(t *testing.T) {
	fl := &testlog.FakeLogger{}
	var reqID string
	srv := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				reqID = r.Header.Get("X-Request-ID")
				w.WriteHeader(http.StatusNoContent)
			},
		),
	)
	defer srv.Close()

	cfg := &config.Config{BaseURL: srv.URL, User: "u", Pass: "p"}
	cli, _ := client.New(cfg, client.WithLogger(fl))
	resp, _ := cli.HTTPClient().Get(srv.URL)
	resp.Body.Close()

	if len(fl.All()) != 2 {
		t.Fatalf("want: 2 log entries, got %d", len(fl.All()))
	}
	if len(reqID) != 32 || !isHex(reqID) {
		t.Errorf("got: X-Request-ID = %q; want 32-char hex", reqID)
	}
}

func TestTLSConfig_Behaviour(t *testing.T) {
	ts := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer ts.Close()

	cases := []struct {
		name string
		cfg  *config.Config
		ok   bool
	}{
		{
			name: "default fails (500 ms timeout)",
			cfg: &config.Config{
				BaseURL:        ts.URL,
				User:           "u",
				Pass:           "p",
				RetryMax:       0,
				RequestTimeout: 500 * time.Millisecond},
			ok: false,
		},
		{
			name: "custom tls.Config",
			cfg:  &config.Config{BaseURL: ts.URL, User: "u", Pass: "p", TLSConfig: &tls.Config{InsecureSkipVerify: true}},
			ok:   true,
		},
	}

	for _, tc := range cases {
		t.Run(
			tc.name, func(t *testing.T) {
				cli, _ := client.New(tc.cfg, client.WithLogger(logger.New(io.Discard)))
				resp, err := cli.HTTPClient().Get(ts.URL)
				if tc.ok {
					if err != nil {
						t.Fatalf("want: success, got: %v", err)
					}
					resp.Body.Close()
				} else {
					if err == nil {
						t.Fatalf("want: TLS error, got: none")
					}
				}
			},
		)
	}
}

func isHex(s string) bool {
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
