# Full Stack Application
# Complete production setup with monitoring, databases, and applications

terraform {
  required_providers {
    dokploy = {
      source  = "thefrozenfire/dokploy"
      version = "~> 0.1"
    }
  }
}

provider "dokploy" {
  host    = var.dokploy_host
  api_key = var.dokploy_api_key
}

# =============================================================================
# Projects
# =============================================================================

resource "dokploy_project" "app" {
  name        = "saas-platform"
  description = "SaaS platform services"
}

resource "dokploy_project" "infra" {
  name        = "infrastructure"
  description = "Monitoring and infrastructure"
}

# =============================================================================
# Environments
# =============================================================================

resource "dokploy_environment" "app_prod" {
  project_id  = dokploy_project.app.id
  name        = "app-prod"
  description = "Production application services"
}

resource "dokploy_environment" "infra_prod" {
  project_id  = dokploy_project.infra.id
  name        = "infra-prod"
  description = "Infrastructure services"
}

# =============================================================================
# Applications
# =============================================================================

resource "dokploy_application" "web" {
  environment_id = dokploy_environment.app_prod.id
  name           = "web"
  description    = "Next.js web application"
}

resource "dokploy_application" "api" {
  environment_id = dokploy_environment.app_prod.id
  name           = "api"
  description    = "GraphQL API server"
}

resource "dokploy_application" "worker" {
  environment_id = dokploy_environment.app_prod.id
  name           = "worker"
  description    = "Background job processor"
}

# =============================================================================
# Databases
# =============================================================================

resource "dokploy_postgres" "main" {
  environment_id = dokploy_environment.app_prod.id
  name           = "postgres"
  description    = "Primary PostgreSQL database"
  database_name  = "saas_production"
  database_user  = "saas_app"
  docker_image   = "postgres:16-alpine"
}

resource "dokploy_redis" "cache" {
  environment_id = dokploy_environment.app_prod.id
  name           = "redis-cache"
  description    = "Application cache"
  docker_image   = "redis:7-alpine"
}

resource "dokploy_redis" "queue" {
  environment_id = dokploy_environment.app_prod.id
  name           = "redis-queue"
  description    = "Job queue"
  docker_image   = "redis:7-alpine"
}

# =============================================================================
# Monitoring Stack
# =============================================================================

resource "dokploy_compose" "monitoring" {
  environment_id = dokploy_environment.infra_prod.id
  name           = "monitoring"
  description    = "Prometheus + Grafana + Loki"
  compose_type   = "docker-compose"
  compose_file   = <<-EOT
    version: '3.8'
    services:
      prometheus:
        image: prom/prometheus:latest
        volumes:
          - prometheus_data:/prometheus
        ports:
          - "9090:9090"

      grafana:
        image: grafana/grafana:latest
        volumes:
          - grafana_data:/var/lib/grafana
        ports:
          - "3001:3000"
        environment:
          - GF_SECURITY_ADMIN_PASSWORD=admin

      loki:
        image: grafana/loki:latest
        volumes:
          - loki_data:/loki
        ports:
          - "3100:3100"

    volumes:
      prometheus_data:
      grafana_data:
      loki_data:
  EOT
}

# =============================================================================
# Variables
# =============================================================================

variable "dokploy_host" {
  description = "Dokploy server URL"
  type        = string
}

variable "dokploy_api_key" {
  description = "Dokploy API key"
  type        = string
  sensitive   = true
}

# =============================================================================
# Outputs
# =============================================================================

output "applications" {
  value = {
    web    = dokploy_application.web.id
    api    = dokploy_application.api.id
    worker = dokploy_application.worker.id
  }
}

output "databases" {
  value = {
    postgres = {
      id       = dokploy_postgres.main.id
      name     = dokploy_postgres.main.database_name
      user     = dokploy_postgres.main.database_user
      password = dokploy_postgres.main.database_password
    }
    redis_cache = dokploy_redis.cache.id
    redis_queue = dokploy_redis.queue.id
  }
  sensitive = true
}

output "monitoring" {
  value = {
    compose_id = dokploy_compose.monitoring.id
  }
}
