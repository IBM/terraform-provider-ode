# Copyright (c) IBM Corporation
# SPDX-License-Identifier: Apache-2.0



variable "ode_host" {
  type        = string
  description = "ODE API host"
}

variable "ode_username" {
  type        = string
  description = "ODE API username"
}

variable "ode_password" {
  type        = string
  description = "ODE API password"
  sensitive   = true
}

####################################### Stock Image Variables #######################################

# variable "image_uuid" {
#     type        = string
#     description = "UUID of the image to look up"
# }

variable "image_label" {
  type        = string
  description = "Image label (name) to filter on"
  default     = "IBM Stock image (1.0.2)"
}

variable "image_version" {
  type        = number
  description = "z/OS stock image version"
  default     = 1

}
####################################### IPL Target Variables #######################################


variable "ssh_target_user" {
  type        = string
  description = "Linux target SSH username"
}

variable "ssh_target_password" {
  type        = string
  description = "Linux target SSH password"
  sensitive   = true
}

variable "target_label" {
  type        = string
  description = "Label for the target"
  default     = "terraformz-dev-target"
}

variable "target_description" {
  type        = string
  description = "Optional description"
  default     = "x86 Terraform Z Dev Target"
}

variable "hostname" {
  type        = string
  description = "Linux target hostname or IP"
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
}

variable "network_type" {
  type        = string
  description = "Always 'IPTABLE'"
  default     = "IPTABLE"

}
####################################### Instance Variables #######################################

# variable "ssh_target_key_file" {
#     description = "Path to private key (if using key-based auth)."
#     type        = string
#     sensitive   = true
# }

variable "ssh_target_admin_user" {
  description = "Linux SSH username."
  type        = string
}

variable "ssh_target_admin_password" {
  description = "Linux SSH password (if using password auth)."
  type        = string
  sensitive   = true
}

# General block
variable "instance_label" {
  description = "Label for the instance."
  type        = string
  default     = "terraformz-dev-z-instance"
}

variable "instance_description" {
  description = "Optional instance description."
  type        = string
  default     = "Terraform-z/OS-Dev-Instance"
}

// When running without target dep
# variable "target_uuid" {
#     description = "UUID of the target environment."
#     type        = string
#     default     = "b6261cf0-9506-49f7-a25f-30644ad97aac"
# }

variable "ssh_public_key" {
  description = "Public SSH key for the Linux user (optional)."
  type        = string
}

variable "deployment_directory" {
  description = "Path on the target where ODE deploys the instance."
  type        = string
  default     = "/opt"
}


// When running without image dep
# variable "sysres_component_uuid" {
#     description = "UUID of the SYSRES component."
#     type        = string
# }

# Emulator block
variable "cp" {
  description = "Number of CP engines."
  type        = number
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

