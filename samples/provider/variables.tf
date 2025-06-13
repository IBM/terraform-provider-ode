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
  type        = string
  sensitive   = true
}