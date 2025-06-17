// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package transport_test

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/transport"
)

type roundTripperMock struct{}

func (roundTripperMock) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, nil
}

func TestWithTLSConfig(t *testing.T) {

	cfg := &tls.Config{MinVersion: tls.VersionTLS12}

	customBase := &http.Transport{
		IdleConnTimeout: 30 * time.Second,
		MaxIdleConns:    10,

		DialContext: (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
	}

	tests := []struct {
		name       string
		base       http.RoundTripper
		applyTwice bool
		wantType   bool // expect *http.Transport returned
		wantErr    bool // expect ErrInvalidTransport on RoundTrip
		validate   func(*testing.T, *http.Transport)
	}{
		{
			name:     "clone and apply",
			base:     &http.Transport{},
			wantType: true,
			validate: func(t *testing.T, got *http.Transport) {
				if got.TLSClientConfig != cfg {
					t.Error("TLSClientConfig not set correctly")
				}

				orig := &http.Transport{}
				if got == orig {
					t.Error("expected new transport instance, got original")
				}
			},
		},
		{
			name:     "preserve settings",
			base:     customBase,
			wantType: true,
			validate: func(t *testing.T, got *http.Transport) {

				if got.IdleConnTimeout != customBase.IdleConnTimeout ||
					got.MaxIdleConns != customBase.MaxIdleConns {
					t.Error("existing transport settings not preserved")
				}
				if got.TLSClientConfig != cfg {
					t.Error("TLSClientConfig not applied to cloned transport")
				}
			},
		},
		{
			name:       "double application",
			base:       &http.Transport{},
			applyTwice: true,
			wantType:   true,
			validate: func(t *testing.T, got *http.Transport) {
				if got.TLSClientConfig != cfg {
					t.Error("TLSClientConfig not preserved on second application")
				}
			},
		},
		{
			name:     "invalid base",
			base:     roundTripperMock{},
			wantType: false,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(
			tc.name, func(t *testing.T) {

				wrapped := transport.WithTLSConfig(cfg)(tc.base)
				if tc.applyTwice {
					wrapped = transport.WithTLSConfig(cfg)(wrapped)
				}

				if tc.wantType {
					got, ok := wrapped.(*http.Transport)
					if !ok {
						t.Fatalf("expected *http.Transport, got %T", wrapped)
					}

					tc.validate(t, got)
				} else {

					req, _ := http.NewRequest(http.MethodGet, "http://fake.com", nil)
					_, err := wrapped.RoundTrip(req)
					if !tc.wantErr || !errors.Is(err, transport.ErrInvalidTransport) {
						t.Fatalf("expected ErrInvalidTransport, got %v", err)
					}
				}
			},
		)
	}
}
