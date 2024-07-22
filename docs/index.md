# Adinusa Provider

This Terraform provider allows you to manage [Adinusa](https://adinusa.id/) resources via Terraform.

## Example Usage

```hcl
provider "adinusa" {
  main_api_url = "https://example.adinusa.id/api"
  api_url      = "https://example.adinusa.id/api/pro-training"
  username     = "admin"
  password     = "adminpass"
}
```

## Argument Reference

* `main_api_url` (Required) - The main API URL of Adinusa.
* `api_url` (Required) - The API URL of Adinusa for academy or pro training.
* `username` (Required) - The username used to authenticate with Adinusa.
* `password` (Required) - The password used to authenticate with Adinusa.