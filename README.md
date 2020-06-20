# terraform-provider-microconfig

This provider enables [microconfig](https://github.com/microconfig/microconfig) usage in Terraform.

## Requirements

* `microconfig` binary in `$PATH`

## Example

```terraform
provider "microconfig" {
  source_dir      = "fixtures"
  entrypoint      = "/usr/bin/microconfig"
}

resource "microconfig_service" "payment-backend" {
  environment = "dev"
  name        = "payment-backend"
}

output "data" {
  value = yamldecode(microconfig_service.payment-backend.data["application.yaml"])
}
```