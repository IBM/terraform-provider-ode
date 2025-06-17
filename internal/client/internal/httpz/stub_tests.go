// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build unit
// +build unit

package httpz

import (
	"context"
	"net/url"
	"sync"
)

// --- minimal public surface the target package relies on --------------------

type APIClient interface{}

type Opt func(*opts)

type opts struct{}

// Helpers the production code calls – no-ops are enough to build.
func Header(_, _ string) Opt { return func(*opts) {} }
func JSON(_ any) Opt         { return func(*opts) {} }
func Query(_ url.Values) Opt { return func(*opts) {} }

// --- Test instrumentation ---------------------------------------------------

type Request struct {
	Method string
	Path   string
}

var (
	mu        sync.Mutex
	requests  []Request
	responses []any
	errs      []error
)

// StubResponse queues the next <T, error> pair that Do[T] will return.
func StubResponse(resp any, err error) { // FIFO
	mu.Lock()
	defer mu.Unlock()
	responses = append(responses, resp)
	errs = append(errs, err)
}

// PopRequests returns and clears the record of calls made so far.
func PopRequests() []Request {
	mu.Lock()
	defer mu.Unlock()
	out := requests
	requests = nil
	return out
}

// Generic Do[…] identical in shape to the real one, but purely in-memory.
func Do[T any](_ context.Context, _ APIClient, method, path string, _ ...Opt) (T, error) {
	mu.Lock()
	defer mu.Unlock()

	var zero T
	requests = append(requests, Request{method, path})

	if len(responses) == 0 {
		return zero, nil // default zero value if caller forgets to stub
	}
	resp, err := responses[0], errs[0]
	responses, errs = responses[1:], errs[1:]

	if err != nil {
		return zero, err
	}
	if resp == nil {
		return zero, nil
	}
	return resp.(T), nil
}
