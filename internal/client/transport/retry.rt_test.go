// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package transport_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/transport"
)

type sequenceTransport struct {
	calls     int
	failCount int
}

func (s *sequenceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	s.calls++
	status := http.StatusOK
	if s.calls <= s.failCount {
		status = http.StatusInternalServerError
	}
	return &http.Response{StatusCode: status, Body: http.NoBody}, nil
}

func TestRetryLogic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		failCount  int
		maxRetries int
		minWait    time.Duration
		maxWait    time.Duration
		ctxTimeout time.Duration
		wantCalls  int
		wantStatus int
		wantErr    bool
	}{
		{"no failures", 0, 3, 1 * time.Millisecond, 1 * time.Millisecond, 0, 1, http.StatusOK, false},
		{"fail twice then success", 2, 3, 1 * time.Millisecond, 1 * time.Millisecond, 0, 3, http.StatusOK, false},
		{"exceed retries", 5, 2, 1 * time.Millisecond, 1 * time.Millisecond, 0, 3, http.StatusInternalServerError, false},
		{"deadline exceeded", 5, 5, 10 * time.Millisecond, 10 * time.Millisecond, 15 * time.Millisecond, 0, 0, true},
	}

	for _, tc := range tests {
		t.Run(
			tc.name, func(t *testing.T) {
				seq := &sequenceTransport{failCount: tc.failCount}
				rt := transport.WithRetry(tc.maxRetries, tc.minWait, tc.maxWait)(seq)

				ctx := t.Context()
				if tc.ctxTimeout > 0 {
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(ctx, tc.ctxTimeout)
					defer cancel()
				}
				req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://fake.com", nil)

				resp, err := rt.RoundTrip(req)

				if tc.wantErr {
					if !errors.Is(err, context.DeadlineExceeded) {
						t.Fatalf("expected DeadlineExceeded; got %v", err)
					}
					return
				}

				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if seq.calls != tc.wantCalls {
					t.Errorf("calls = %d; want %d", seq.calls, tc.wantCalls)
				}

				if resp.StatusCode != tc.wantStatus {
					t.Errorf("status = %d; want %d", resp.StatusCode, tc.wantStatus)
				}
			},
		)
	}
}

type bodyCaptureTransport struct {
	Body []byte
}

func (b *bodyCaptureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	data, _ := io.ReadAll(req.Body)
	b.Body = data
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
}

func TestRetryClonesBody(t *testing.T) {
	t.Parallel()

	payload := "test-payload"
	req, _ := http.NewRequest(http.MethodPost, "https://fake.com", io.NopCloser(strings.NewReader(payload)))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(payload)), nil
	}

	capture := &bodyCaptureTransport{}
	rt := transport.WithRetry(1, 1*time.Millisecond, 1*time.Millisecond)(capture)

	_, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := string(capture.Body); got != payload {
		t.Errorf("cloned body = %q; want %q", got, payload)
	}
}
