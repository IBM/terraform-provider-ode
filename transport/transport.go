// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package transport

import (
	"errors"
	"fmt"
	"net/http"
)

var ErrNilWrapper = errors.New("transport: wrapper is nil")

// RoundTripWrapper is a function type that takes an http.RoundTripper as input and returns an http.RoundTripper.
// It's used to represent a wrapper function that modifies the behavior of an HTTP client.
type RoundTripWrapper func(http.RoundTripper) http.RoundTripper

// Chain is a function that takes a base http.RoundTripper and a variadic slice of RoundTripWrapper functions,
// and returns a new http.RoundTripper that applies all the provided wrappers in order.
// If the base transport is nil, it defaults to http.DefaultTransport.
func Chain(base http.RoundTripper, wrappers ...RoundTripWrapper) (http.RoundTripper, error) {
	if base == nil {
		base = http.DefaultTransport
	}
	rt := base
	for i, wrap := range wrappers {
		if wrap == nil {
			return nil, fmt.Errorf("%w at index %d", ErrNilWrapper, i)
		}
		rt = wrap(rt)
	}
	return rt, nil
}
