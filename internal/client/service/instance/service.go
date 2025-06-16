// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package instance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/httpz"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/monitor"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/ssh"
)

const (
	pathProvision   = "/odtz/api/provision-services/v1/linux"
	pathDeprovision = "/odtz/api/provision-services/v1/linux/"
	pathRead        = "/odtz/api/provision-services/v1/provision/"
	headerSSHAuth   = "SSHAuthorization"
)

// sshAuthenticator is an interface for authenticating instances via SSH.
type sshAuthenticator interface {
	AuthenticateInstance(ctx context.Context, sysUUID, user, pass string) (ssh.Token, error)
	AuthenticateKey(ctx context.Context, sysUUID, user, pass string, key []byte) (ssh.Token, error)
}

// AuthenticateInstance authenticates an instance using the provided credentials.
type instanceMonitor interface {
	MonitorInstance(ctx context.Context, uuid string) error
}

// Service represents a service for managing z/OS instances.
type Service struct {
	provider httpz.APIClient
	sshSvc   sshAuthenticator
	monSvc   instanceMonitor
}

// New creates a new Service instance with the provided HTTP client, SSH authenticator, and instance monitor.
func New(p httpz.APIClient, s *ssh.Service, m *monitor.Service) *Service {
	return &Service{provider: p, sshSvc: s, monSvc: m}
}

// Create provisions a new instance and returns its UUID.
func (s *Service) Create(ctx context.Context, in CreateInput) (string, error) {
	var tok ssh.Token
	var err_auth error
	if in.Auth.Password != "" {
		tok, err_auth = s.sshSvc.AuthenticateInstance(ctx, in.Request.General.TargetUUID, in.Auth.Username, in.Auth.Password)
		if err_auth != nil {
			return "", err_auth
		}
	} else if in.Auth.KeyFile != nil {
		tok, err_auth = s.sshSvc.AuthenticateKey(ctx, in.Request.General.TargetUUID, in.Auth.Username, *in.Auth.Passphrase, *in.Auth.KeyFile)
		if err_auth != nil {
			return "", err_auth
		}
	} else {
		return "", fmt.Errorf("invalid input for ssh_target_password or ssh_target_key_file. ")
	}

	provision, err := provisionZInstance(ctx, in.Request, s.provider, tok)
	if err != nil {
		return "", err
	}

	if err = s.monSvc.MonitorInstance(ctx, provision.UUID); err != nil {
		return provision.UUID, err
	}
	return provision.UUID, nil
}

// Get retrieves details for the specified instance.
func (s *Service) Get(ctx context.Context, uuid string) (Data, error) {
	return httpz.Do[Data](ctx, s.provider, http.MethodGet, pathRead+uuid)
}

// Delete deprovisions the specified instance.
func (s *Service) Delete(ctx context.Context, in DeleteInput) error {
	params := map[string]string{
		"force":  strconv.FormatBool(in.Force),
		"resume": strconv.FormatBool(in.Resume),
	}
	resp, err := httpz.Do[DeleteResp](
		ctx, s.provider,
		http.MethodDelete,
		pathDeprovision+in.ProvisionUUID,
		httpz.QueryMap(params),
	)
	if err != nil {
		return err
	}
	return s.monSvc.MonitorInstance(ctx, resp.UUID)
}

// Update is not implemented for this service.
func (s *Service) Update(_ context.Context, _ string, _ interface{}) error {
	return errors.New("not implemented")
}

// buildSSHAuth creates an SSH authentication header for the given token and target UUID.
func buildSSHAuth(token, systemUUID string) (string, error) {
	auth := SSHAuth{token, systemUUID}
	sshAuthHeader, err := json.Marshal(auth)
	if err != nil {
		return "", fmt.Errorf("failed to marshal SSH auth header: %s", err)
	}

	return string(sshAuthHeader), nil
}

// provisionZInstance provisions a new instance using the provided token and request.
func provisionZInstance(ctx context.Context, in LinuxProvisionRequest, provider httpz.APIClient, token ssh.Token) (DeleteResp, error) {
	authHeader, err := buildSSHAuth(token.Value, in.General.TargetUUID)
	if err != nil {
		return DeleteResp{}, err
	}
	out, err := httpz.Do[DeleteResp](
		ctx,
		provider,
		http.MethodPost,
		pathProvision,
		httpz.JSON(in),
		httpz.Header(headerSSHAuth, authHeader),
	)
	if err != nil {
		return DeleteResp{}, err
	}
	return out, nil
}
