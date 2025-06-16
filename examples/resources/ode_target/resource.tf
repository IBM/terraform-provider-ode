

terraform {
  required_version = ">=1.11.0"
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
