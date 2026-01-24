# Create a PostgreSQL database

resource "dokploy_project" "example" {
  name = "my-project"
}

resource "dokploy_environment" "prod" {
  project_id = dokploy_project.example.id
  name       = "prod"
}

resource "dokploy_postgres" "main" {
  environment_id = dokploy_environment.prod.id
  name           = "main-db"
  description    = "Primary PostgreSQL database"
  database_name  = "appdb"
  database_user  = "appuser"
  docker_image   = "postgres:16-alpine"
}

output "postgres_id" {
  value = dokploy_postgres.main.id
}

output "postgres_password" {
  value     = dokploy_postgres.main.database_password
  sensitive = true
}
