// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package transport

import (
	"crypto/tls"
	"errors"
	"net/http"
)

var ErrInvalidTransport = errors.New("transport: WithTLSConfig requires *http.Transport as base")

// WithTLSConfig returns a RoundTripWrapper function that takes an http.RoundTripper and a *tls.Config as input,
// and returns a new http.RoundTripper with the provided TLS configuration applied to the base transport.
func WithTLSConfig(cfg *tls.Config) RoundTripWrapper {
	return func(next http.RoundTripper) http.RoundTripper {
		base, ok := next.(*http.Transport)
		if !ok {
			return &errorTransport{err: ErrInvalidTransport}
		}
		clone := base.Clone()
		clone.TLSClientConfig = cfg
		return clone
	}
}

// errorTransport is a struct that holds an error and implements the http.RoundTripper interface.
// It's used to return an error when the WithTLSConfig function is called with invalid base transport.
type errorTransport struct {
	err error
}

// RoundTrip is a method of the errorTransport struct that implements the http.RoundTripper interface.
// It returns a nil response and the stored error, indicating that the request failed due to an invalid transport.
func (e *errorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, e.err
}
