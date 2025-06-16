// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"net/http"
	"os"
	"time"

	logger "github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/logger"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/image"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/instance"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/target"
)

func main() {

	type Client struct {
		base     string
		http     *http.Client
		logger   logger.Logger
		Target   *target.Service
		Image    *image.Service
		Instance *instance.Service
	}
	c := &Client{
		logger: logger.New(
			os.Stdout,
			logger.WithFormatter(logger.Inline()),
			logger.WithColor(true),
		),
	}

	inline := logger.New(
		os.Stdout,
		logger.WithFormatter(logger.Inline()),
		logger.WithColor(true),
	)

	json := logger.New(
		os.Stdout,
		logger.WithFormatter(logger.JSON("  ")),
		logger.WithColor(true),
	)

	c.logger.Info("app_start", "env", "production")
	inline.Debug("load_config", "file", "/etc/app.yaml")
	inline.Warn("cache_miss", "key", "user_42")
	inline.Error("file_error", "path", "data.db", "retry", false)
	inline.Trace("init_step", "step", 1)

	time.Sleep(50 * time.Millisecond)
	inline.Info("api_call", "method", "GET", "path", "/users", "status", 200)

	json.Info("app_start", "env", "staging")
	json.Debug("load_plugins", "count", 3)
	json.Warn("deprecated_flag", "flag", "old_mode")
	json.Error("db_conn", "service", "users", "timeout_ms", 5000)
	json.Trace("health_check", "uptime_s", time.Now().Unix())

	time.Sleep(30 * time.Millisecond)
	json.Error("api_call", "method", "POST", "path", "/orders", "status", 500, "error", "timeout")
}
