# Copyright (c) IBM Corporation
# SPDX-License-Identifier: Apache-2.0



terraform {
  required_version=">=1.11.0"
  required_providers {
    ode = {
      source  = "IBM/ode"
      version = "1.0.0"
    }
  }
}

provider "ode" {
  ode_host     = var.ode_host
  ode_username = var.ode_username
  ode_password = var.ode_password
  ode_tls = {
    ca_file = "/path/to/ca_file"
  }
}