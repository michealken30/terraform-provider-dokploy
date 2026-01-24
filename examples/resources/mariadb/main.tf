# Create a MariaDB database

resource "dokploy_project" "example" {
  name = "my-project"
}

resource "dokploy_environment" "prod" {
  project_id = dokploy_project.example.id
  name       = "prod"
}

resource "dokploy_mariadb" "main" {
  environment_id = dokploy_environment.prod.id
  name           = "mariadb"
  description    = "MariaDB database"
  database_name  = "wordpress"
  database_user  = "wp_user"
  docker_image   = "mariadb:11"
}

output "mariadb_id" {
  value = dokploy_mariadb.main.id
}
