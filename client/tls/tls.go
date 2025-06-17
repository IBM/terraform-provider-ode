// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

type Option func(*tls.Config) error

type Params struct {
	CA         string
	SNI        string
	SkipVerify bool
}
type Builder struct {
	params *Params
	opts   []Option
}

// NewBuilder initializes a TlsBuilder.
func NewBuilder(params *Params) *Builder {
	return &Builder{params: params}
}

// WithOptions adds cfg options to the builder.
func (b *Builder) WithOptions(opts ...Option) *Builder {
	b.opts = append(b.opts, opts...)
	return b
}

// Build builds a tls.Config.
func (b *Builder) Build() (*tls.Config, error) {
	// RootCA Validation
	if b.params.CA != "" {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(b.params.CA)) {
			return nil, fmt.Errorf("invalid CA certificate: failed to parse PEM")
		}
	}

	// Init
	tlsConfig := &tls.Config{
		InsecureSkipVerify: b.params.SkipVerify,
	}

	if b.params.SNI != "" {
		tlsConfig.ServerName = b.params.SNI
	}

	if b.params.CA != "" {
		caPool := x509.NewCertPool()
		caPool.AppendCertsFromPEM([]byte(b.params.CA))
		tlsConfig.RootCAs = caPool
	} else if !b.params.SkipVerify {
		systemCAs, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("failed to load system CAs: %w", err)
		}
		tlsConfig.RootCAs = systemCAs
	}

	for _, opt := range b.opts {
		if err := opt(tlsConfig); err != nil {
			return nil, fmt.Errorf("option error: %w", err)
		}
	}

	return tlsConfig, nil
}

// WithClientCert adds a client certificate for mutual TLS.
func WithClientCert(certPEM, keyPEM string) Option {
	return func(c *tls.Config) error {
		cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
		if err != nil {
			return fmt.Errorf("invalid client cert: %w", err)
		}
		c.Certificates = []tls.Certificate{cert}
		return nil
	}
}
