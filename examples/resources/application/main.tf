# Create an application within an environment

resource "dokploy_project" "example" {
  name = "my-project"
}

resource "dokploy_environment" "prod" {
  project_id = dokploy_project.example.id
  name       = "prod"
}

resource "dokploy_application" "api" {
  environment_id = dokploy_environment.prod.id
  name           = "api-server"
  description    = "Backend API application"
}

output "application_id" {
  value = dokploy_application.api.id
}
