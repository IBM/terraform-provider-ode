# Copyright (c) IBM Corporation
# SPDX-License-Identifier: Apache-2.0



variable "ode_host" {
  description = "ODE API server URL."
  type        = string
}

variable "ode_username" {
  description = "ODE API username."
  type        = string
}

variable "ode_password" {
  description = "ODE API password."
  type        = string
  sensitive   = true
}

variable "ssh_target_user" {
  description = "Linux SSH username."
  type        = string
  sensitive   = true
}

variable "ssh_target_password" {
  description = "Linux SSH password (if using password auth)."
  type        = string
  sensitive   = true
}

variable "ssh_target_key_file" {
  description = "Path to private key (if using key-based auth)."
  type        = string
  sensitive   = true
}

# General block
variable "label" {
  description = "Label for the instance."
  type        = string
  default     = "terraformz-test-zos-instance"
}

variable "description" {
  description = "Optional instance description."
  type        = string
  default     = "Terraform-z/OS-Test-Instance"
}

variable "target_uuid" {
  description = "UUID of the target environment."
  type        = string
}

variable "image_uuid" {
  description = "UUID of the image to provision."
  type        = string
}

variable "ssh_public_key" {
  description = "Public SSH key for the Linux user (optional)."
  type        = string
}

variable "deployment_directory" {
  description = "Path on the target where ODE deploys the instance."
  type        = string
  default     = "/opt"
}

variable "sysres_component_uuid" {
  description = "UUID of the SYSRES component."
  type        = string
}

# Emulator block
variable "cp" {
  description = "Number of CP engines."
  type        = number
  default     = 3
}

variable "ziip" {
  description = "Number of zIIP engines."
  type        = number
  default     = 0
}

variable "ram" {
  description = "RAM in bytes (5 GiB = 5368709120)."
  type        = number
  default     = 8589934592
}

# z/OS Credentials block
variable "zos_username" {
  description = "z/OS user ID."
  type        = string
  default     = "ibmuser"
}

variable "zos_password" {
  description = "z/OS password."
  type        = string
  sensitive   = true
}