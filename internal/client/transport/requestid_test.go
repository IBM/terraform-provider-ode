// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package transport_test

import (
	"context"
	"net/http"
	"regexp"
	"testing"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/transport"
)

type headerCaptureRoundTripper struct {
	Headers http.Header
	CtxID   string
}

func (h *headerCaptureRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	h.Headers = req.Header.Clone()
	h.CtxID = transport.RequestIDFromContext(req.Context())
	return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: http.NoBody}, nil
}

func TestContextHelpers(t *testing.T) {

	existingID := "ctx-123"
	ctx := transport.ContextWithRequestID(context.Background(), existingID)
	if got := transport.RequestIDFromContext(ctx); got != existingID {
		t.Errorf("RequestIDFromContext = %q; want %q", got, existingID)
	}

	if got := transport.RequestIDFromContext(context.Background()); got != "" {
		t.Errorf("RequestIDFromContext(empty) = %q; want empty string", got)
	}
}

func TestWithRequestID_RoundTrip(t *testing.T) {

	tests := []struct {
		name      string
		ctxSetup  func() context.Context
		expectsID *regexp.Regexp
	}{
		{
			name: "provided ID is propagated",
			ctxSetup: func() context.Context {
				return transport.ContextWithRequestID(context.Background(), "req-abc-123")
			},
			expectsID: regexp.MustCompile(`^req-abc-123$`),
		},
		{
			name: "generated ID is hex 32 chars",
			ctxSetup: func() context.Context {
				return context.Background()
			},
			expectsID: regexp.MustCompile(`^[a-f0-9]{32}$`),
		},
	}

	for _, tc := range tests {
		t.Run(
			tc.name, func(t *testing.T) {
				capture := &headerCaptureRoundTripper{}
				rt := transport.WithRequestID()(capture)

				req, _ := http.NewRequestWithContext(tc.ctxSetup(), http.MethodGet, "https://fake.com", nil)
				resp, err := rt.RoundTrip(req)
				if err != nil {
					t.Fatalf("RoundTrip returned error: %v", err)
				}
				if resp.StatusCode != http.StatusOK {
					t.Errorf("expected status %d; got %d", http.StatusOK, resp.StatusCode)
				}

				headerID := capture.Headers.Get("X-Request-ID")
				if !tc.expectsID.MatchString(headerID) {
					t.Errorf("header ID = %q; does not match %v", headerID, tc.expectsID)
				}
				if !tc.expectsID.MatchString(capture.CtxID) {
					t.Errorf("context ID = %q; does not match %v", capture.CtxID, tc.expectsID)
				}
			},
		)
	}
}
