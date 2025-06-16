# Copyright (c) IBM Corporation
# SPDX-License-Identifier: Apache-2.0



output "all_image_count" {
  description = "Total number of images"
  value       = length(data.ode_image.all.image_list)
}

output "all_first_image" {
  description = "UUID of the first image in the list"
  value       = data.ode_image.all.image_list[0]
}

output "all_first_image_uuid" {
  description = "UUID of the first image in the list"
  value       = data.ode_image.all.image_list[0].uuid
}

output "all_second_image" {
  description = "UUID of the first image in the list"
  value       = data.ode_image.all.image_list[1]
}

output "all_second_image_uuid" {
  description = "UUID of the first image in the list"
  value       = data.ode_image.all.image_list[1].uuid
}

output "by_uuid_label" {
  description = "Label for the image fetched by UUID"
  value       = data.ode_image.by_uuid.image_list[0].label
}

output "by_uuid_version" {
  description = "Version for the image fetched by UUID"
  value       = data.ode_image.by_uuid.image_list[0].version
}

output "by_uuid_flat_sysres" {
  description = "SYSRES component UUID for the flattened UUID lookup"
  value       = data.ode_image.by_uuid_flat.sysres_component_uuid
}

output "by_uuid_flat_load_suffix" {
  description = "LoadSuffix for the flattened UUID lookup"
  value       = data.ode_image.by_uuid_flat.ipl_parameter.load_suffix
}

output "by_label_total_versions" {
  description = "Number of versions for the given label"
  value       = length(data.ode_image.by_label.image_list)
}

output "by_label_versions" {
  description = "All version numbers for the given label"
  value       = data.ode_image.by_label.image_list[*].version
}

output "by_label_flat_uuid" {
  description = "UUID for the flattened label-only lookup"
  value       = data.ode_image.by_label_flat.uuid
}

output "by_label_flat_version" {
  description = "Version for the flattened label-only lookup"
  value       = data.ode_image.by_label_flat.version
}

output "by_label_version_sysres" {
  description = "SYSRES component UUID for the label+version lookup"
  value       = data.ode_image.by_label_version.image_list[0].sysres_component_uuid
}

output "by_label_version_load_suffix" {
  description = "LoadSuffix for the label+version lookup"
  value       = data.ode_image.by_label_version.image_list[0].ipl_parameter.load_suffix
}

output "pinned_image_details" {
  description = "All key fields for the flattened label+version lookup"
  value = {
    uuid                  = data.ode_image.by_label_version_flat.uuid
    label                 = data.ode_image.by_label_version_flat.label
    version               = data.ode_image.by_label_version_flat.version
    sysres_component_uuid = data.ode_image.by_label_version_flat.sysres_component_uuid
    load_suffix           = data.ode_image.by_label_version_flat.ipl_parameter.load_suffix
  }
}

