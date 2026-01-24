# Bitbucket provider with app password
resource "dokploy_bitbucket" "main" {
  name    = "company-bitbucket"
  auth_id = "user-auth-id"

  bitbucket_username       = "my-username"
  bitbucket_email          = "user@example.com"
  app_password             = var.bitbucket_app_password
  bitbucket_workspace_name = "my-workspace"
}

# Bitbucket with API token
resource "dokploy_bitbucket" "api_token" {
  name    = "bitbucket-api"
  auth_id = "user-auth-id"

  bitbucket_username = "my-username"
  api_token          = var.bitbucket_api_token
}
