// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package transport

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const RequestID = "X-Request-ID"

// requestIDKey is a custom key type used to store the request ID in the context.
type requestIDKey struct{}

// requestIDTransport is a struct that wraps an existing http.RoundTripper and adds the request ID functionality.
type requestIDTransport struct {
	next http.RoundTripper
}

// WithRequestID returns a function that takes an http.RoundTripper as input and returns a new http.RoundTripper that adds the request ID to the outgoing HTTP requests.
func WithRequestID() RoundTripWrapper {
	return func(next http.RoundTripper) http.RoundTripper {
		return &requestIDTransport{next: defaultTransport(next)}
	}
}

// ContextWithRequestID adds the request ID to the given context.
func ContextWithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RoundTrip performs the actual HTTP request, adding the request ID to the outgoing HTTP requests.
func (t *requestIDTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	id := extractOrGenerateRequestID(req.Context())
	ctx := context.WithValue(req.Context(), requestIDKey{}, id)
	cloned := req.Clone(ctx)
	cloned.Header.Set(RequestID, id)
	return t.next.RoundTrip(cloned)
}

// RequestIDFromContext retrieves the request ID from the given context.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey{}).(string); ok {
		return v
	}
	return ""
}

// extractOrGenerateRequestID checks if a request ID is already present in the context. If not, it generates a new random ID.
func extractOrGenerateRequestID(ctx context.Context) string {
	if id := RequestIDFromContext(ctx); id != "" {
		return id
	}
	return generateRandomID()
}

// generateRandomID generates a random 16-byte ID and encodes it as a hexadecimal string.
func generateRandomID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic("transport: cannot generate request ID")
	}
	return hex.EncodeToString(buf[:])
}
