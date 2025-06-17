---
page_title: "Managing On-Demand Environments with GitOps, Terraform, and Ansible Automation Platform"
subcategory: ""
description: |-
  The following demonstrates how Git, Terraform and Ansible Automation Platform can be used together to achieve Infrastructure as Code and Configuration as Code.
---

# Managing On-Demand Environments with GitOps, Terraform, and Ansible Automation Platform

## Introduction

Automation is the backbone of scalable, reliable, and secure infrastructure. When you work with On-Demand Environments, combining GitOps, Terraform, Ansible Automation Platform, and Event-Driven Ansible creates a powerful, event-driven and declarative infrastructure pipeline.

In this guide, you will learn how to use GitOps, Terraform, Ansible Automation Platform, and Event-Driven Ansible to provision and configure On-Demand z/OS Environments. This powerful combination of tools allows for efficient and automated management of z/OS instances, ensuring that your infrastructure is always up-to-date and configured correctly.

## Key components overview

- GitOps is a methodology that uses Git as the single source of truth for infrastructure and application configurations. It enables version-controlled, auditable, and automated deployments.
- Terraform by HashiCorp is an infrastructure as code (IaC) tool that allows you to define and provision infrastructure by using a declarative configuration language.
- Ansible Automation Platform (AAP) provides enterprise-grade automation capabilities, including workflows, RBAC, and integrations with CI/CD pipelines.
- Event-Driven Ansible (EDA) listens for events (for example, webhook triggers, alerts, or Git changes) and automatically triggers Ansible playbooks in response.

## GitOps (GitHub Hooks)

GitHub Webhooks (or Hooks for Github Enterprise) are a way to trigger actions based on events in your GitHub repository. For example, you can use hooks to automatically run tests, deploy code, or provision infrastructure whenever code is pushed to the repository. To set up a GitHub webhook, follow these steps:

1. Navigate to your GitHub repository and click Settings
2. In the Webhooks or Hooks section, click Add webhook
3. Enter the payload URL, which is the endpoint that will receive the webhook events
4. Choose the events that will trigger the webhook, such as push or pull request
5. Save the webhook

## Terraform

- Terraform is an open-source infrastructure as code (IaC) tool that allows you to define and provision infrastructure using a high-level configuration language.

- For example, On-Demand Environments can be provisioned using the On-Demand Environments provider.

### Setting up Terraform

1. **Install Terraform**: Download and install Terraform from the [official website](https://www.terraform.io/downloads.html). For information about IBM Terraform for Z and LinuxONE offering see the [documentation website](https://www.ibm.com/docs/en/terraform-z-linuxone).
2. **ConfigureOn-Demand Environments provider**: Set up the On-Demand Environments provider in your Terraform configuration file.
```
     provider "ode" {
       ode_host     = "https://your-ode-hostname:port"
       ode_username = "your-ode-user"
       ode_password = "your-ode-password"
       ode_tls = {
          ca_file = file ("ca-cert.pem")
          server_name = "example.com"
       }

     }
```

### Writing Terraform configuration

Create a Terraform configuration file (for example, `main.tf`) to define the resources that you want to provision. The following example configuration is for provisioning a z/OS instance:
```
     resource "ode_instance" "zos_25" {
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

## Ansible Automation Platform

Ansible Automation Platform is a powerful tool for automating the provisioning, configuration, and management of IT environments.

- **Agentless architecture**: Ansible does not require any agents to be installed on the managed nodes, making it easy to set up and use.
- **Idempotency**: Ansible ensures that the desired state of the system is achieved, regardless of the current state, by applying changes only when necessary.
- **Extensibility**: Ansible can be extended with custom modules and plugins to meet specific automation needs.
- **Integration**: Ansible integrates with a wide range of tools and platforms, including cloud providers, container orchestration systems, and more.

## Event-Driven Ansible

Use Event-Driven Ansible to trigger playbooks based on events such as:

- Terraform apply completion
- Git push events
- Monitoring alerts

For example, use a webhook from GitHub to trigger an EDA rulebook that runs a playbook to configure a newly provisioned z/OS instance.

```
    - name: Git Push Trigger
      sources:
        - ansible.eda.webhook:
            port: 5000
      rules:
        - name: Configure on Git Push
          condition: event.payload.ref == 'refs/heads/main'
          action:
            run_playbook:
              name: configure_ode_instance.yml

```

For sample code, please visit https://github.com/IBM/terraform-ibm-z-linuxone-samples