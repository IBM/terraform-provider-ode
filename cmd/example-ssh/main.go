// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/config"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/logger"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/client"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/ssh"
)

func main() {
	base := flag.String("base", os.Getenv("BASE_URL"), "ODT base URL")
	apiUser := flag.String("apiUser", os.Getenv("API_USERNAME"), "API username")
	apiPass := flag.String("apiPass", os.Getenv("API_PASSWORD"), "API password")
	host := flag.String("host", "", "SSH target hostname")
	uuid := flag.String("system-uuid", "", "System UUID")
	sshUser := flag.String("ssh-user", "", "SSH username")
	sshPass := flag.String("ssh-pass", "", "SSH password")
	keyPath := flag.String("key", "", "Path to SSH private key")
	flag.Parse()

	if *base == "" || *apiUser == "" || *apiPass == "" || *host == "" || *uuid == "" || *keyPath == "" {
		log.Fatal("all flags -base -api-user -api-pass -host -uuid -ssh-user -ssh-pass -key are required")
	}
	json := logger.New(
		os.Stdout,
		logger.WithFormatter(logger.JSON("  ")),
		logger.WithColor(true),
	)
	cfg := config.Config{
		BaseURL:   *base,
		User:      *apiUser,
		Pass:      *apiPass,
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpc := &http.Client{Timeout: 30 * time.Second}
	cli, err := client.New(
		&cfg,
		client.WithHTTPClient(httpc),
		client.WithLogger(json),
	)
	if err != nil {
		log.Fatalf("client build: %v", err)
	}

	svc := ssh.New(cli)
	ctx := context.Background()

	t1, err := svc.AuthenticateTarget(ctx, *host, *sshUser, *sshPass, 22)
	if err != nil {
		log.Fatalf("AuthenticateTarget: %v", err)
	}
	fmt.Println("Target token:", t1.Value)

	t2, err := svc.AuthenticateInstance(ctx, *uuid, *sshUser, *sshPass)
	if err != nil {
		log.Fatalf("AuthenticateInstance: %v", err)
	}
	fmt.Println("Instance token:", t2.Value)

	rawKey, err := os.ReadFile(*keyPath)
	if err != nil {
		log.Fatalf("read key: %v", err)
	}
	t3, err := svc.AuthenticateKey(ctx, *uuid, *sshUser, *sshPass, rawKey)
	if err != nil {
		log.Fatalf("AuthenticateKey: %v", err)
	}
	fmt.Println("Key token:", t3.Value)
}
