# Adinusa Terraform Provider

This Terraform provider allows you to manage [Adinusa](https://adinusa.id/) resources via Terraform.

## Installation

To use this provider, you need to install it. You can do this by adding it to your Terraform configuration.

### Terraform Configuration

Add the following to your `main.tf` file to use the Jumpserver provider:

```hcl
terraform {
  required_providers {
    jumpserver = {
      source  = "atwatanmalikm/adinusa"
      version = "~> 1.0.0"
    }
  }
}

provider "adinusa" {
  main_api_url = "https://example.adinusa.id/api"
  api_url      = "https://example.adinusa.id/api/pro-training"
  username     = "admin"
  password     = "adminpass"
}

```

## Resources

This provider supports the following resources:

* `adinusa_class`
* `adinusa_enroll_user`

## Resource Definitions

For detailed information on each resource, see the following documentation:

* [Adinusa Class Resource](docs/resources/class.md)
* [Adinusa Enroll User Resource](docs/resources/enroll_user.md)