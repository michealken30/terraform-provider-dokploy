# Register a remote server for deployments

# First, create an SSH key for authentication
resource "dokploy_ssh_key" "deploy" {
  name            = "deploy-key"
  description     = "Deployment SSH key"
  private_key     = file("~/.ssh/id_rsa")
  public_key      = file("~/.ssh/id_rsa.pub")
  organization_id = var.organization_id
}

resource "dokploy_server" "worker" {
  name        = "worker-01"
  description = "Worker server for background jobs"
  ip_address  = "192.168.1.100"
  port        = 22
  username    = "deploy"
  ssh_key_id  = dokploy_ssh_key.deploy.id
}

variable "organization_id" {
  description = "Dokploy organization ID"
  type        = string
}

output "server_id" {
  value = dokploy_server.worker.id
}
