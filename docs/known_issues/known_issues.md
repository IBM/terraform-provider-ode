---
page_title: "Known Issues"
subcategory: ""
description: |-
  Known issues with the current version of the On-Demand Environments provider:
---

# Known Issues

  The following issues are known with the current version of the On-Demand Environments provider:
  - If the `timeouts` attribute of the `ode_instance` resource is specified and the timeout occurs, the On-Demand Environments provider state file might not always be updated with the final status of the instance. In this case, you must manually remove the state file and delete the instance directly with the On-Demand Environments UI. Alternately, you can try to import the instance to sync up the state file, then run the  `terraform apply` or `terraform destroy` command again if needed.
  - The `ipl` block and its attributes of `device_address, iodf_address,load_suffix` are not being used if specified. On-Demand Environments will calculate these values based on the z/OS images.
  - During `terraform apply` or `terraform destroy`, if Ctrl + C is entered, the On-Demand Environments provider is not able to send a cancel request to the On-Demand Environments backend server. The state file therefore is out of sync with the actual status of the instance. You can manually remove the state file and delete the instance directly with On-Demand Environments UI. Alternately, you can try to import the instance to sync up the state file, then run `terraform apply` or `terraform destroy` command again.
  - On the `ode_target` resource, `iptable_setting` is required. Providing empty values for `tcp_forward_ports`, `tcp_reroute_ports`, `udp_forward_ports`, or `udp_reroute_ports` could lead to a terraform drift.
  