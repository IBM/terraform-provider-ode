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
  ode_tls = {
    ca_file     = file("/path/to/ca_file")
    server_name = "your-ode-server-name-matching-ca-certificate"
  }

}