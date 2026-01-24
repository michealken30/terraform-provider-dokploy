# Upload an SSL certificate

resource "dokploy_certificate" "wildcard" {
  name             = "wildcard-cert"
  certificate_data = file("certs/fullchain.pem")
  private_key      = file("certs/privkey.pem")
  auto_renew       = false
  organization_id  = var.organization_id
}

variable "organization_id" {
  description = "Dokploy organization ID"
  type        = string
}

output "certificate_id" {
  value = dokploy_certificate.wildcard.id
}
