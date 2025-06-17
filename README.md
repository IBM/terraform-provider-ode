# On-Demand Environments provider

The On-Demand Environments provider is used to manage On-Demand z/OS Environments with IBM Test Accelerator for Z as Terraform resources.

 On-Demand Environments is a component within IBM Test Accelerator for Z. For more details regarding IBM Test Accelerator for Z or  On-Demand Environments, visit the official documentation [site](https://www.ibm.com/products/test-accelerator-z).

## Usage Example

```
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
  ode_host     = "https://your-ode-hostname:port"
  ode_username = "admin"
  ode_password = var.ode_password
}

data "ode_image" "zos_image" {
  flatten = true
  filter = {
    label   = "IBM Stock image (1.0.2)"
    version = 1
  }
}

resource "ode_ipl_target" "install_target" {
  label                = "dev-target"
  description          = "x86 Dev Target"
  hostname             = ""mytarget.domain.com"
  concurrent_transfers = 3
  install_dir          = "/opt"
  dns_ip_primary       = "9.23.33.105"
  dns_domain_origin    = "https://your-ode-hostname:port"

  ssh_target_user     = "user1"
  ssh_target_password = var.ssh_target_password

}

resource "ode_instance" "zos_25" {
  depends_on          = [ode_ipl_target.install_target]
  ssh_target_user     = "user1"
  ssh_target_password = var.ssh_target_admin_password

  general = {
    label                 = "dev-z-instance"
    description           = "zOS dev instance"
    target_uuid           = ode_ipl_target.install_target.id
    image_uuid            = data.ode_image.zos_image.uuid
    ssh_public_key        = var.ssh_public_key
    deployment_directory  = "/opt"
    sysres_component_uuid = data.ode_image.zos_image.sysres_component_uuid
  }

  emulator = {
    cp  = 5
    ram = 8589934592
  }
}
```

## Requirements

- [Go](https://golang.org/doc/install) >= v1.24
- [Terraform](https://www.terraform.io/downloads.html) 1.11 or later

## Contributing

This provider is not currently accepting contributions. However, we encourage you to open git issues for bugs, comments or feature requests.
