// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package transport

import (
	"net/http"
	"time"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/logger"
)

// loggingTransport is a struct that embeds a logger.Logger and an http.RoundTripper.
// It's used to log the request and response details, including the duration of the request.
type loggingTransport struct {
	logger logger.Logger
	next   http.RoundTripper
}

// WithLogging is a higher-order function that takes a logger.Logger and returns a RoundTripWrapper function.
// This wrapper function creates a new loggingTransport instance with the provided http.RoundTripper and logger.
func WithLogging(l logger.Logger) RoundTripWrapper {
	return func(next http.RoundTripper) http.RoundTripper {
		return &loggingTransport{
			logger: l,
			next:   defaultTransport(next),
		}
	}
}

// RoundTrip sends the HTTP request and logs the request and response details.
// It logs the start of the request, including the method, URL, and request ID.
// If an error occurs during the request, it logs the error and returns the error.
// If the request is successful, it logs the completion of the request, including the status code and duration.
func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	id := req.Header.Get(RequestID)
	if id == "" {
		id = "unknown"
	}
	t.logger.Debug("request started", "method", req.Method, "url", req.URL.String(), "request_id", id)
	resp, err := t.next.RoundTrip(req)
	duration := time.Since(start).Milliseconds()
	if err != nil {
		t.logger.Error("request failed", "error", err.Error(), "duration", duration, "request_id", id, "error", err)
		return nil, err
	}
	t.logger.Debug("request completed", "status", resp.StatusCode, "duration", duration, "request_id", id)
	return resp, nil
}
