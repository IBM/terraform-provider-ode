// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/httpz"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/monitor"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/ssh"
)

const (
	vtapPath      = "/odtz/api/target-environment-services/v1/linux/macvtap"
	iptablesPath  = "/odtz/api/target-environment-services/v1/linux/iptable"
	targetPath    = "/odtz/api/target-environment-services/v1/target/"
	linuxPath     = "/odtz/api/target-environment-services/v1/linux/"
	headerSSHAuth = "SSHAuthorization"
)

type Service struct {
	httpz.APIClient
	ssh sshAuthenticator
	mon monitorTarget
}

// sshAuthenticator is an interface for authenticating targets via SSH.
type sshAuthenticator interface {
	AuthenticateTarget(ctx context.Context, host, user, pass string, port int) (ssh.Token, error)
	AuthenticateTargetWithKey(ctx context.Context, host string, username string, passphrase string, port int, key []byte) (ssh.Token, error)
}

// monitorTarget is an interface for monitoring targets.
type monitorTarget interface {
	MonitorTarget(ctx context.Context, uuid string) error
}

// New creates a new Service instance with the provided HTTP client, SSH authenticator, and monitor.
func New(p httpz.APIClient, s *ssh.Service, m *monitor.Service) *Service {
	return &Service{APIClient: p, ssh: s, mon: m}
}

// CreateIPTables creates iptables rules for the specified target.
func (s *Service) CreateIPTables(ctx context.Context, in CreateTargetInput) (string, error) {
	in.Request.Automated = true

	return s.create(
		ctx, iptablesPath, in.Request.Hostname, in.Auth, in.Request.SSHPort, in.Request,
	)
}

// CreateMacVtap creates a macVtap interface for the specified target.
func (s *Service) CreateMacVtap(ctx context.Context, in MacvtapCreateInput) (string, error) {
	if err := validatePort(in.Port); err != nil {
		return "", err
	}

	tok, err := s.ssh.AuthenticateTarget(ctx, in.Hostname, in.Username, in.Password, in.Port)
	if err != nil {
		return "", err
	}

	target, err := registerTarget(ctx, vtapPath, in.Hostname, tok, in.Port, s.APIClient, in.MacVtapRequest)
	if err != nil {
		return "", err
	}
	if err = s.mon.MonitorTarget(ctx, target.UUID); err != nil {
		return target.UUID, err
	}
	return target.UUID, nil

}

// create is a helper function for creating targets with SSH authentication.
func (s *Service) create(ctx context.Context, path string, host string, auth SSHCredentials, port int, payload any) (string, error) {
	if err := validatePort(port); err != nil {
		return "", err
	}

	var tok ssh.Token
	var err_auth error
	if auth.Password != "" {
		tok, err_auth = s.ssh.AuthenticateTarget(ctx, host, auth.Username, auth.Password, port)
		if err_auth != nil {
			return "", err_auth
		}
	} else if auth.KeyFile != nil {
		tok, err_auth = s.ssh.AuthenticateTargetWithKey(ctx, host, auth.Username, *auth.Passphrase, port, *auth.KeyFile)
		if err_auth != nil {
			return "", err_auth
		}
	} else {
		return "", fmt.Errorf("invalid input for ssh_target_password or ssh_target_key_file. ")
	}

	target, err := registerTarget(ctx, path, host, tok, port, s.APIClient, payload)
	if err != nil {
		return "", err
	}
	if err = s.mon.MonitorTarget(ctx, target.UUID); err != nil {
		return target.UUID, err
	}
	return target.UUID, nil
}

// Get retrieves details for the specified target.
func (s *Service) Get(ctx context.Context, id string) (*Target, error) {
	return httpz.Do[*Target](ctx, s.APIClient, http.MethodGet, targetPath+id)
}

// Delete decommissions the specified target.
func (s *Service) Delete(ctx context.Context, id string, force, resume bool) error {
	q := url.Values{
		"force":  {strconv.FormatBool(force)},
		"resume": {strconv.FormatBool(resume)},
	}
	resp, err := httpz.Do[SuccessResponse](
		ctx, s.APIClient,
		http.MethodDelete,
		linuxPath+id,
		httpz.Query(q),
	)
	if err != nil {
		return err
	}
	return s.mon.MonitorTarget(ctx, resp.UUID)
}

// validatePort checks if the provided port number is valid.
func validatePort(p int) error {
	if p < 1 || p > 65535 {
		return fmt.Errorf("invalid port: %d", p)
	}
	return nil
}

// buildSSHAuth creates an SSH authentication header for the given token, host, and port.
func buildSSHAuth(token, host string, port int) (string, error) {
	auth := sshAuth{token, host, port}
	sshAuthHeader, err := json.Marshal(auth)
	if err != nil {
		return "", fmt.Errorf("failed to marshal SSH auth header: %s", err)
	}
	return string(sshAuthHeader), nil
}

// Send a POST request to register the target with the provided payload and authentication header.
func registerTarget(ctx context.Context, path string, host string, tok ssh.Token, port int, s httpz.APIClient, payload any) (SuccessResponse, error) {
	authHeader, err := buildSSHAuth(tok.Value, host, port)
	if err != nil {
		return SuccessResponse{}, err
	}

	out, err := httpz.Do[SuccessResponse](
		ctx, s, http.MethodPost, path,
		httpz.JSON(payload),
		httpz.Header(headerSSHAuth, authHeader),
	)
	return out, err
}
