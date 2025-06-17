// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package transport_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/transport"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/transport/testutil"
)

func TestWithLogging(t *testing.T) {

	type wantEntry struct{ Level, Msg string }
	tests := []struct {
		name        string
		next        roundTripperFunc
		expectError bool
		want        []wantEntry
	}{
		{
			name: "success",
			next: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
			},
			expectError: false,
			want: []wantEntry{
				{Level: "DEBUG", Msg: "request started"},
				{Level: "DEBUG", Msg: "request completed"},
			},
		},
		{
			name:        "failure",
			next:        func(req *http.Request) (*http.Response, error) { return nil, errors.New("error") },
			expectError: true,
			want: []wantEntry{
				{Level: "DEBUG", Msg: "request started"},
				{Level: "ERROR", Msg: "request failed"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(
			tc.name, func(t *testing.T) {
				logger := &testutil.FakeLogger{}
				rt := transport.WithLogging(logger)(tc.next)

				req, _ := http.NewRequest(http.MethodGet, "https://fake.com", nil)
				req.Header.Set("X-Request-ID", "test-id")

				resp, err := rt.RoundTrip(req)
				if tc.expectError {
					if err == nil {
						t.Fatal("expected error; got nil")
					}
				} else {
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					if resp.StatusCode != http.StatusOK {
						t.Errorf("expected status %d; got %d", http.StatusOK, resp.StatusCode)
					}
				}

				entries := logger.All()
				if len(entries) != len(tc.want) {
					t.Fatalf("%s: expected %d log entries; got %d", tc.name, len(tc.want), len(entries))
				}

				for i, want := range tc.want {
					got := entries[i]
					if got.Level != want.Level || got.Msg != want.Msg {
						t.Errorf(
							"%s: entry %d = (%s, %q); want (%s, %q)", tc.name, i, got.Level, got.Msg, want.Level,
							want.Msg,
						)
					}
				}

				last := entries[len(entries)-1]
				if tc.expectError {
					if errVal, ok := last.Fields["error"]; !ok || fmt.Sprint(errVal) != "error" {
						t.Errorf("%s: expected error field 'error'; got %v", tc.name, errVal)
					}
				} else {
					if _, ok := last.Fields["status"]; !ok {
						t.Errorf("%s: expected status field in completion log", tc.name)
					}
				}
			},
		)
	}
}
