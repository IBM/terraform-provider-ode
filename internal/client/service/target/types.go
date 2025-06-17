// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package target

type SSHCredentials struct {
	Username   string  `json:"username,omitempty"`
	Password   string  `json:"password,omitempty"`
	KeyFile    *[]byte `json:"key_file,omitempty"`
	Passphrase *string `json:"passphrase,omitempty"`
}

type sshAuth struct {
	Token string `json:"ssh-token"`
	Host  string `json:"hostname"`
	Port  int    `json:"port"`
}

type CreateTargetInput struct {
	Request IPTablesRequest `json:"-"`
	Auth    SSHCredentials  `json:"-"`
}

type IPTablesRequest struct {
	Label                string          `json:"label"`
	Description          string          `json:"description,omitempty"`
	Hostname             string          `json:"hostname"`
	SSHPort              int             `json:"ssh-port"`
	ICPort               int             `json:"ic-port"`
	Automated            bool            `json:"automated"`
	DownloadDirectory    string          `json:"download-directory"`
	DNSIPPrimary         string          `json:"dns-ip-primary"`
	DNSDomainOrigin      string          `json:"dns-domain-origin"`
	ConcurrentTransfers  int             `json:"concurrent-transfers"`
	Resume               bool            `json:"resume"`
	TerminalPortStart    int             `json:"terminal-port-start"`
	TerminalPortEnd      *int            `json:"terminal-port-end,omitempty"`
	PrivilegeCommandUUID *string         `json:"privilege-command-uuid,omitempty"`
	NetworkType          string          `json:"network-type,omitempty"`
	IPTablesSetting      IPTablesSetting `json:"iptable-setting"`
}

type IPTablesSetting struct {
	ZosIPAddress    string               `json:"zos-ip-address"`
	ZosSSHRoutePort int                  `json:"zos-ssh-route-port"`
	TCPForwardPorts []ForwardPortRange   `json:"tcp-forward-ports,omitempty"`
	UDPForwardPorts []ForwardPortRange   `json:"udp-forward-ports,omitempty"`
	TCPReroutePorts []ReroutePortMapping `json:"tcp-reroute-ports,omitempty"`
	UDPReroutePorts []ReroutePortMapping `json:"udp-reroute-ports,omitempty"`
}

type SuccessResponse struct {
	UUID string `json:"uuid"`
}

type Target struct {
	UUID                string               `json:"uuid"`
	Label               string               `json:"label,omitempty"`
	Description         string               `json:"description,omitempty"`
	Hostname            string               `json:"hostname"`
	Type                string               `json:"type"`
	Status              string               `json:"status"`
	Online              bool                 `json:"online"`
	Resume              bool                 `json:"resume"`
	SSHPort             int                  `json:"port"`
	ICPort              int                  `json:"ic-port"`
	DownloadDirectory   string               `json:"download-directory,omitempty"`
	DNSIPPrimary        string               `json:"dns-ip-primary,omitempty"`
	DNSDomainOrigin     string               `json:"dns-domain-origin,omitempty"`
	ConcurrentTransfers int                  `json:"concurrent-transfers"`
	TerminalPortStart   int                  `json:"terminal-port-start,omitempty"`
	TerminalPortEnd     int                  `json:"terminal-port-end,omitempty"`
	NetworkType         string               `json:"network-type,omitempty"`
	ZosIPAddress        string               `json:"zos-ip-address"`
	ZosSSHRoutePort     int                  `json:"zos-ssh-route-port"`
	CreationResumable   bool                 `json:"creation-resumable"`
	TCPForwardPorts     []ForwardPortRange   `json:"tcp-forward-ports,omitempty"`
	TCPReroutePorts     []ReroutePortMapping `json:"tcp-reroute-ports,omitempty"`
	UDPForwardPorts     []ForwardPortRange   `json:"udp-forward-ports,omitempty"`
	UDPReroutePorts     []ReroutePortMapping `json:"udp-reroute-ports,omitempty"`
}

type ForwardPortRange struct {
	StartPort int `json:"start-port"`
	EndPort   int `json:"end-port"`
}

type ReroutePortMapping struct {
	LinuxPort int `json:"linux-port"`
	ZosPort   int `json:"zos-port"`
}

type MacVtapRequest struct {
	Automated           bool           `json:"automated"`
	ConcurrentTransfers int            `json:"concurrent-transfers"`
	Description         string         `json:"description"`
	DNSIPPrimary        string         `json:"dns-ip-primary"`
	DNSDomainOrigin     string         `json:"dns-domain-origin"`
	DownloadDirectory   string         `json:"download-directory"`
	Hostname            string         `json:"hostname"`
	Label               string         `json:"label"`
	NetworkType         string         `json:"network-type"`
	Resume              bool           `json:"resume"`
	SSHPort             int            `json:"ssh-port"`
	ICPort              int            `json:"ic-port"`
	TerminalPortStart   int            `json:"terminal-port-start"`
	MacvtapSetting      MacvtapSetting `json:"macvtap-setting"`
}

type MacvtapSetting struct {
	IPStart         string   `json:"ip-start,omitempty"`
	IPEnd           string   `json:"ip-end,omitempty"`
	PhysicalAdapter string   `json:"physical-adapter,omitempty"`
	Macvtaps        []string `json:"macvtaps,omitempty"`
	DefaultRoute    string   `json:"default-route,omitempty"`
	MaxInstances    int      `json:"max-instances,omitempty"`
}

type MacvtapCreateInput struct {
	MacVtapRequest
	Username string `json:"-"`
	Password string `json:"-"`
	Port     int    `json:"-"`
}
