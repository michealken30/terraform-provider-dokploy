# Create a Docker Compose service

resource "dokploy_project" "example" {
  name = "my-project"
}

resource "dokploy_environment" "prod" {
  project_id = dokploy_project.example.id
  name       = "prod"
}

resource "dokploy_compose" "monitoring" {
  environment_id = dokploy_environment.prod.id
  name           = "monitoring"
  description    = "Prometheus and Grafana stack"
  compose_type   = "docker-compose"
  compose_file   = <<-EOT
    version: '3.8'
    services:
      prometheus:
        image: prom/prometheus:latest
        ports:
          - "9090:9090"
      grafana:
        image: grafana/grafana:latest
        ports:
          - "3000:3000"
  EOT
}

output "compose_id" {
  value = dokploy_compose.monitoring.id
}
