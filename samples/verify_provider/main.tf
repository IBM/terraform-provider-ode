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

data "ode_image" "zos_image" {
  flatten = true
  filter = {
    label   = var.image_label
    version = var.image_version
  }
}

resource "ode_target" "install_target" {
  label                = var.target_label
  description          = var.target_description
  hostname             = var.hostname
  ssh_port             = var.ssh_port
  ic_port              = var.ic_port
  concurrent_transfers = var.concurrent_transfers
  resume               = var.resume
  install_dir          = var.install_dir
  dns_ip_primary       = var.dns_ip_primary
  dns_domain_origin    = var.dns_domain_origin
  terminal_port_start  = var.terminal_port_start
  network_type         = var.network_type

  ssh_target_user     = var.ssh_target_user
  ssh_target_password = var.ssh_target_password

  iptable_setting = {
    zos_ip_address     = "172.26.1.2"
    zos_ssh_route_port = 2022

    tcp_forward_ports = [
      { start_port = 0, end_port = 21 },
      { start_port = 23, end_port = 2021 },
      { start_port = 2023, end_port = 3269 },
      { start_port = 3271, end_port = 8442 },
      { start_port = 8444, end_port = 9449 },
      { start_port = 9452, end_port = 65535 }
    ]

    udp_forward_ports = [
      { start_port = 111, end_port = 111 },
      { start_port = 514, end_port = 514 },
      { start_port = 1023, end_port = 1023 },
      { start_port = 1044, end_port = 1049 },
      { start_port = 2049, end_port = 2049 }
    ]

    tcp_reroute_ports = [
      { linux_port = 2022, zos_port = 22 }
    ]

    udp_reroute_ports = [
      { linux_port = 2022, zos_port = 22 }
    ]
  }
}


resource "ode_instance" "zos_25" {
  depends_on          = [ode_ipl_target.install_target]
  ssh_target_user     = var.ssh_target_admin_user
  ssh_target_password = var.ssh_target_admin_password

  general = {
    label                 = var.instance_label
    description           = var.instance_description
    target_uuid           = ode_ipl_target.install_target.id
    image_uuid            = data.ode_image.zos_image.uuid
    ssh_public_key        = var.ssh_public_key
    deployment_directory  = var.deployment_directory
    sysres_component_uuid = data.ode_image.zos_image.sysres_component_uuid
  }

  emulator = {
    cp  = var.cp
    ram = var.ram
  }
}