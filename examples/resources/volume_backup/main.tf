# Back up application uploads volume
resource "dokploy_volume_backup" "app_uploads" {
  name            = "uploads-backup"
  volume_name     = "myapp-uploads"
  prefix          = "uploads"
  cron_expression = "0 2 * * *" # Daily at 2 AM
  destination_id  = dokploy_destination.s3.id

  application_id = dokploy_application.web.id
  service_type   = "application"

  keep_latest_count = 7
  enabled           = true
}

# Back up Redis persistence volume with service stop
resource "dokploy_volume_backup" "redis_data" {
  name            = "redis-backup"
  volume_name     = "redis-data"
  prefix          = "redis"
  cron_expression = "0 3 * * *" # Daily at 3 AM
  destination_id  = dokploy_destination.s3.id

  redis_id     = dokploy_redis.cache.id
  service_type = "redis"

  # Stop Redis during backup for consistency
  turn_off          = true
  keep_latest_count = 14
  enabled           = true
}

# Back up compose service volume
resource "dokploy_volume_backup" "compose_data" {
  name            = "compose-data-backup"
  volume_name     = "app-data"
  prefix          = "compose-data"
  cron_expression = "0 4 * * 0" # Weekly on Sunday at 4 AM
  destination_id  = dokploy_destination.s3.id

  compose_id   = dokploy_compose.stack.id
  service_type = "compose"
  service_name = "worker" # Specific service in compose stack

  keep_latest_count = 4
  enabled           = true
}
