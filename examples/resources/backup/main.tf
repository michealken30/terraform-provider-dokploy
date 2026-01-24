# Daily PostgreSQL backup
resource "dokploy_backup" "postgres_daily" {
  schedule          = "0 0 * * *" # Daily at midnight
  enabled           = true
  prefix            = "postgres-daily"
  destination_id    = dokploy_destination.s3.id
  database          = "myapp"
  database_type     = "postgres"
  backup_type       = "database"
  postgres_id       = dokploy_postgres.db.id
  keep_latest_count = 7
}

# Weekly MySQL backup with retention
resource "dokploy_backup" "mysql_weekly" {
  schedule          = "0 0 * * 0" # Weekly on Sunday
  enabled           = true
  prefix            = "mysql-weekly"
  destination_id    = dokploy_destination.s3.id
  database          = "production"
  database_type     = "mysql"
  backup_type       = "database"
  mysql_id          = dokploy_mysql.db.id
  keep_latest_count = 4
}

# MongoDB backup
resource "dokploy_backup" "mongo" {
  schedule       = "0 */6 * * *" # Every 6 hours
  enabled        = true
  prefix         = "mongo"
  destination_id = dokploy_destination.s3.id
  database       = "documents"
  database_type  = "mongo"
  backup_type    = "database"
  mongo_id       = dokploy_mongo.db.id
}
