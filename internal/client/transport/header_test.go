// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package transport_test

import (
	"net/http"
	"testing"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/transport"
)

type headerCapture struct {
	Headers http.Header
}

func (h *headerCapture) RoundTrip(req *http.Request) (*http.Response, error) {
	h.Headers = req.Header.Clone()
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       http.NoBody,
	}, nil
}

func TestWithHeadersAndBasicAuth(t *testing.T) {

	tests := []struct {
		name    string
		wrapper transport.RoundTripWrapper
		expects map[string][]string
	}{
		{
			name:    "static headers",
			wrapper: transport.WithHeaders(http.Header{"Test-Foo": {"bar"}, "Test-Bar": {"baz"}}),
			expects: map[string][]string{"Test-Foo": {"bar"}, "Test-Bar": {"baz"}},
		},
		{
			name:    "basic auth header",
			wrapper: transport.WithBasicAuth("admin", "secret"),
			expects: map[string][]string{"Authorization": {"Basic YWRtaW46c2VjcmV0"}},
		},
	}

	for _, tc := range tests {
		t.Run(
			tc.name, func(t *testing.T) {
				capture := &headerCapture{}
				rt := tc.wrapper(capture)

				req, _ := http.NewRequest(http.MethodGet, "https://fake.com", nil)
				resp, err := rt.RoundTrip(req)
				if err != nil {
					t.Fatalf("RoundTrip returned unexpected error: %v", err)
				}

				if resp.StatusCode != http.StatusOK {
					t.Errorf("expected status %d; got %d", http.StatusOK, resp.StatusCode)
				}

				for key, want := range tc.expects {
					got := capture.Headers.Values(key)
					if len(got) != len(want) {
						t.Errorf("header %q: expected %v; got %v", key, want, got)
						continue
					}
					for i := range want {
						if got[i] != want[i] {
							t.Errorf("header %q[%d]: expected %q; got %q", key, i, want[i], got[i])
						}
					}
				}
			},
		)
	}
}
