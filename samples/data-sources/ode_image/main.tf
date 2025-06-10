# Copyright (c) IBM Corporation
# SPDX-License-Identifier: Apache-2.0



terraform {
  required_version=">=1.12.0"
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

data "ode_image" "all" {}

data "ode_image" "by_uuid" {
  uuid = var.image_uuid
}

data "ode_image" "by_uuid_flat" {
  uuid    = var.image_uuid
  flatten = true
}

data "ode_image" "by_label" {
  filter = {
    label = var.image_label
  }
}

data "ode_image" "by_label_flat" {
  filter = { label = var.image_label }
}

data "ode_image" "by_label_version" {
  filter = {
    label   = var.image_label
    version = var.image_version
  }
}

data "ode_image" "by_label_version_flat" {
  flatten = true
  filter = {
    label   = var.image_label
    version = var.image_version
  }
}