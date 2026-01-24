# Configure a container registry

resource "dokploy_registry" "dockerhub" {
  registry_name   = "dockerhub"
  registry_type   = "dockerhub"
  registry_url    = "https://index.docker.io/v1/"
  username        = var.dockerhub_username
  password        = var.dockerhub_token
  organization_id = var.organization_id
}

variable "dockerhub_username" {
  description = "Docker Hub username"
  type        = string
}

variable "dockerhub_token" {
  description = "Docker Hub access token"
  type        = string
  sensitive   = true
}

variable "organization_id" {
  description = "Dokploy organization ID"
  type        = string
}

output "registry_id" {
  value = dokploy_registry.dockerhub.id
}
