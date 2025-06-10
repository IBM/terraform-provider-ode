// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package instance

type Request struct {
	General             General  `json:"general"`
	Emulator            Emulator `json:"emulator"`
	ZosCreds            ZosCreds `json:"zos-creds,omitempty"`
	IPL                 *IPL     `json:"ipl,omitempty"`
	DeploymentDirectory string   `json:"deployment-directory,omitempty"`
	ValidateLinux       bool     `json:"validate-linux,omitempty"`
}

type SSHCredentials struct {
	Username   string  `json:"username,omitempty"`
	Password   string  `json:"password,omitempty"`
	KeyFile    *[]byte `json:"key_file,omitempty"`
	Passphrase *string `json:"passphrase,omitempty"`
}

type SSHAuth struct {
	Token      string `json:"ssh-token"`
	SystemUUID string `json:"system-uuid"`
}
type Emulator struct {
	CP   int64 `json:"cp"`
	Ziip int64 `json:"ziip"`
	Ram  int64 `json:"ram"`
}

type IPL struct {
	DeviceAddress string `json:"device-address,omitempty"`
	IODFAddress   string `json:"iodf-address,omitempty"`
	LoadSuffix    string `json:"load-suffix,omitempty"`
}

type General struct {
	Label               string `json:"label"`
	Description         string `json:"description,omitempty"`
	SSHPublicKey        string `json:"ssh-public-key,omitempty"`
	TargetUUID          string `json:"target-uuid"`
	ImageUUID           string `json:"image-uuid"`
	SysResComponentUUID string `json:"sysres-component-uuid,omitempty"`
}

type ZosCreds struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type LinuxProvisionRequest struct {
	General             General  `json:"general"`
	Emulator            Emulator `json:"emulator"`
	IPL                 *IPL     `json:"ipl,omitempty"`
	ZosCreds            ZosCreds `json:"zos-creds,omitempty"`
	DeploymentDirectory string   `json:"deployment-directory,omitempty"`
	ValidateLinux       bool     `json:"validate-linux,omitempty"`
}

type CreateInput struct {
	Request LinuxProvisionRequest `json:"-"`
	Auth    SSHCredentials        `json:"-"`
}

type Data struct {
	Label               string   `json:"label,omitempty"`
	Description         string   `json:"description,omitempty"`
	ImageUUID           string   `json:"image-uuid,omitempty"`
	TargetUUID          string   `json:"target-uuid,omitempty"`
	SysResComponentUUID string   `json:"sysres-component-uuid,omitempty"`
	DeploymentDirectory string   `json:"deployment-directory,omitempty"`
	SSHTargetUser       string   `json:"linux-username,omitempty"`
	ProvisionUUID       string   `json:"uuid"`
	SSHPublicKey        string   `json:"zos-ssh-public-key-value"`
	Successful          bool     `json:"successful,omitempty"`
	Failed              bool     `json:"failed,omitempty"`
	Cancelled           bool     `json:"cancelled,omitempty"`
	InProgress          bool     `json:"in-progress,omitempty"`
	User                *User    `json:"user,omitempty"`
	Emulator            Emulator `json:"emulator"`
	IPL                 *IPL     `json:"ipl,omitempty"`
}

type DeleteInput struct {
	ProvisionUUID string `json:"provision_uuid"`
	Force         bool   `json:"force"`
	Resume        bool   `json:"resume"`
}

type DeleteResp struct {
	UUID string `json:"uuid"`
}

type User struct {
	SSHPublicKeyUser string `json:"ssh-public-key"`
}
