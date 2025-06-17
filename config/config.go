// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"crypto/tls"
	"errors"
	"time"
)

// DefaultRetryMax is the maximum number of retries for failed requests.
const DefaultRetryMax = 3

// DefaultRetryMin is the minimum time to wait before retrying a failed request.
const DefaultRetryMin = 1 * time.Second

// DefaultRetryMaxWait is the maximum time to wait between retries.
const DefaultRetryMaxWait = 30 * time.Second

// DefaultRequestTimeout is the default timeout for HTTP requests.
const DefaultRequestTimeout = 30 * time.Second

// DefaultMonitorInt is the interval for monitoring the service.
const DefaultMonitorInt = 30 * time.Second

// DefaultMonitorTimeout is the timeout for monitoring the service.
const DefaultMonitorTimeout = 90 * time.Minute

// UrlRequired is an error message indicating that the base URL is required.
const UrlRequired = "base URL is required"

// UserRequired is an error message indicating that the user is required.
const UserRequired = "user is required"

// PasswordRequired is an error message indicating that the password is required.
const PasswordRequired = "pass is required"

// Config represents the configuration for the application.
type Config struct {
	BaseURL         string
	User            string
	Pass            string
	TLSSkipVerify   bool
	TLSConfig       *tls.Config
	CABundlePath    string
	RequestTimeout  time.Duration
	IdleConnTimeout time.Duration
	RetryMax        int
	RetryMin        time.Duration
	RetryMaxWait    time.Duration
	MonitorInterval time.Duration
	MonitorTimeout  time.Duration
	LogLevel        string
}

// FillDefaults sets default values for the configuration fields if they are not already set.
func (c *Config) FillDefaults() {
	if c.RequestTimeout == 0 {
		c.RequestTimeout = DefaultRequestTimeout
	}
	if c.RetryMax == 0 {
		c.RetryMax = DefaultRetryMax
	}
	if c.RetryMin == 0 {
		c.RetryMin = DefaultRetryMin
	}
	if c.RetryMaxWait == 0 {
		c.RetryMaxWait = DefaultRetryMaxWait
	}
	if c.MonitorInterval == 0 {
		c.MonitorInterval = DefaultMonitorInt
	}
	if c.MonitorTimeout == 0 {
		c.MonitorTimeout = DefaultMonitorTimeout
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}

}

// Validate checks if the required fields are set in the configuration.
func (c *Config) Validate() error {
	switch {
	case c.BaseURL == "":
		return errors.New(UrlRequired)
	case c.User == "":
		return errors.New(UserRequired)
	case c.Pass == "":
		return errors.New(PasswordRequired)
	default:
		return nil
	}
}
