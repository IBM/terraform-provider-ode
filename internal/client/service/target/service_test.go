// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/httpz"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/ssh"
)

func init() { httpz.SetAllowInsecure(true) }

type stubClient struct {
	base string
	hc   *http.Client
}

func (s *stubClient) BaseURL() string          { return s.base }
func (s *stubClient) HTTPClient() *http.Client { return s.hc }

type fakeSSH struct {
	failHost string
}

func (f *fakeSSH) AuthenticateTarget(_ context.Context, host, _, _ string, _ int) (ssh.Token, error) {
	if host == f.failHost {
		return ssh.Token{}, errors.New("ssh auth failed")
	}
	return ssh.Token{Value: "tok"}, nil
}

func (f *fakeSSH) AuthenticateTargetWithKey(_ context.Context, host string, username, pass string, port int, key []byte) (ssh.Token, error) {
	if host == f.failHost {
		return ssh.Token{}, errors.New("ssh auth failed")
	}
	return ssh.Token{Value: "tok"}, nil
}

type fakeMon struct {
	fail bool
}

func (f *fakeMon) MonitorTarget(_ context.Context, _ string) error {
	if f.fail {
		return errors.New("monitor fail")
	}
	return nil
}

func goodIPT() CreateTargetInput {
	return CreateTargetInput{
		Request: IPTablesRequest{
			Automated:           true,
			ConcurrentTransfers: 6,
			DNSIPPrimary:        "9.9.9.9",
			DNSDomainOrigin:     "ex.com",
			DownloadDirectory:   "/opt",
			Hostname:            "h",
			Label:               "lab",
			NetworkType:         "IPTABLE",
			SSHPort:             22,
			ICPort:              8443,
			TerminalPortStart:   3270,
			IPTablesSetting: IPTablesSetting{
				ZosIPAddress:    "172.26.1.2",
				ZosSSHRoutePort: 2022,
				TCPForwardPorts: []ForwardPortRange{{StartPort: 0, EndPort: 21}},
				UDPForwardPorts: []ForwardPortRange{{StartPort: 111, EndPort: 111}},
				TCPReroutePorts: []ReroutePortMapping{{LinuxPort: 2022, ZosPort: 22}},
				UDPReroutePorts: []ReroutePortMapping{{LinuxPort: 2022, ZosPort: 22}},
			},
		},
		Auth: SSHCredentials{
			Username: "u",
			Password: "p",
		},
	}
}

func goodMAC() MacvtapCreateInput {
	return MacvtapCreateInput{
		MacVtapRequest: MacVtapRequest{
			Automated:           true,
			ConcurrentTransfers: 3,
			DNSIPPrimary:        "8.8.8.8",
			DNSDomainOrigin:     "ex.com",
			DownloadDirectory:   "/opt",
			Hostname:            "h2",
			Label:               "mvt",
			NetworkType:         "MACVTAP",
			SSHPort:             22,
			ICPort:              8443,
			TerminalPortStart:   3270,
			MacvtapSetting:      MacvtapSetting{DefaultRoute: "172.10.9.1"},
		},
		Username: "u",
		Password: "p",
		Port:     22,
	}
}

func goodIPTwithKey() CreateTargetInput {
	k := []byte("dummy")
	p := "dummy passphrase "
	return CreateTargetInput{
		Request: IPTablesRequest{
			Automated:           true,
			ConcurrentTransfers: 6,
			DNSIPPrimary:        "9.9.9.9",
			DNSDomainOrigin:     "ex.com",
			DownloadDirectory:   "/opt",
			Hostname:            "h",
			Label:               "lab",
			NetworkType:         "IPTABLE",
			SSHPort:             22,
			ICPort:              8443,
			TerminalPortStart:   3270,
			IPTablesSetting: IPTablesSetting{
				ZosIPAddress:    "172.26.1.2",
				ZosSSHRoutePort: 2022,
				TCPForwardPorts: []ForwardPortRange{{StartPort: 0, EndPort: 21}},
				UDPForwardPorts: []ForwardPortRange{{StartPort: 111, EndPort: 111}},
				TCPReroutePorts: []ReroutePortMapping{{LinuxPort: 2022, ZosPort: 22}},
				UDPReroutePorts: []ReroutePortMapping{{LinuxPort: 2022, ZosPort: 22}},
			},
		},
		Auth: SSHCredentials{
			Username:   "u",
			Password:   "",
			KeyFile:    &k,
			Passphrase: &p,
		},
	}
}

func TestCreate_GreenPath(t *testing.T) {
	check := func(path string, input any) {
		var gotBody bytes.Buffer
		var hdrOK bool
		srv := httptest.NewServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != path || r.Method != http.MethodPost {
						t.Errorf("got: path %s, want: %s", r.URL.Path, path)
					}
					h := r.Header.Get(headerSSHAuth)
					hdrOK = strings.Contains(h, `"ssh-token":"tok"`)
					_, err := io.Copy(&gotBody, r.Body)
					if err != nil {
						t.Fatalf("failure to copy body")
					}
					w.Header().Set("Content-Type", "application/json")
					_, err = w.Write([]byte(`{"uuid":"999"}`))
					if err != nil {
						t.Fatalf("failure to write uuid")
					}
				},
			),
		)
		t.Cleanup(srv.Close)
		svc := &Service{APIClient: &stubClient{srv.URL, srv.Client()}, ssh: &fakeSSH{}, mon: &fakeMon{}}
		var err error
		switch v := input.(type) {
		case CreateTargetInput:
			_, err = svc.CreateIPTables(t.Context(), v)
		case MacvtapCreateInput:
			_, err = svc.CreateMacVtap(t.Context(), v)
		}
		if err != nil {
			t.Errorf("create failed: %v", err)
		}
		if !hdrOK {
			t.Errorf("missing ssh-token header")
		}
		if bytes.Contains(gotBody.Bytes(), []byte(`"Username"`)) {
			t.Errorf("control fields in JSON")
		}
	}
	check(iptablesPath, goodIPT())
	check(vtapPath, goodMAC())
	check(iptablesPath, goodIPTwithKey())
}

func TestCreate_Errors(t *testing.T) {
	svc := &Service{}
	bad := goodIPT()
	bad.Request.SSHPort = 70000
	if _, err := svc.CreateIPTables(t.Context(), bad); err == nil {
		t.Errorf("expected port validation error")
	}

	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	svc = &Service{APIClient: &stubClient{srv.URL, srv.Client()}, ssh: &fakeSSH{failHost: "bad"}, mon: &fakeMon{}}
	bad2 := goodIPT()
	bad2.Request.Hostname = "bad"
	if _, err := svc.CreateIPTables(t.Context(), bad2); err == nil {
		t.Errorf("expected ssh auth error")
	}

	srv400 := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, `{"code":1}`, http.StatusBadRequest)
			},
		),
	)
	t.Cleanup(srv400.Close)
	svc = &Service{APIClient: &stubClient{srv400.URL, srv400.Client()}, ssh: &fakeSSH{}, mon: &fakeMon{}}
	if _, err := svc.CreateIPTables(t.Context(), goodIPT()); err == nil {
		t.Errorf("expected server 400 error")
	}

	srvOK := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{"uuid":"x"}`))
				if err != nil {
					t.Fatalf("failure to write uuid")
				}
			},
		),
	)
	t.Cleanup(srvOK.Close)
	svc = &Service{APIClient: &stubClient{srvOK.URL, srvOK.Client()}, ssh: &fakeSSH{}, mon: &fakeMon{fail: true}}
	if _, err := svc.CreateIPTables(t.Context(), goodIPT()); err == nil {
		t.Errorf("expected monitor error")
	}
}

func TestGetAndDeletePaths(t *testing.T) {
	srv := httptest.NewServer(http.NewServeMux())
	t.Cleanup(srv.Close)
	svc := &Service{APIClient: &stubClient{srv.URL, srv.Client()}, ssh: &fakeSSH{}, mon: &fakeMon{}}
	if _, err := svc.Get(t.Context(), "none"); err == nil {
		t.Errorf("expected 404 error")
	}

	mux, ok := srv.Config.Handler.(*http.ServeMux)
	if !ok {
		t.Fatalf("failure type checking http.ServeMux")
	}
	mux.HandleFunc(
		targetPath+"u1", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"uuid":"u1","status":"OK","online":true}`))
			if err != nil {
				t.Fatalf("failure to write uuid")
			}
		},
	)
	mux.HandleFunc(
		linuxPath+"u1", func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if q.Get("force") != "true" || q.Get("resume") != "false" {
				t.Errorf("bad query %v", q)
			}
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"uuid":"u1"}`))
			if err != nil {
				t.Fatalf("failure to write uuid")
			}
		},
	)

	if got, err := svc.Get(t.Context(), "u1"); err != nil || got.UUID != "u1" {
		t.Errorf("get failed: %v, got UUID %v", err, got.UUID)
	}
	if err := svc.Delete(t.Context(), "u1", true, false); err != nil {
		t.Errorf("delete failed: %v", err)
	}
}
