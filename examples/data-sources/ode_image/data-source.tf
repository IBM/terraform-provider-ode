

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