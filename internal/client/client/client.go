// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/config"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/httpz"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/logger"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/image"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/instance"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/monitor"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/ssh"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/target"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/transport"
)

// _ httpz.APIClient is a type assertion that the Client struct implements the httpz.APIClient interface.
var _ httpz.APIClient = &Client{}

func (c *Client) HTTPClient() *http.Client { return c.http }
func (c *Client) BaseURL() string          { return c.cfg.BaseURL }

// Client is the main struct that holds the configuration, services, and HTTP client for the Terraform client.
type Client struct {
	http   *http.Client
	logger logger.Logger
	cfg    *config.Config

	Target   *target.Service
	Image    *image.Service
	Instance *instance.Service
}

// Option is a function type that takes a *Client as input and returns an error.
// It's used to customize the Client during its creation.
type Option func(*Client) error

// New creates a new Client instance with the provided options.
// It applies the default configuration, validates it, and sets up the services and HTTP client.
func New(cfg *config.Config, opts ...Option) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config pointer is nil")
	}

	cfg.FillDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	c := &Client{
		http: &http.Client{},
		cfg:  cfg,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	// Logger
	if c.logger == nil {
		c.logger = defaultLogger(cfg)
	}

	// TLS
	if c.cfg.TLSConfig == nil {
		c.cfg.TLSConfig = &tls.Config{}
	}

	// HTTP
	c.http.Timeout = c.cfg.RequestTimeout
	baseRT := c.http.Transport
	if baseRT == nil {
		baseRT = http.DefaultTransport
	}

	wrapper, err := transport.Chain(
		baseRT,
		transport.WithTLSConfig(c.cfg.TLSConfig),
		transport.WithBasicAuth(c.cfg.User, c.cfg.Pass),
		transport.WithRetry(c.cfg.RetryMax, c.cfg.RetryMin, c.cfg.RetryMaxWait),
		transport.WithLogging(c.logger),
		transport.WithRequestID(),
	)
	if err != nil {
		return nil, err
	}
	c.http.Transport = wrapper

	// Services.
	monitorSvc := monitor.New(c, c.cfg.MonitorInterval, c.cfg.MonitorTimeout)
	sshSvc := ssh.New(c)
	c.Image = image.New(c)
	c.Target = target.New(c, sshSvc, monitorSvc)
	c.Instance = instance.New(c, sshSvc, monitorSvc)

	return c, nil
}

// WithHTTPClient is an Option function that sets the internal http.Client used by the Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) error {
		if hc == nil {
			return fmt.Errorf("http client is nil")
		}
		c.http = hc
		return nil
	}
}

// WithLogger is an Option function that sets the logger for the Client.
func WithLogger(l logger.Logger) Option {
	return func(c *Client) error {
		if l == nil {
			return fmt.Errorf("logger is nil")
		}
		c.logger = l
		return nil
	}
}

func defaultLogger(cfg *config.Config) logger.Logger {
	level := logger.Info
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = logger.Debug
	case "warn":
		level = logger.Warn
	case "error":
		level = logger.Error
	case "trace":
		level = logger.Trace
	}

	return logger.New(
		os.Stdout,
		logger.WithFormatter(logger.Inline()),
		logger.WithColor(true),
		logger.WithMinLevel(level),
	)
}
