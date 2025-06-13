// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/config"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/client"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/logger"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/monitor"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/ssh"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/target"
)

func main() {
	base := flag.String("base", os.Getenv("BASE_URL"), "ODT base URL")
	apiUser := flag.String("apiUser", os.Getenv("API_USERNAME"), "API username")
	apiPass := flag.String("apiPass", os.Getenv("API_PASSWORD"), "API password")

	linuxHost := flag.String("linuxHost", "", "Linux hostname or IP for SSH (required)")
	sshPort := flag.Int("sshport", 22, "SSH port on Linux linuxHost")
	sshUser := flag.String("sshuser", "", "SSH username")
	sshPass := flag.String("sshpass", "", "SSH password (required)")

	label := flag.String("label", "", "target label")
	desc := flag.String("desc", "", "optional description")

	insecure := flag.Bool("insecure", true, "skip TLS verify")
	flag.Parse()

	if *linuxHost == "" || *apiUser == "" || *apiPass == "" || *sshPass == "" {
		flag.Usage()
		os.Exit(2)
	}

	httpc := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: *insecure},
		},
	}
	jsonLog := logger.New(
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
		client.WithLogger(jsonLog),
	)
	if err != nil {
		log.Fatalf("client: %v", err)
	}

	sshSvc := ssh.New(cli)
	monSvc := monitor.New(cli, 15*time.Second, 60*time.Minute)
	tgtSvc := target.New(cli, sshSvc, monSvc)

	ctx := context.Background()

	in := target.CreateTargetInput{
		Request: target.IPTablesRequest{
			Label:               *label,
			Description:         *desc,
			Resume:              false,
			ConcurrentTransfers: 6,
			Hostname:            *linuxHost,
			SSHPort:             *sshPort,
			ICPort:              8443,
			Automated:           true,
			DNSIPPrimary:        "",
			DNSDomainOrigin:     "",
			DownloadDirectory:   "/opt",
			TerminalPortStart:   0,
			NetworkType:         "IPTABLE",
			IPTablesSetting:     defaultIPTables(),
		},
		Auth: target.SSHCredentials{
			Username: *sshUser,
			Password: *sshPass,
		},
	}

	uuid, err := tgtSvc.CreateIPTables(ctx, in)
	if err != nil {
		log.Fatalf("create: %v", err)
	}
	log.Printf("created target UUID %s", uuid)

	tg, err := tgtSvc.Get(ctx, uuid)
	if err != nil {
		log.Fatalf("get: %v", err)
	}
	log.Printf("target status: %s  online=%v", tg.Status, tg.Online)

	if err := tgtSvc.Delete(ctx, uuid, true, false); err != nil {
		log.Fatalf("delete: %v", err)
	}
	log.Printf("deleted target %s", uuid)
}

func defaultIPTables() target.IPTablesSetting {
	return target.IPTablesSetting{
		ZosSSHRoutePort: 2022,
		ZosIPAddress:    "172.26.1.2",
		TCPForwardPorts: []target.ForwardPortRange{
			portRange(0, 21), portRange(23, 2021), portRange(2023, 3269),
			portRange(3271, 8442), portRange(8444, 9449), portRange(9452, 65535),
		},
		UDPForwardPorts: []target.ForwardPortRange{
			portRange(111, 111), portRange(514, 514), portRange(1023, 1023),
			portRange(1044, 1049), portRange(2049, 2049),
		},
		TCPReroutePorts: []target.ReroutePortMapping{{LinuxPort: 2022, ZosPort: 22}},
		UDPReroutePorts: []target.ReroutePortMapping{{LinuxPort: 2022, ZosPort: 22}},
	}
}

func portRange(start, end int) target.ForwardPortRange {
	return target.ForwardPortRange{StartPort: start, EndPort: end}
}
