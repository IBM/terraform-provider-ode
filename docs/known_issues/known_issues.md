---
page_title: "Known Issues"
subcategory: ""
description: |-
  Known issues with the current version of the ODE provider:
---
  The following issues are known with the current version of the ODE provider:
  - If the `timeouts` attribute of the `ode_instance` resource is specified and the timeout occurs, the ODE provider state file might not always be updated with the final status of the instance. In this case, you must manually remove the ODE state file and delete the instance directly with the ODE UI. Alternately, you can try to import the instance to sync up the ODE state file, then run Terraform Apply or Destroy again if needed.
  - The `ipl` block and its attributes of `device_address, iodf_address,load_suffix` are not being used if specified.  ODE will calculate these values based on the z/OS images.
  - During Terraform Apply or Destroy, if Ctrl + C is entered, the ODE provider is not able to send a Cancel request to the ODE backend server. The ODE state file therefore is out of sync with the actual status of the instance. You can manually remove the ODE state file and delete the instance directly with ODE UI. Alternately, you can try to import the instance to sync up the ODE state file, then run Terraform Apply or Destroy again. 