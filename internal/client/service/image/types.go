// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package image

type listResponse []Group

type IPL struct {
	SysResDevice string `json:"sysres-device"`
	IODFDevice   string `json:"iodf-device"`
	LoadSuffix   string `json:"load-suffix"`
}

type StockImage struct {
	UUID                string   `json:"uuid"`
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	Version             int      `json:"version"`
	Type                string   `json:"type"`
	Size                int64    `json:"size"`
	IPLParameter        *IPL     `json:"ipl-parameter"`
	SysResComponentUUID string   `json:"sysres-component-uuid"`
	ProvisionUUIDs      []string `json:"provision-uuids"`
}

type ListInput struct {
	SourceEnvironmentUUIDs []string
	Versions               bool
	UUID                   string
	Label                  string
	Version                *int
}

type Group struct {
	Name        string       `json:"name"`
	StockImages []StockImage `json:"image-list"`
}
