# Create a MySQL database

resource "dokploy_project" "example" {
  name = "my-project"
}

resource "dokploy_environment" "prod" {
  project_id = dokploy_project.example.id
  name       = "prod"
}

resource "dokploy_mysql" "main" {
  environment_id = dokploy_environment.prod.id
  name           = "mysql-db"
  description    = "MySQL database"
  database_name  = "myapp"
  database_user  = "appuser"
  docker_image   = "mysql:8.0"
}

output "mysql_id" {
  value = dokploy_mysql.main.id
}
