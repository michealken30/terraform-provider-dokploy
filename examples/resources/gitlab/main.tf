# Basic GitLab provider with OAuth
resource "dokploy_gitlab" "main" {
  name       = "company-gitlab"
  gitlab_url = "https://gitlab.com"
  auth_id    = "user-auth-id"

  # OAuth configuration
  application_id = var.gitlab_app_id
  secret         = var.gitlab_secret
  redirect_uri   = "https://dokploy.example.com/oauth/callback"
}

# Self-hosted GitLab with group restriction
resource "dokploy_gitlab" "self_hosted" {
  name       = "internal-gitlab"
  gitlab_url = "https://gitlab.company.com"
  auth_id    = "user-auth-id"

  # Restrict to specific group
  group_name = "my-team"

  # OAuth tokens
  access_token  = var.gitlab_access_token
  refresh_token = var.gitlab_refresh_token
}
