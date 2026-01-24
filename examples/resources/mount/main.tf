# Docker volume mount
resource "dokploy_mount" "data" {
  type         = "volume"
  volume_name  = "app-data"
  mount_path   = "/app/data"
  service_type = "application"
  service_id   = dokploy_application.web.id
}

# Bind mount from host
resource "dokploy_mount" "config" {
  type         = "bind"
  host_path    = "/etc/app/config"
  mount_path   = "/app/config"
  service_type = "application"
  service_id   = dokploy_application.web.id
}

# File mount with inline content
resource "dokploy_mount" "env_file" {
  type         = "file"
  content      = <<-EOT
    DATABASE_URL=postgresql://user:pass@localhost/db
    SECRET_KEY=abc123
  EOT
  mount_path   = "/app/.env"
  service_type = "application"
  service_id   = dokploy_application.web.id
}

# Mount for PostgreSQL database
resource "dokploy_mount" "postgres_data" {
  type         = "volume"
  volume_name  = "postgres-data"
  mount_path   = "/var/lib/postgresql/data"
  service_type = "postgres"
  service_id   = dokploy_postgres.db.id
}
