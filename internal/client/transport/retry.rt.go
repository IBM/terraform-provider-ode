// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package transport

import (
	"net/http"
	"time"
)

const (
	retryFactor = 2
)

// retryTransport is a struct that wraps an existing http.RoundTripper and adds retry functionality.
type retryTransport struct {
	next     http.RoundTripper
	retryMax int
	minWait  time.Duration
	maxWait  time.Duration
}

// WithRetry returns a RoundTripWrapper function that wraps an http.RoundTripper with the provided retry settings.
func WithRetry(retryCount int, minWait, maxWait time.Duration) RoundTripWrapper {
	return func(next http.RoundTripper) http.RoundTripper {
		return &retryTransport{
			next:     defaultTransport(next),
			retryMax: retryCount,
			minWait:  minWait,
			maxWait:  maxWait,
		}
	}
}

// RoundTrip sends an HTTP request with retry logic.
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
		wait = t.minWait
	)
	for attempt := 0; attempt <= t.retryMax; attempt++ {
		if ctxErr := req.Context().Err(); ctxErr != nil {
			return nil, ctxErr
		}
		cloned := cloneRequest(req)
		resp, err = t.next.RoundTrip(cloned)
		if err == nil && (resp == nil || resp.StatusCode < 500) {
			return resp, nil
		}
		if attempt == t.retryMax {
			break
		}
		select {
		case <-time.After(wait):
		case <-req.Context().Done():
			return resp, req.Context().Err()
		}
		wait = nextBackoff(wait, t.maxWait)
	}
	return resp, err
}

// nextBackoff calculates the next wait time based on the current wait time and the maximum wait time, using an exponential backoff strategy.
func nextBackoff(current, maxTime time.Duration) time.Duration {
	next := current * retryFactor
	if next > maxTime {
		return maxTime
	}
	return next
}

// cloneRequest creates a copy of the provided HTTP request, including its context and body.
func cloneRequest(req *http.Request) *http.Request {
	cloned := req.Clone(req.Context())
	if req.Body == nil || req.GetBody == nil {
		return cloned
	}
	body, err := req.GetBody()
	if err != nil {
		return cloned
	}
	cloned.Body = body
	return cloned
}
