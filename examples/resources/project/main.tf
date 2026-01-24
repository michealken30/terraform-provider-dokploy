# Create a Dokploy project

resource "dokploy_project" "example" {
  name        = "my-project"
  description = "Example project for applications"
}

output "project_id" {
  value = dokploy_project.example.id
}
