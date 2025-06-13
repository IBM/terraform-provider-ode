// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package ssh

import (
	"bytes"
	"context"
	"net/http"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/httpz"
)

const (
	v2Path              = "/odtz/api/auth-services/v2/authenticate-ssh"
	v1Path              = "/odtz/api/auth-services/v1/authenticate-ssh"
	keyPath             = "/odtz/api/auth-services/v1/authenticate-ssh-key"
	v2keyPath           = "/odtz/api/auth-services/v2/authenticate-ssh-key"
	fieldSSHCreds       = "ssh-credentials"
	tarFieldSSHCreds    = "sshCredentials"
	fieldKeyFile        = "key-file"
	tarFieldKeyFile     = "keyFile"
	sshKeyFileName      = "sshkey.pem"
	headerAccept        = "Accept"
	mimeApplicationJSON = "application/json"
)

// Service represents a service for authenticating SSH connections.
type Service struct{ cli httpz.APIClient }

// New creates a new Service instance with the provided HTTP client.
func New(c httpz.APIClient) *Service { return &Service{cli: c} }

// Token represents an authentication token.
type Token struct {
	Value string `json:"token"`
}

// sshRequest is a struct representing the payload for SSH authentication requests.
type sshRequest struct {
	Hostname   string `json:"hostname,omitempty"`
	SystemUUID string `json:"system-uuid,omitempty"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Port       int    `json:"port,omitempty"`
}

// authenticate sends an authentication request to the API and returns the token.
func (s *Service) authenticate(ctx context.Context, path string, payload any) (Token, error) {
	return httpz.Do[Token](ctx, s.cli, http.MethodPost, path, httpz.JSON(payload))
}

// AuthenticateTarget authenticates an SSH connection using a hostname, username, and password.
func (s *Service) AuthenticateTarget(ctx context.Context, host, user, pass string, port int) (Token, error) {
	return s.authenticate(ctx, v2Path, sshRequest{Hostname: host, Username: user, Password: pass, Port: port})
}

// AuthenticateInstance authenticates an SSH connection using a system UUID, username, and password (v1 only).
func (s *Service) AuthenticateInstance(ctx context.Context, sysUUID, user, pass string) (Token, error) {
	return s.authenticate(ctx, v1Path, sshRequest{SystemUUID: sysUUID, Username: user, Password: pass})
}

// AuthenticateKey authenticates an SSH connection using a system UUID, username, password, and SSH key.
func (s *Service) AuthenticateKey(ctx context.Context, sysUUID, user, pass string, key []byte) (Token, error) {
	creds := sshRequest{SystemUUID: sysUUID, Username: user, Password: pass}
	data, err := httpz.Marshal(creds)
	if err != nil {
		return Token{}, err
	}

	files := []httpz.File{{
		Field: fieldKeyFile,
		Name:  sshKeyFileName,
		R:     bytes.NewReader(key),
	}}

	opts := []httpz.Opt{
		httpz.Multipart(map[string]string{fieldSSHCreds: string(data)}, files),
		httpz.Header(headerAccept, mimeApplicationJSON),
	}

	return httpz.Do[Token](ctx, s.cli, http.MethodPost, keyPath, opts...)
}

func (s *Service) AuthenticateTargetWithKey(ctx context.Context, host, username string, pass string, port int, key []byte) (Token, error) {

	creds := sshRequest{Hostname: host, Username: username, Password: pass, Port: port}
	data, err := httpz.Marshal(creds)
	if err != nil {
		return Token{}, err
	}

	files := []httpz.File{{
		Field: tarFieldKeyFile,
		Name:  sshKeyFileName,
		R:     bytes.NewReader(key),
	}}

	opts := []httpz.Opt{
		httpz.Multipart(map[string]string{tarFieldSSHCreds: string(data)}, files),
		httpz.Header(headerAccept, mimeApplicationJSON),
	}

	return httpz.Do[Token](ctx, s.cli, http.MethodPost, v2keyPath, opts...)
}
