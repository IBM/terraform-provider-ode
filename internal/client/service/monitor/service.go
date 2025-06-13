// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package monitor

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/httpz"
)

var (
	ErrFailed      = errors.New("monitor: reported failure")
	ErrNoData      = errors.New("monitor: no data found in stream")
	ErrEventFailed = errors.New("monitor: failure in event")
	ErrTimeout     = errors.New("monitor: timeout")
)

const (
	pathMonitorTargetStatus           = "/odtz/api/monitor-services/v1/target-status/"
	pathMonitorInstanceStatus         = "/odtz/api/monitor-services/v1/provision-status/"
	acceptEventStream                 = "text/event-stream"
	dataPrefix                        = "data:"
	doesNotExistMsg                   = "does not exist"
	OverallPercentageDone     float64 = 100
)

// Service represents a service for monitoring targets or instances.
type Service struct {
	provider httpz.APIClient
	interval time.Duration
	timeout  time.Duration
}

// New creates a new Service instance with the provided HTTP client, interval, and timeout.
func New(client httpz.APIClient, interval, timeout time.Duration) *Service {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	if timeout <= 0 {
		timeout = 90 * time.Minute
	}
	return &Service{provider: client, interval: interval, timeout: timeout}
}

// MonitorTarget monitors a target with the given UUID.
func (s *Service) MonitorTarget(ctx context.Context, uuid string) error {
	return s.wait(ctx, s.provider.BaseURL()+pathMonitorTargetStatus+uuid)
}

// MonitorInstance monitors an instance with the given UUID.
func (s *Service) MonitorInstance(ctx context.Context, uuid string) error {
	return s.wait(ctx, s.provider.BaseURL()+pathMonitorInstanceStatus+uuid)
}

// wait checks for updates from the monitoring service at the specified interval.
func (s *Service) wait(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		data, err := s.readStream(ctx, url)
		if err != nil {
			return err
		}

		if data.Error != nil {
			if err = evaluateData(data); err != nil {
				if isIgnorableError(err) {
					return nil
				}
				return err
			}
		}

		if data.Done {
			if data.OverallPercentage == 0 || data.OverallPercentage == OverallPercentageDone {
				return nil
			}
			return ErrEventFailed
		}

		err = waitOrTimeout(ctx, ticker)
		if err != nil {
			return err
		}
	}
}

// readStream reads the event stream from the monitoring service at the given URL.
func (s *Service) readStream(ctx context.Context, urlStr string) (Data, error) {
	req, err := httpz.New(
		ctx,
		http.MethodGet,
		urlStr,
		httpz.Header("Accept", acceptEventStream),
	)
	if err != nil {
		return Data{}, fmt.Errorf("monitor: creating request: %w", err)
	}
	req.Close = true

	resp, err := s.provider.HTTPClient().Do(req)
	if err != nil {
		return Data{}, fmt.Errorf("monitor: executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Data{}, decodeError(resp)
	}
	return parseEvent(resp.Body)
}

// waitOrTimeout determines if we should exit or continue the poll
func waitOrTimeout(ctx context.Context, ticker *time.Ticker) error {
	select {
	case <-ctx.Done():
		return ErrTimeout
	case <-ticker.C:
		return nil
	}
}

// parseEvent parses the event stream and returns the Data of "data" line in the response.
func parseEvent(r io.Reader) (Data, error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); strings.HasPrefix(line, dataPrefix) {
			jsonData := strings.TrimSpace(strings.TrimPrefix(line, dataPrefix))
			var respData Data
			if err := json.Unmarshal([]byte(jsonData), &respData); err != nil {
				return Data{}, fmt.Errorf("decoding status response: %w", err)
			}
			return respData, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return Data{}, fmt.Errorf("error reading Stream: %w", err)
	}
	return Data{}, ErrNoData
}

// evaluateData checks the data and returns a terminal error if any.
func evaluateData(data Data) error {
	msg := data.Error.Message
	if decoded, err := url.QueryUnescape(msg); err == nil {
		msg = decoded
	}
	if strings.Contains(msg, doesNotExistMsg) {
		return fmt.Errorf("ignorable: %s", msg)
	}
	return fmt.Errorf("%w (code %d): %s", ErrFailed, data.Error.Code, msg)
}

// decodeError converts raw JSON error data into a human readable message.
func decodeError(resp *http.Response) error {
	var er ErrorRecord
	_ = json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&er)

	msg := er.Message
	if decoded, err := url.QueryUnescape(msg); err == nil {
		msg = decoded
	}
	return fmt.Errorf("%w (%d): %s", ErrFailed, resp.StatusCode, msg)
}

// isIgnorableError checks if an error in the monitor payload is ignorable
func isIgnorableError(err error) bool {
	return strings.Contains(err.Error(), doesNotExistMsg)
}
