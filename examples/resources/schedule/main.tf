# Application cron job
resource "dokploy_schedule" "cleanup" {
  name            = "daily-cleanup"
  cron_expression = "0 0 * * *"
  command         = "php artisan cleanup:run"
  shell_type      = "bash"
  schedule_type   = "application"
  application_id  = dokploy_application.web.id
  enabled         = true
  timezone        = "America/New_York"
}

# Compose service cron job
resource "dokploy_schedule" "db_optimize" {
  name            = "weekly-optimize"
  cron_expression = "0 3 * * 0"
  command         = "mysqlcheck --optimize --all-databases"
  shell_type      = "sh"
  schedule_type   = "compose"
  compose_id      = dokploy_compose.stack.id
  service_name    = "mysql"
  enabled         = true
}

# Server-level cron job
resource "dokploy_schedule" "docker_prune" {
  name            = "docker-prune"
  cron_expression = "0 4 * * *"
  command         = "docker system prune -af --volumes"
  shell_type      = "bash"
  schedule_type   = "server"
  server_id       = dokploy_server.worker.id
  enabled         = true
}

# Health check script
resource "dokploy_schedule" "health_check" {
  name            = "health-check"
  cron_expression = "*/5 * * * *"
  command         = "curl -sf http://localhost:3000/health || exit 1"
  shell_type      = "bash"
  schedule_type   = "application"
  application_id  = dokploy_application.api.id
  enabled         = true
}
