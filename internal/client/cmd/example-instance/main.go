// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/config"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/logger"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/instance"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/client"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/monitor"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/ssh"
)

var defaultPayload = []byte(`{
  "emulator": {
    "cp": ,
    "ram": 
  },
  "general": {
    "label": "",
    "description": "",
    "target-uuid": "",
    "resume": ,
    "update": ,
    "image-uuid": "",
    "ssh-public-key": "",
    "sysres-component-uuid": ""
  },
  "zos-creds": {},
  "deployment-directory": "/opt",
  "validate-linux": true
}`)

func main() {
	base := flag.String("base", os.Getenv("BASE_URL"), "ODT base URL")
	apiUser := flag.String("apiUser", os.Getenv("API_USERNAME"), "API username")
	apiPass := flag.String("apiPass", os.Getenv("API_PASSWORD"), "API password")
	targetUUID := flag.String("targetUUID", "", "Target UUID")
	targetUser := flag.String("targetUser", "", "API basic auth user")
	targetPass := flag.String("targetPass", "", "API basic auth pass")
	flag.Parse()

	if *base == "" || *apiUser == "" || *targetUUID == "" {
		log.Fatal("flags -base, -targetUUID  are required")
	}

	httpc := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	jsonz := logger.New(
		os.Stdout,
		logger.WithFormatter(logger.JSON("  ")),
		logger.WithColor(true),
		logger.WithMinLevel(logger.Debug),
	)

	cfg := config.Config{
		BaseURL:   *base,
		User:      *apiUser,
		Pass:      *apiPass,
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
	}

	cli, err := client.New(
		&cfg,
		client.WithHTTPClient(httpc),
		client.WithLogger(jsonz),
	)
	if err != nil {
		log.Fatalf("client build: %v", err)
	}

	var pr instance.LinuxProvisionRequest
	if err = json.Unmarshal(defaultPayload, &pr); err != nil {
		log.Fatalf("unmarshal default payload: %v", err)
	}

	sshSvc := ssh.New(cli)
	monSvc := monitor.New(cli, 30*time.Second, 90*time.Minute)
	svc := instance.New(cli, sshSvc, monSvc)

	ctx := context.Background()
	input := instance.CreateInput{
		Request: pr,
		Auth: instance.SSHCredentials{
			Username: *targetUser,
			Password: *targetPass,
		},
	}

	fmt.Println("Creating provision…")
	uuid, err := svc.Create(ctx, input)
	if err != nil {
		log.Fatalf("Create failed: %v", err)
	}
	fmt.Printf("Provisioned UUID: %s\n", uuid)

	fmt.Println("Fetching provision…")
	data, err := svc.Get(ctx, uuid)
	if err != nil {
		log.Fatalf("Get failed: %v", err)
	}
	fmt.Printf("Provision data: %+v\n", data)

	fmt.Println("Deleting provision…")
	if err = svc.Delete(
		ctx, instance.DeleteInput{ProvisionUUID: uuid, Force: false, Resume: false},
	); err != nil {
		log.Fatalf("Delete failed: %v", err)
	}
	fmt.Println("Delete completed")
}
