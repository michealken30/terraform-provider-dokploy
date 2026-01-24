terraform {
  required_providers {
    dokploy = {
      source = "registry.terraform.io/thefrozenfire/dokploy"
    }
  }
}

variable "dokploy_host" {
  description = "Dokploy server URL"
  type        = string
}

variable "dokploy_api_key" {
  description = "Dokploy API key"
  type        = string
  sensitive   = true
}

provider "dokploy" {
  host    = var.dokploy_host
  api_key = var.dokploy_api_key
}

# List all projects
data "dokploy_projects" "all" {}

# Get a specific project by name
data "dokploy_project" "example" {
  name = "my-project"
}

# List all servers
data "dokploy_servers" "all" {}

# Get a specific server by name
data "dokploy_server" "example" {
  name = "my-server"
}

# List all SSH keys
data "dokploy_ssh_keys" "all" {}

# Get a specific SSH key by name
data "dokploy_ssh_key" "example" {
  name = "my-ssh-key"
}

# Get environments for a project
data "dokploy_environments" "example" {
  project_id = data.dokploy_project.example.id
}

# Output examples
output "all_project_names" {
  value = [for p in data.dokploy_projects.all.projects : p.name]
}

output "project_environments" {
  value = [for e in data.dokploy_project.example.environments : e.name]
}

output "all_server_names" {
  value = [for s in data.dokploy_servers.all.servers : s.name]
}

output "server_ip" {
  value = data.dokploy_server.example.ip_address
}

output "ssh_key_id" {
  value = data.dokploy_ssh_key.example.id
}
