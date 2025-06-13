# Copyright (c) IBM Corporation
# SPDX-License-Identifier: Apache-2.0



variable "ode_host" {
  type        = string
  description = "ODE API host"
}

variable "ode_username" {
  type        = string
  description = "API username"
}

variable "ode_password" {
  type        = string
  description = "API password"
  sensitive   = true
}

variable "ssh_target_user" {
  type        = string
  description = "Linux SSH username"
}

variable "ssh_target_password" {
  type        = string
  description = "Linux SSH password"
  sensitive   = true
}

variable "label" {
  type        = string
  description = "Label for the target"
  default     = "terraformz-dev-target"
}

variable "description" {
  type        = string
  description = "Optional description"
  default     = "x86 Terraform Z Dev Target"
}

variable "hostname" {
  type        = string
  description = "Linux hostname or IP"
}

variable "ssh_port" {
  type        = number
  description = "SSH port"
  default     = 22
}

variable "ic_port" {
  type        = number
  description = "Instance controller port"
  default     = 8443
}

variable "install_dir" {
  type        = string
  description = "ODE installation directory"
  default     = "/opt"
}

variable "concurrent_transfers" {
  type        = number
  description = "Parallel transfer count"
  default     = 10
}

variable "resume" {
  type        = bool
  description = "Whether to resume a failed action"
  default     = false
}

variable "dns_ip_primary" {
  type        = string
  description = "Primary DNS IP"
}

variable "dns_domain_origin" {
  type        = string
  description = "DNS search domain"
}

variable "terminal_port_start" {
  type        = number
  description = "Start port for 3270 terminal"
  default     = 3270
}

variable "network_type" {
  type        = string
  description = "Always 'IPTABLE'"
  default     = "IPTABLE"
}