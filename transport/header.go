// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package transport

import (
	"encoding/base64"
	"net/http"
)

const AuthorizationHeader = "Authorization"

// HeaderTransport is a struct that embeds an http.RoundTripper and an http.Header.
// It's used to add custom headers to HTTP requests.
type HeaderTransport struct {
	headers http.Header
	next    http.RoundTripper
}

// WithHeaders is a higher-order function that takes an http.Header and returns a RoundTripWrapper function.
// This wrapper function creates a new HeaderTransport instance with the provided http.RoundTripper and custom headers.
func WithHeaders(headers http.Header) RoundTripWrapper {
	return func(next http.RoundTripper) http.RoundTripper {
		return NewHeaderTransport(next, headers)
	}
}

// NewHeaderTransport creates a new HeaderTransport instance with the provided http.RoundTripper and http.Header.
// It also initializes the next field with a default transport using the defaultTransport function.
func NewHeaderTransport(next http.RoundTripper, headers http.Header) http.RoundTripper {
	return &HeaderTransport{
		next:    defaultTransport(next),
		headers: headers,
	}
}

// RoundTrip sends the HTTP request with custom headers.
// It clones the original request, adds the custom headers from the HeaderTransport instance,
// and then calls the RoundTrip method of the embedded http.RoundTripper.
func (t *HeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	for key, values := range t.headers {
		for _, value := range values {
			cloned.Header.Add(key, value)
		}
	}
	return t.next.RoundTrip(cloned)
}

// WithBasicAuth is a convenience function that generates a basic authentication header using the provided username and password.
// It encodes the credentials using Base64 and sets the Authorization header with the "Basic" scheme.
// This function returns a RoundTripWrapper created using the WithHeaders function.
func WithBasicAuth(user, pass string) RoundTripWrapper {
	creds := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	h := make(http.Header)
	h.Set(AuthorizationHeader, "Basic "+creds)
	return WithHeaders(h)
}
