---
page_title: "Managing IBM On-demand Environments with GitOps, Terraform, and Ansible Automation Platform"
subcategory: ""
description: |-
  The following demonstrates how Git, Terraform and Ansible Automation Platform can be used together to achieve Infrastructure as Code and Configuration as Code.
---

# Managing IBM On-demand Environments (ODE) with GitOps, Terraform, and Ansible Automation Platform

## Introduction
In today’s fast-paced DevOps world, automation is the backbone of scalable, reliable, and secure infrastructure. When working with IBM On-demand Environments, combining GitOps, Terraform, Ansible Automation Platform, and Event-Driven Ansible creates a powerful, event-driven, and declarative infrastructure pipeline.

In this guide, we will explore how to use GitOps, Terraform, Ansible Automation Platform, and Event-Driven Ansible to provision and configure IBM On-demand Environments. This powerful combination of tools allows for efficient and automated management of ODE instances, ensuring that your infrastructure is always up-to-date and configured correctly.

##  Key Components Overview
1. GitOps
GitOps is a methodology that uses Git as the single source of truth for infrastructure and application configurations. It enables version-controlled, auditable, and automated deployments.

2. Terraform
Terraform by HashiCorp is an Infrastructure as Code (IaC) tool that allows you to define and provision infrastructure using a declarative configuration language.

3. Ansible Automation Platform (AAP)
AAP provides enterprise-grade automation capabilities, including workflows, RBAC, and integrations with CI/CD pipelines.

4. Event-Driven Ansible (EDA)
EDA listens for events (e.g., webhook triggers, alerts, or Git changes) and automatically triggers Ansible playbooks in response.

## GitOps (GitHub Hooks)

GitHub Hooks are a way to trigger actions based on events in your GitHub repository. For example, you can use hooks to automatically run tests, deploy code, or provision infrastructure whenever code is pushed to the repository. To set up a GitHub Hook, follow these steps:

1. Navigate to your GitHub repository and click on "Settings".
2. In the "Webhooks" section, click on "Add webhook".
3. Enter the payload URL, which is the endpoint that will receive the webhook events.
4. Choose the events that will trigger the webhook, such as "push" or "pull request".
5. Save the webhook.

## Terraform

- Terraform is an open-source infrastructure as code (IaC) tool that allows you to define and provision infrastructure using a high-level configuration language. 

- For example, IBM On-demand Environments can be provisioned using the IBM ODE provider

### Setting Up Terraform

1. **Install Terraform**: Download and install Terraform from the [official website](https://www.terraform.io/downloads.html).
2. **Configure IBM ODE Provider**: Set up the IBM ODE provider in your Terraform configuration file. 
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

### Writing Terraform Configuration

Create a Terraform configuration file (e.g., `main.tf`) to define the resources you want to provision. Below is an example configuration for provisioning an ODE instance:
```
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


## Ansible Automation Platform

Ansible Automation Platform is a powerful tool for automating the provisioning, configuration, and management of IT environments. 

- **Agentless Architecture**: Ansible does not require any agents to be installed on the managed nodes, making it easy to set up and use.
- **Idempotency**: Ansible ensures that the desired state of the system is achieved, regardless of the current state, by applying changes only when necessary.
- **Extensibility**: Ansible can be extended with custom modules and plugins to meet specific automation needs.
- **Integration**: Ansible integrates with a wide range of tools and platforms, including cloud providers, container orchestration systems, and more.

## Event-Driven Ansible

Use Event-Driven Ansible to trigger playbooks based on events such as:

- Terraform apply completion
- Git push events
- Monitoring alerts

Example: Use a webhook from GitHub to trigger an EDA rulebook that runs a playbook to configure a newly provisioned VM.
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

## Conclusion

In this guide, we have explored how to use GitHub Hooks, Terraform, and Ansible Automation Platform, along with Event-Driven Ansible, to provision and configure IBM On-demand Environments. By leveraging these tools, you can create a robust and automated infrastructure that meets your needs.

For sample codes, please visit https://github.com/IBM/terraform-ibm-z-linuxone-samples

Happy automating!
