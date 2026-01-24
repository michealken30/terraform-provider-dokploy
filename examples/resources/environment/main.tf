# Create an environment within a project

resource "dokploy_project" "example" {
  name = "my-project"
}

resource "dokploy_environment" "staging" {
  project_id  = dokploy_project.example.id
  name        = "staging"
  description = "Staging environment"
}

output "environment_id" {
  value = dokploy_environment.staging.id
}
