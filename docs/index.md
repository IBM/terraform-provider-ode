---
layout: ""
page_title: "Provider: On-demand Environments (ode) provider"
description: |-
  The On-Demand Environments (ODE) provider allows managing of z/OS instances and Linux targets as Terraform resources.

---

# On-demand Environments (ODE) provider

The On-Demand Environments (ODE) provider allows managing of z/OS instances and Linux targets as Terraform resources.


## Examples 

The following examples demonstrate the Terraform code to use the ODE provider in different scenarios.

### Configuring the provider credentials

```terraform
terraform {
  required_version = ">=1.12.0"
  required_providers {
    ode = {
      source  = "IBM/ode"
      version = "1.0.0"
    }
  }
}

provider "ode" {
  ode_host     = "https://your-ode-hostname:port"
  ode_username = "your-ode-user"
  ode_password = "your-ode-password"
  ode_tls = {
    ca_file     = file("/path/to/ca_file")
    server_name = "your-ode-server-name-matching-ca-certificate"
  }

}
```

### Example ODE provider with data source configuration

The following example shows an ODE data source struct that you can use.

```terraform
terraform {
  required_version = ">=1.12.0"
  required_providers {
    ode = {
      source  = "IBM/ode"
      version = "1.0.0"
    }
  }
}

provider "ode" {
  ode_host     = "https://your-ode-hostname:port"
  ode_username = "your-ode-user"
  ode_password = "your-ode-password"
}

###########################################################
#  Use one of the filters below for a data source
###########################################################

data "ode_image" "all" {}

data "ode_image" "by_uuid" {
  uuid = "11111111-aaaa-bbbb-cccc-222222222222"
}

data "ode_image" "by_uuid_flat" {
  uuid    = "11111111-aaaa-bbbb-cccc-222222222222"
  flatten = true
}

data "ode_image" "by_label" {
  filter = {
    label = "some-label"
  }
}

data "ode_image" "by_label_flat" {
  filter = { label = "some-label" }
}

data "ode_image" "by_label_version" {
  filter = {
    label   = "some-label"
    version = some-integer # integer
  }
}

data "ode_image" "by_label_version_flat" {
  flatten = true
  filter = {
    label   = "some-label"
    version = some-integer # integer
  }
}
```

### Example ODE provider with target configuration

The following example shows an ODE target source struct that you can use.

```terraform
terraform {
  required_version = ">=1.12.0"
  required_providers {
    ode = {
      source  = "IBM/ode"
      version = "1.0.0"
    }
  }
}


provider "ode" {
  ode_host     = "https://your-ode-hostname:port"
  ode_username = "your-ode-user"
  ode_password = "your-ode-password"
}

resource "ode_target" "from_json_match" {
  label                = "your_label"
  description          = "your_description"
  hostname             = "your_target_hostname"
  ssh_port             = 22
  ic_port              = 8443
  concurrent_transfers = 3
  install_dir          = "/opt"
  dns_ip_primary       = "your_dsn_ip_primary"
  terminal_port_start  = 3270
  resume               = "your_resume"
  dns_domain_origin    = "your_dns_domain_origin"
  network_type         = "your_network_type"
  ssh_target_user      = "your_ssh_target_user"
  ssh_target_password  = "your_ssh_target_password"

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
```

### Example of an ODE instance

The following example shows an ODE instance with data source and target resource that you can use.

```terraform
terraform {
  required_version = ">=1.12.0"
  required_providers {
    ode = {
      source  = "IBM/ode"
      version = "1.0.0"
    }
  }
}

provider "ode" {
  ode_host     = "https://your-ode-hostname:port"
  ode_username = "your-ode-user"
  ode_password = "your-ode-password"
}

data "ode_image" "zos_image" {
  flatten = true
  filter = {
    label   = "some-image-label"
    version = some-integer
  }
}

resource "ode_target" "install_target" {
  label                = "your_label"
  description          = "your_description"
  hostname             = "your_target_hostname"
  ssh_port             = 22
  ic_port              = 8443
  concurrent_transfers = 3
  install_dir          = "/opt"
  terminal_port_start  = 3270
  resume               = "your_resume"
  dns_ip_primary       = "your_dns_ip_primary"
  dns_domain_origin    = "your_dns_domain_origin"
  network_type         = "your_network_type"
  ssh_target_user      = "your_ssh_target_user"
  ssh_target_password  = "your_ssh_target_password"

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
  ssh_target_user     = "your_ssh_target_admin_user"
  ssh_target_password = "your_ssh_target_admin_password"

  general = {
    label                 = "your_instance_label"
    description           = "your_instance_description"
    target_uuid           = ode_ipl_target.install_target.id
    image_uuid            = data.ode_image.zos_image.uuid
    ssh_public_key        = "your_ssh_public_key"
    deployment_directory  = "/opt"
    sysres_component_uuid = data.ode_image.zos_image.sysres_component_uuid
  }

  emulator = {
    cp  = 4
    ram = 5368709120
  }
}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `ode_host` (String) Host for On-Demand Environments. This can also be sourced from the `ODE_HOST` environment variable.
- `ode_password` (String, Sensitive) Password for On-Demand Environments authentication. This can also be sourced from the `ODE_PASSWORD` environment variable.
- `ode_tls` (Attributes) Certificate for On-Demand Environments. (see [below for nested schema](#nestedatt--ode_tls))
- `ode_username` (String) Username for On-Demand Environments authentication. This can also be sourced from the `ODE_USERNAME` environment variable.

<a id="nestedatt--ode_tls"></a>
### Nested Schema for `ode_tls`

Optional:

- `ca_file` (String) CA file for On-Demand Environments. This can also be sourced from the ODE_TLS_CA_FILE environment variable.
- `insecure_skip_verify` (Boolean) Insecure SSL certificate for On-Demand Environments. This can also be sourced from the ODE_TLS_INSECURE_SKIP_VERIFY environment variable.
- `server_name` (String) Server name for On-Demand Environments. This can also be sourced from the ODE_TLS_SERVER_NAME environment variable.