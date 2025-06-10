

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