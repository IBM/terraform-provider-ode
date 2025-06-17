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

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/client"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/config"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/logger"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/monitor"
)

func main() {
	base := flag.String("base", "", "API base URL")
	apiUser := flag.String("user", "", "API basic-auth user")
	apiPass := flag.String("pass", "", "API basic-auth pass")
	uuid := flag.String("uuid", "", "Target or instance UUID")
	isInstance := flag.Bool("instance", true, "set to true to monitor instance instead of target")
	interval := flag.Duration("interval", 30*time.Second, "poll interval")
	timeout := flag.Duration("timeout", 90*time.Minute, "overall timeout")
	flag.Parse()

	if *base == "" || *apiUser == "" || *apiPass == "" || *uuid == "" {
		log.Fatal("flags -base, -user, -pass and -uuid are required")
	}

	httpc := &http.Client{
		Timeout:   10 * time.Second, // per-check timeout
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	jsonz := logger.New(
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
	cli, err := client.New(
		&cfg,
		client.WithHTTPClient(httpc),

		client.WithLogger(jsonz),
	)
	if err != nil {
		log.Fatalf("client build: %v", err)
	}

	svc := monitor.New(cli, *interval, *timeout)
	ctx := context.Background()

	var monErr error
	if *isInstance {
		monErr = svc.MonitorInstance(ctx, *uuid)
	} else {
		monErr = svc.MonitorTarget(ctx, *uuid)
	}

	if monErr != nil {
		log.Fatalf("monitor failed: %v", monErr)
	}
	fmt.Println("monitor completed successfully")
}
