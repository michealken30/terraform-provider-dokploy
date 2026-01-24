# Create a MongoDB database

resource "dokploy_project" "example" {
  name = "my-project"
}

resource "dokploy_environment" "prod" {
  project_id = dokploy_project.example.id
  name       = "prod"
}

resource "dokploy_mongo" "main" {
  environment_id = dokploy_environment.prod.id
  name           = "mongodb"
  description    = "MongoDB document store"
  docker_image   = "mongo:7"
}

output "mongo_id" {
  value = dokploy_mongo.main.id
}
