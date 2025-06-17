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

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/client"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/config"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/logger"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/image"
)

func main() {
	base := flag.String("base", os.Getenv("BASE_URL"), "ODT base URL")
	user := flag.String("user", os.Getenv("API_USERNAME"), "API username")
	pass := flag.String("pass", os.Getenv("API_PASSWORD"), "API password")
	list := flag.Bool("list", true, "call /images instead of /image/{uuid}")
	vers := flag.Bool("vers", true, "include all versions")
	flag.Parse()

	if *base == "" || *user == "" || *pass == "" {
		flag.Usage()
		os.Exit(2)
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
		User:      *user,
		Pass:      *pass,
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

	svc := image.New(cli)
	ctx := context.Background()

	switch {
	case *list:
		listImages(ctx, svc, *vers)
	default:
		log.Println("nothing to do – supply -uuid or -list")
	}
}

func listImages(ctx context.Context, svc *image.Service, includeVers bool) {
	imgs, err := svc.List(
		ctx, image.ListInput{
			Versions: includeVers,
		},
	)
	if err != nil {
		log.Fatalf("list: %v", err)
	}
	log.Printf("got %d images\n", len(imgs))
	for _, im := range imgs {
		log.Printf(
			"%s  v%-2d  %s  (%d bytes)\n",
			im.UUID, im.Version, im.Name, im.Size,
		)
	}
}
