# Self-hosted Gitea provider
resource "dokploy_gitea" "main" {
  name      = "company-gitea"
  gitea_url = "https://gitea.example.com"

  # OAuth configuration
  client_id     = var.gitea_client_id
  client_secret = var.gitea_client_secret
  redirect_uri  = "https://dokploy.example.com/oauth/callback"
}

# Gitea with organization restriction
resource "dokploy_gitea" "org" {
  name      = "gitea-org"
  gitea_url = "https://gitea.example.com"

  gitea_username    = "deploy-user"
  organization_name = "my-org"

  access_token = var.gitea_access_token
}
