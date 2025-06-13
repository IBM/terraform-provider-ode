# Copyright (c) IBM Corporation
# SPDX-License-Identifier: Apache-2.0



variable "ode_host" {
  type        = string
  description = "The ODE API host (e.g., https://example.com)"
}

variable "ode_username" {
  type        = string
  description = "Username for ODE authentication"
}

variable "ode_password" {
  type      = string
  sensitive = true
}

variable "image_uuid" {
  type        = string
  description = "UUID of the image to look up"
}

variable "image_label" {
  type        = string
  description = "Image label (name) to filter on"
}

variable "image_version" {
  type        = number
  description = "Image version to pin"
}
