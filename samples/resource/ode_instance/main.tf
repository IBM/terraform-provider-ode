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

resource "ode_instance" "this" {
  ssh_target_user     = var.ssh_target_user
  ssh_target_password = var.ssh_target_password

  general = {
    label                 = var.label
    description           = var.description
    target_uuid           = var.target_uuid
    image_uuid            = var.image_uuid
    ssh_public_key        = var.ssh_public_key
    deployment_directory  = var.deployment_directory
    sysres_component_uuid = var.sysres_component_uuid
  }

  emulator = {
    cp  = var.cp
    ram = var.ram
  }

}
