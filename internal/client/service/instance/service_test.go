// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package instance

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/httpz"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/ssh"
)

// cool way to test without forcing https
func init() { httpz.SetAllowInsecure(true) }

type mockApiClient struct {
	base   string
	client *http.Client
}

func (s *mockApiClient) BaseURL() string          { return s.base }
func (s *mockApiClient) HTTPClient() *http.Client { return s.client }

type fakeSSH struct {
	fail bool
	call int32
}

func (f *fakeSSH) AuthenticateInstance(_ context.Context, _ string, _ string, _ string) (ssh.Token, error) {
	atomic.AddInt32(&f.call, 1)
	if f.fail {
		return ssh.Token{}, errors.New("ssh fail")
	}
	return ssh.Token{Value: "tok"}, nil
}

func (f *fakeSSH) AuthenticateKey(_ context.Context, host string, username, pass string, key []byte) (ssh.Token, error) {
	atomic.AddInt32(&f.call, 1)
	if f.fail {
		return ssh.Token{}, errors.New("ssh  key file failed")
	}
	return ssh.Token{Value: "tok"}, nil
}

type fakeMon struct {
	fail bool
	call int32
}

func (f *fakeMon) MonitorInstance(_ context.Context, _ string) error {
	atomic.AddInt32(&f.call, 1)
	if f.fail {
		return errors.New("mon fail")
	}
	return nil
}

func goodInput() CreateInput {
	return CreateInput{
		Request: LinuxProvisionRequest{
			Emulator: Emulator{
				Ziip: 0,
				CP:   3,
				Ram:  685434,
			},
			General: General{
				Label:        "inst",
				TargetUUID:   "t1",
				ImageUUID:    "img",
				SSHPublicKey: "ssh-rsa AAA...",
			},
			ValidateLinux: true,
		},
		Auth: SSHCredentials{
			Username: "root",
			Password: "pw",
		},
	}
}

func TestCreate_Happy(t *testing.T) {
	var body bytes.Buffer
	var headerOK bool

	srv := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != pathProvision || r.Method != http.MethodPost {
					t.Fatalf("bad path %s", r.URL.Path)
				}
				h := r.Header.Get(headerSSHAuth)
				headerOK = strings.Contains(h, `"ssh-token":"tok"`)
				io.Copy(&body, r.Body)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"uuid":"xyz"}`))
			},
		),
	)
	t.Cleanup(srv.Close)

	svc := &Service{
		provider: &mockApiClient{base: srv.URL, client: srv.Client()},
		sshSvc:   &fakeSSH{},
		monSvc:   &fakeMon{},
	}

	id, err := svc.Create(context.Background(), goodInput())
	if err != nil || id != "xyz" || !headerOK {
		t.Fatalf("create fail: %v headerOK=%v", err, headerOK)
	}
	if strings.Contains(body.String(), "TargetPass") {
		t.Fatalf("password leaked into body")
	}
}

func TestCreate_Errors(t *testing.T) {
	// ssh auth error
	svc := &Service{sshSvc: &fakeSSH{fail: true}}
	if _, err := svc.Create(context.Background(), goodInput()); err == nil {
		t.Fatalf("expected ssh error")
	}

	// server returns 400
	srv := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, `error`, http.StatusBadRequest)
			},
		),
	)
	t.Cleanup(srv.Close)
	svc = &Service{
		provider: &mockApiClient{base: srv.URL, client: srv.Client()},
		sshSvc:   &fakeSSH{},
		monSvc:   &fakeMon{},
	}
	if _, err := svc.Create(context.Background(), goodInput()); err == nil {
		t.Fatalf("expected server error")
	}

	// monitor error
	srvOK := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"uuid":"u"}`))
			},
		),
	)
	t.Cleanup(srvOK.Close)
	svc = &Service{
		provider: &mockApiClient{base: srvOK.URL, client: srvOK.Client()},
		sshSvc:   &fakeSSH{},
		monSvc:   &fakeMon{fail: true},
	}
	if _, err := svc.Create(context.Background(), goodInput()); err == nil {
		t.Fatalf("expected monitor error")
	}
}

func TestGet_Delete(t *testing.T) {
	// mock server that returns 200 on delete and get
	srv := httptest.NewServer(http.NewServeMux())
	t.Cleanup(srv.Close)

	mux := srv.Config.Handler.(*http.ServeMux)
	mux.HandleFunc(
		pathRead+"1234", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"provision_uuid":"1234","successful":true}`))
		},
	)
	mux.HandleFunc(
		pathDeprovision+"1234", func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if q.Get("force") != "true" || q.Get("resume") != "false" {
				t.Fatalf("bad query %v", q)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"uuid":"1234"}`))
		},
	)

	svc := &Service{
		provider: &mockApiClient{base: srv.URL, client: srv.Client()},
		sshSvc:   &fakeSSH{},
		monSvc:   &fakeMon{},
	}

	if _, err := svc.Get(context.Background(), "1234"); err != nil {
		t.Fatalf("get failed %v", err)
	}

	if err := svc.Delete(
		context.Background(), DeleteInput{
			ProvisionUUID: "1234", Force: true, Resume: false,
		},
	); err != nil {
		t.Fatalf("delete failed %v", err)
	}
}

func TestDelete_MonitorFail(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"uuid":"d1"}`))
			},
		),
	)
	t.Cleanup(srv.Close)

	svc := &Service{
		provider: &mockApiClient{base: srv.URL, client: srv.Client()},
		sshSvc:   &fakeSSH{},
		monSvc:   &fakeMon{fail: true},
	}

	err := svc.Delete(context.Background(), DeleteInput{ProvisionUUID: "d1"})
	if err == nil {
		t.Fatalf("expected monitor fail on delete")
	}
}
