---
page_title: "Provisioning a z/OS Instance with IBM ODE Provider and Changing Password with Ansible"
subcategory: ""
description: |-
  The following demonstrates how Ansible and Terraform can be used together to achieve Infrastructure as Code and Configuration as Code.
---

# Provisioning a z/OS Instance with IBM ODE Provider and Changing Password with Ansible
Terraform and Ansible are powerful tools that, when combined, enable a seamless workflow for provisioning and configuring infrastructure. Let’s break down how they work together in this practical use case.

## Terraform: Infrastructure as Code (IaC)
Terraform is used to provision infrastructure. It defines resources like virtual machines, networks, and storage using declarative configuration files. In the tutorial:

- Terraform uses the IBM ODE provider to provision a z/OS instance.
- It defines the instance’s properties (CPU, RAM, image, SSH keys, etc.) in a `main.tf` file.
- It also dynamically creates an Ansible inventory using the `ansible_host` resource.

## Ansible: Configuration as Code (CaC)
Ansible is used to configure the provisioned infrastructure. It connects to the instance and executes tasks like installing software, changing settings, or managing users. In the guide:

- An Ansible playbook (change_ibmuser_password.yml) is used to change the ibmuser password on the z/OS instance.
- The playbook is triggered automatically after provisioning using Terraform’s `local-exec` provisioner.

# Workflow Integration: Step-by-Step
Here’s how Terraform and Ansible work together in this use case:

1. Provision with Terraform
Terraform provisions the z/OS instance using the IBM ODE provider:
   ```

   resource "ode_instance" "demo-zos-env" {
     # Configuration for z/OS instance
   }
   ```

2. Generate Ansible Inventory
Terraform dynamically creates an Ansible inventory entry:
   ```

   resource "ansible_host" "zos-vm" {
     name = ode_instance.demo-zos-env.hostname
     variables = {
       ansible_host = ode_instance.demo-zos-env.hostname,
       ansible_user = "ibmuser",
       ansible_port = 2022
     }
   }
   ```

   Use Ansible inventory plugin as:
   inventory.yml
   ```
   ---
   plugin: cloud.terraform.terraform_provider

   ```

   This ensures Ansible knows how to connect to the new instance.

1. Trigger Ansible Automatically
Terraform uses a null_resource with a local-exec provisioner to run the Ansible playbook:
   ```

   resource "null_resource" "run_ansible" {
     depends_on          = [ode_instance.hcf-zdt1-zos-env]
     provisioner "local-exec" {
       command = "ansible-playbook -i inventory.yml change_ibmuser_password.yml"
     }
   }
   ```

   This bridges the gap between provisioning and configuration.

1. Configure with Ansible
The Ansible playbook connects to the instance and changes the password using Ansible Core for z/OS `zos_tso_command` module:
   ```

   - name: Change ibmuser password
     hosts: all
     vars:
      PYZ: "/path/to/python"
      ZOAU: "/path/to/zoau"
      ansible_python_interpreter: "{{ PYZ }}/bin/python3"
      environment_vars:
         _BPXK_AUTOCVT: "ON"
         ZOAU_HOME: "{{ ZOAU }}"
         LIBPATH: "{{ ZOAU }}/lib:{{ PYZ }}/lib:/lib:/usr/lib:."
         PATH: "{{ ZOAU }}/bin:{{ PYZ }}/bin:/bin:/var/bin"
         _CEE_RUNOPTS: "FILETAG(AUTOCVT,AUTOTAG) POSIX(ON)"
         _TAG_REDIR_ERR: "txt"
         _TAG_REDIR_IN: "txt"
         _TAG_REDIR_OUT: "txt"
         LANG: "C"

     tasks:
       - name: change IBMUSER pwd
         ibm.ibm_zos_core.zos_tso_command:
           commands:
             - ALTUSER IBMUSER PHRASE('new_secure_password') NOEXPIRE RESUME
   ````


# Benefits of This Integration
- Automation: No manual steps between provisioning and configuration.
- Consistency: Infrastructure and configuration are version-controlled and repeatable.
- Scalability: Easily scale to multiple instances or environments.