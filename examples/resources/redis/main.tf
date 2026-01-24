# Create a Redis cache

resource "dokploy_project" "example" {
  name = "my-project"
}

resource "dokploy_environment" "prod" {
  project_id = dokploy_project.example.id
  name       = "prod"
}

resource "dokploy_redis" "cache" {
  environment_id = dokploy_environment.prod.id
  name           = "cache"
  description    = "Redis cache layer"
  docker_image   = "redis:7-alpine"
}

output "redis_id" {
  value = dokploy_redis.cache.id
}
