# Database Cluster Setup
# Creates multiple database types for a data platform

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
resource "dokploy_project" "data_platform" {
  name        = "data-platform"
  description = "Data storage and processing services"
}

# Environment
resource "dokploy_environment" "data" {
  project_id  = dokploy_project.data_platform.id
  name        = "data"
  description = "Data services environment"
}

# Primary PostgreSQL - Application data
resource "dokploy_postgres" "primary" {
  environment_id = dokploy_environment.data.id
  name           = "primary-db"
  description    = "Primary application database"
  database_name  = "app_production"
  database_user  = "app_user"
  docker_image   = "postgres:16-alpine"
}

# Analytics PostgreSQL - Read-optimized
resource "dokploy_postgres" "analytics" {
  environment_id = dokploy_environment.data.id
  name           = "analytics-db"
  description    = "Analytics and reporting database"
  database_name  = "analytics"
  database_user  = "analyst"
  docker_image   = "postgres:16-alpine"
}

# MongoDB - Document storage
resource "dokploy_mongo" "documents" {
  environment_id = dokploy_environment.data.id
  name           = "document-store"
  description    = "Document and event storage"
  docker_image   = "mongo:7"
}

# Redis - Caching layer
resource "dokploy_redis" "cache" {
  environment_id = dokploy_environment.data.id
  name           = "cache"
  description    = "Application cache"
  docker_image   = "redis:7-alpine"
}

# Redis - Message queue
resource "dokploy_redis" "queue" {
  environment_id = dokploy_environment.data.id
  name           = "queue"
  description    = "Job queue (BullMQ)"
  docker_image   = "redis:7-alpine"
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
output "postgres_primary" {
  value = {
    id       = dokploy_postgres.primary.id
    name     = dokploy_postgres.primary.database_name
    user     = dokploy_postgres.primary.database_user
    password = dokploy_postgres.primary.database_password
  }
  sensitive = true
}

output "postgres_analytics" {
  value = {
    id       = dokploy_postgres.analytics.id
    name     = dokploy_postgres.analytics.database_name
    user     = dokploy_postgres.analytics.database_user
    password = dokploy_postgres.analytics.database_password
  }
  sensitive = true
}

output "mongodb" {
  value = {
    id       = dokploy_mongo.documents.id
    password = dokploy_mongo.documents.database_password
  }
  sensitive = true
}

output "redis_endpoints" {
  value = {
    cache = dokploy_redis.cache.id
    queue = dokploy_redis.queue.id
  }
}
