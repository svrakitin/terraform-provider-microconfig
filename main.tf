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
