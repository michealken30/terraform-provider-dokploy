# Slack notification
resource "dokploy_notification" "slack" {
  name              = "deploy-alerts"
  notification_type = "slack"

  # Event triggers
  app_deploy      = true
  app_build_error = true
  database_backup = true

  # Slack-specific configuration
  slack_webhook_url = var.slack_webhook_url
  slack_channel     = "#deployments"
}

# Discord notification
resource "dokploy_notification" "discord" {
  name              = "build-alerts"
  notification_type = "discord"

  # Event triggers
  app_build_error = true
  docker_cleanup  = true

  # Discord-specific configuration
  discord_webhook_url = var.discord_webhook_url
  discord_decoration  = true
}

# Telegram notification
resource "dokploy_notification" "telegram" {
  name              = "server-alerts"
  notification_type = "telegram"

  # Event triggers
  server_threshold = true
  dokploy_restart  = true

  # Telegram-specific configuration
  telegram_bot_token        = var.telegram_bot_token
  telegram_chat_id          = var.telegram_chat_id
  telegram_message_thread_id = ""
}

# Email notification
resource "dokploy_notification" "email" {
  name              = "backup-alerts"
  notification_type = "email"

  # Event triggers
  database_backup = true
  volume_backup   = true

  # Email-specific configuration
  email_smtp_server  = "smtp.example.com"
  email_smtp_port    = 587
  email_username     = "alerts@example.com"
  email_password     = var.smtp_password
  email_from_address = "alerts@example.com"
  email_to_addresses = ["admin@example.com", "ops@example.com"]
}
