// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package monitor

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

type mockClient struct {
	baseURL string
	client  *http.Client
}

func (m *mockClient) BaseURL() string          { return m.baseURL }
func (m *mockClient) HTTPClient() *http.Client { return m.client }

func newTestService(baseURL string, interval, timeout time.Duration) *Service {
	return New(
		&mockClient{
			baseURL: baseURL,
			client:  &http.Client{Timeout: 2 * time.Second},
		}, interval, timeout,
	)
}

// monitorStream builds the stream exactly like the response we get from the monitor apis.
func monitorStream(event, id, data string) string {
	var sb strings.Builder
	if event != "" {
		sb.WriteString("event: ")
		sb.WriteString(event)
		sb.WriteString("\n")
	}
	if id != "" {
		sb.WriteString("id: ")
		sb.WriteString(id)
		sb.WriteString("\n")
	}
	if data != "" {
		sb.WriteString("data: ")
		sb.WriteString(data)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	return sb.String()
}

// monitorEvent builds a single response of the monitor apis.
func monitorEvent(json string) string {
	return monitorStream("monitor", "1", json)
}

func monitorTestServer(t *testing.T, stream, status []string) *httptest.Server {
	t.Helper()
	events := stream
	codes := status
	return httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", acceptEventStream)
				code := http.StatusOK
				if len(codes) > 0 && codes[0] != "" {
					var err error
					code, err = strconv.Atoi(codes[0])
					if err != nil {
						t.Fatalf("invalid status code: %v", codes[0])
					}
					codes = codes[1:]
				}
				w.WriteHeader(code)

				if len(events) > 0 {
					fmt.Fprint(w, events[0])
					events = events[1:]
				}
			},
		),
	)
}

func TestMonitorInstance_Cases(t *testing.T) {
	tests := map[string]struct {
		stream      []string
		status      []string
		interval    time.Duration
		timeout     time.Duration
		wantErr     error
		description string
	}{
		"Successful Event Stream": {
			stream: []string{
				monitorEvent(`{"overall-percentage":10,"done":false}`),
				monitorEvent(`{"overall-percentage":100,"done":true}`),
			},
			interval:    100 * time.Millisecond,
			timeout:     2 * time.Second,
			wantErr:     nil,
			description: "Creates a successful process with multiple progress updates until completion",
		},
		"Successful Event Stream After Completion and Resetting Stream to 0": {
			stream: []string{
				monitorEvent(`{"overall-percentage":10,"done":false,"error":null}`),

				monitorEvent(`{"overall-percentage":0.0,"done":true,"error":null}`),
			},
			interval:    100 * time.Millisecond,
			timeout:     2 * time.Second,
			wantErr:     nil,
			description: "Creates a successful process with multiple progress updates until completion",
		},
		"Report Error When Event Pauses with Overall Percentage < 100 and Done is True": {
			stream: []string{
				monitorEvent(`{"overall-percentage":0.0,"done":false,"error":null}`),
				monitorEvent(`{"overall-percentage":9.0,"done":true,"error":null}`),
			},
			interval:    100 * time.Millisecond,
			timeout:     1 * time.Second,
			wantErr:     ErrEventFailed,
			description: "Returns an error if the progress stops and done is true",
		},

		"Report Error When Monitor Event Reports Error in Stream": {
			stream: []string{
				monitorEvent(`{"overall-percentage":10,"done":false}`),
				monitorEvent(`{"done":true,"error":{"code":60001,"message":"error"}}`),
			},
			interval:    100 * time.Millisecond,
			timeout:     1 * time.Second,
			wantErr:     ErrFailed,
			description: "Creates a provisioning failure with a structured error on completion",
		},
		"Error If No Data Found During Stream Scan": {
			stream:      []string{},
			interval:    100 * time.Millisecond,
			timeout:     2 * time.Second,
			wantErr:     ErrNoData,
			description: "Creates a process that never completes, expecting a timeout error",
		},
		"Timeout Error": {
			stream: []string{
				monitorEvent(`{"overall-percentage":10,"done":false}`),
				monitorEvent(`{"overall-percentage":50,"done":false}`),
			},
			interval:    5 * time.Millisecond,
			timeout:     10 * time.Millisecond,
			wantErr:     context.DeadlineExceeded,
			description: "Creates a process that never completes, expecting a timeout error",
		},
		"Last DNE Response After Delete Returns Successfully and Ignores DNE Error": {
			stream: []string{
				monitorEvent(`{"overall-percentage":40,"done":false}`),
				monitorEvent(`{"done":false,"error":{"code":50113,"message":"The%20provision%20object%20does%20not%20exist."}}`),
			},
			interval:    100 * time.Millisecond,
			timeout:     1 * time.Second,
			wantErr:     nil,
			description: "Creates a delete operation ending with an ignorable 'does not exist' error, expecting success",
		},
		"Return Bad Request During a Bad Request": {
			stream:      []string{},
			status:      []string{"400"},
			interval:    100 * time.Millisecond,
			timeout:     1 * time.Second,
			wantErr:     fmt.Errorf("monitor: reported failure (%d)", http.StatusBadRequest),
			description: "Creates a HTTP 400 response, expecting a failure with a parsed HTTP error",
		},
		"Return Failure Message During a Bad Request": {
			stream: []string{
				monitorEvent(`{"done":false,"error":{"code":50113,"message":"The%20provision%20object%20does%20not%20exist."}}`),
			},
			status:      []string{"400"},
			interval:    100 * time.Millisecond,
			timeout:     1 * time.Second,
			wantErr:     fmt.Errorf("monitor: reported failure (%d)", http.StatusBadRequest),
			description: "Creates a HTTP 400 response, expecting a failure with a parsed HTTP error",
		},
	}

	for name, tt := range tests {
		t.Run(
			name, func(t *testing.T) {
				srv := monitorTestServer(t, tt.stream, tt.status)
				defer srv.Close()

				svc := newTestService(srv.URL, tt.interval, tt.timeout)
				err := svc.MonitorInstance(t.Context(), "unit-test")

				if tt.wantErr != nil {
					if name == "Timeout Error" {
						if !strings.Contains(err.Error(), tt.wantErr.Error()) && !strings.Contains(err.Error(), ErrTimeout.Error()) {
							t.Errorf("failed: %s \nwant: %s, \ngot: %s", tt.description, tt.wantErr, err)
						}
					} else if err == nil || !strings.Contains(err.Error(), tt.wantErr.Error()) {
						t.Errorf("failed: %s \nwant: %s, \ngot: %s", tt.description, tt.wantErr, err)
					}
				} else if err != nil {
					t.Errorf("failed: %s Unexpected error: %v", tt.description, err)
				}
			},
		)
	}
}

func TestNewService_DefaultsAreApplied(t *testing.T) {
	tests := []struct {
		name         string
		intervalIn   time.Duration
		timeoutIn    time.Duration
		wantInterval time.Duration
		wantTimeout  time.Duration
	}{
		{
			name:         "Defaults applied for zero inputs",
			intervalIn:   0,
			timeoutIn:    0,
			wantInterval: 30 * time.Second,
			wantTimeout:  90 * time.Minute,
		},
		{
			name:         "Defaults applied for negative inputs",
			intervalIn:   -1,
			timeoutIn:    -1,
			wantInterval: 30 * time.Second,
			wantTimeout:  90 * time.Minute,
		},
		{
			name:         "Custom values are respected",
			intervalIn:   15 * time.Second,
			timeoutIn:    45 * time.Minute,
			wantInterval: 15 * time.Second,
			wantTimeout:  45 * time.Minute,
		},
	}

	for _, tc := range tests {
		t.Run(
			tc.name, func(t *testing.T) {
				svc := New(&mockClient{}, tc.intervalIn, tc.timeoutIn)

				if svc.interval != tc.wantInterval {
					t.Errorf("want: interval %v, got %v", tc.wantInterval, svc.interval)
				}
				if svc.timeout != tc.wantTimeout {
					t.Errorf("want: timeout %v, got %v", tc.wantTimeout, svc.timeout)
				}
			},
		)
	}
}
