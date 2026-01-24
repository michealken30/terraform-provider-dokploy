# Create an SSH key for server authentication

resource "dokploy_ssh_key" "deploy" {
  name            = "deploy-key"
  description     = "SSH key for server deployments"
  private_key     = file("~/.ssh/id_rsa")
  public_key      = file("~/.ssh/id_rsa.pub")
  organization_id = var.organization_id
}

variable "organization_id" {
  description = "Dokploy organization ID (from bootstrap output)"
  type        = string
}

output "ssh_key_id" {
  value = dokploy_ssh_key.deploy.id
}
