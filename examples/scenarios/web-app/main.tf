# Web Application Deployment
# Deploys a frontend, API, and Redis session store

terraform {
  required_providers {
    dokploy = {
      source  = "reserve-protocol/dokploy"
      version = "~> 0.1"
    }
  }
}

provider "dokploy" {
  host    = var.dokploy_host
  api_key = var.dokploy_api_key
}

# Project
resource "dokploy_project" "webapp" {
  name        = "my-webapp"
  description = "Production web application"
}

# Environments
resource "dokploy_environment" "prod" {
  project_id  = dokploy_project.webapp.id
  name        = "prod"
  description = "Production environment"
}

resource "dokploy_environment" "staging" {
  project_id  = dokploy_project.webapp.id
  name        = "staging"
  description = "Staging environment"
}

# Production Applications
resource "dokploy_application" "frontend" {
  environment_id = dokploy_environment.prod.id
  name           = "frontend"
  description    = "React frontend application"
}

resource "dokploy_application" "api" {
  environment_id = dokploy_environment.prod.id
  name           = "api"
  description    = "Node.js API server"
}

# Session store
resource "dokploy_redis" "sessions" {
  environment_id = dokploy_environment.prod.id
  name           = "sessions"
  description    = "Redis session storage"
  docker_image   = "redis:7-alpine"
}

# Staging Applications
resource "dokploy_application" "frontend_staging" {
  environment_id = dokploy_environment.staging.id
  name           = "frontend"
  description    = "Staging frontend"
}

resource "dokploy_application" "api_staging" {
  environment_id = dokploy_environment.staging.id
  name           = "api"
  description    = "Staging API"
}

# Variables
variable "dokploy_host" {
  description = "Dokploy server URL"
  type        = string
}

variable "dokploy_api_key" {
  description = "Dokploy API key"
  type        = string
  sensitive   = true
}

# Outputs
output "production" {
  value = {
    frontend_id = dokploy_application.frontend.id
    api_id      = dokploy_application.api.id
    redis_id    = dokploy_redis.sessions.id
  }
}

output "staging" {
  value = {
    frontend_id = dokploy_application.frontend_staging.id
    api_id      = dokploy_application.api_staging.id
  }
}
