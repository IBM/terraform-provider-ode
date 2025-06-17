// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package transport_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/transport"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type fakeRT struct {
	called bool
	req    *http.Request
	resp   *http.Response
	err    error
}

func (f *fakeRT) RoundTrip(req *http.Request) (*http.Response, error) {
	f.called = true
	f.req = req
	return f.resp, f.err
}

func TestChain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		base        http.RoundTripper
		wrappers    []transport.RoundTripWrapper
		expectErr   bool
		errContains string
		checkRT     func(*testing.T, http.RoundTripper)
	}{
		{
			name:     "order preservation",
			base:     &fakeRT{resp: &http.Response{StatusCode: http.StatusOK}},
			wrappers: []transport.RoundTripWrapper{record("A"), record("B"), record("C")},
			checkRT: func(t *testing.T, rt http.RoundTripper) {

				_, err := rt.RoundTrip(mkReq())
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				got := getCalls()
				want := []string{"C", "B", "A"}
				if len(got) != len(want) {
					t.Errorf("call count = %d; want %d", len(got), len(want))
				}
				for i := range want {
					if got[i] != want[i] {
						t.Errorf("order[%d] = %q; want %q", i, got[i], want[i])
					}
				}
			},
		},
		{
			name:     "response passthrough",
			base:     &fakeRT{resp: &http.Response{StatusCode: http.StatusTeapot}},
			wrappers: nil,
			checkRT: func(t *testing.T, rt http.RoundTripper) {
				resp, err := rt.RoundTrip(mkReq())
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if resp.StatusCode != http.StatusTeapot {
					t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusTeapot)
				}
			},
		},
		{
			name:     "error propagation",
			base:     &fakeRT{err: errors.New("bad")},
			wrappers: nil,
			checkRT: func(t *testing.T, rt http.RoundTripper) {
				_, err := rt.RoundTrip(mkReq())
				if err == nil || !strings.Contains(err.Error(), "bad") {
					t.Errorf("error = %v; want contains 'bad'", err)
				}
			},
		},
		{
			name:     "nil base uses default",
			base:     nil,
			wrappers: nil,
			checkRT: func(t *testing.T, rt http.RoundTripper) {
				if rt != http.DefaultTransport {
					t.Errorf("RoundTripper = %T; want http.DefaultTransport", rt)
				}
			},
		},
		{
			name:        "nil wrapper error",
			base:        &fakeRT{},
			wrappers:    []transport.RoundTripWrapper{nil},
			expectErr:   true,
			errContains: "wrapper is nil at index 0",
			checkRT:     nil,
		},
	}

	for _, tc := range tests {
		t.Run(
			tc.name, func(t *testing.T) {
				rt, err := transport.Chain(tc.base, tc.wrappers...)
				if tc.expectErr {
					if err == nil {
						t.Fatal("expected error; got nil")
					}
					if !strings.Contains(err.Error(), tc.errContains) {
						t.Errorf("error = %q; want contains %q", err.Error(), tc.errContains)
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected Chain error: %v", err)
				}

				tc.checkRT(t, rt)
			},
		)
	}
}

var calls []string

func record(name string) transport.RoundTripWrapper {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(
			func(req *http.Request) (*http.Response, error) {
				calls = append(calls, name)
				return next.RoundTrip(req)
			},
		)
	}
}

func getCalls() []string {
	defer func() { calls = nil }()
	return calls
}

func mkReq() *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "http://fake.com", nil)
	return req
}
