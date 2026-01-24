# Import Existing Dokploy Resources
#
# This example demonstrates how to import existing Dokploy infrastructure
# into Terraform management using the import block syntax (Terraform 1.5+).
#
# Workflow:
# 1. Run `terraform apply -target=data.*` to discover existing resource IDs
# 2. Add import blocks for resources you want to import
# 3. Run `terraform plan -generate-config-out=generated.tf` to generate config
# 4. Review and customize the generated configuration
# 5. Run `terraform apply` to complete the import

terraform {
  required_providers {
    dokploy = {
      source = "reserve-protocol/dokploy"
    }
  }
}

provider "dokploy" {
  # Configure via environment variables:
  # export DOKPLOY_HOST="https://your-dokploy-instance.com"
  # export DOKPLOY_API_KEY="your-api-key"
}

# =============================================================================
# Step 1: Resource Discovery
# =============================================================================
# Run: terraform apply -target=data.dokploy_projects.all
# to see all available resource IDs

# Discover all projects
data "dokploy_projects" "all" {}

output "available_projects" {
  description = "All projects available for import"
  value = {
    for p in data.dokploy_projects.all.projects : p.name => {
      id          = p.id
      description = p.description
    }
  }
}

# Discover all servers
data "dokploy_servers" "all" {}

output "available_servers" {
  description = "All servers available for import"
  value = {
    for s in data.dokploy_servers.all.servers : s.name => {
      id         = s.id
      ip_address = s.ip_address
      status     = s.status
    }
  }
}

# Discover all SSH keys
data "dokploy_ssh_keys" "all" {}

output "available_ssh_keys" {
  description = "All SSH keys available for import"
  value = {
    for k in data.dokploy_ssh_keys.all.ssh_keys : k.name => {
      id = k.id
    }
  }
}

# Discover all registries
data "dokploy_registries" "all" {}

output "available_registries" {
  description = "All registries available for import"
  value = {
    for r in data.dokploy_registries.all.registries : r.name => {
      id   = r.id
      url  = r.url
      type = r.type
    }
  }
}

# Discover all certificates
data "dokploy_certificates" "all" {}

output "available_certificates" {
  description = "All certificates available for import"
  value = {
    for c in data.dokploy_certificates.all.certificates : c.name => {
      id         = c.id
      auto_renew = c.auto_renew
    }
  }
}

# Discover all destinations
data "dokploy_destinations" "all" {}

output "available_destinations" {
  description = "All backup destinations available for import"
  value = {
    for d in data.dokploy_destinations.all.destinations : d.name => {
      id     = d.id
      bucket = d.bucket
      region = d.region
    }
  }
}

# =============================================================================
# Step 2: Environment Discovery (per project)
# =============================================================================
# After identifying which project to import, uncomment and customize:

# data "dokploy_environments" "project_envs" {
#   project_id = "YOUR_PROJECT_ID_HERE"
# }
#
# output "available_environments" {
#   description = "All environments and services in the project"
#   value = {
#     for env in data.dokploy_environments.project_envs.environments : env.name => {
#       id           = env.id
#       applications = { for app in env.applications : app.name => app.id }
#       compose      = { for c in env.compose : c.name => c.id }
#       postgres     = { for pg in env.postgres : pg.name => pg.id }
#       mysql        = { for m in env.mysql : m.name => m.id }
#       mariadb      = { for m in env.mariadb : m.name => m.id }
#       mongo        = { for m in env.mongo : m.name => m.id }
#       redis        = { for r in env.redis : r.name => r.id }
#     }
#   }
# }

# =============================================================================
# Step 3: Import Blocks (Terraform 1.5+)
# =============================================================================
# Uncomment and customize these blocks with your resource IDs.
# Then run: terraform plan -generate-config-out=generated.tf

# Import a project
# import {
#   to = dokploy_project.existing
#   id = "abc123-project-id"
# }

# Import an environment
# import {
#   to = dokploy_environment.production
#   id = "def456-environment-id"
# }

# Import an application
# import {
#   to = dokploy_application.api
#   id = "ghi789-application-id"
# }

# Import a compose service
# import {
#   to = dokploy_compose.monitoring
#   id = "jkl012-compose-id"
# }

# Import databases
# import {
#   to = dokploy_postgres.main_db
#   id = "mno345-postgres-id"
# }

# import {
#   to = dokploy_mysql.legacy_db
#   id = "pqr678-mysql-id"
# }

# import {
#   to = dokploy_redis.cache
#   id = "stu901-redis-id"
# }

# Import infrastructure resources
# import {
#   to = dokploy_server.production
#   id = "vwx234-server-id"
# }

# import {
#   to = dokploy_ssh_key.deploy_key
#   id = "yza567-sshkey-id"
# }

# import {
#   to = dokploy_registry.dockerhub
#   id = "bcd890-registry-id"
# }

# import {
#   to = dokploy_certificate.wildcard
#   id = "efg123-certificate-id"
# }

# import {
#   to = dokploy_destination.s3_backup
#   id = "hij456-destination-id"
# }
