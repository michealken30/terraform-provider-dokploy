# Bootstrap a fresh Dokploy instance
# Creates admin user and generates API key for provider configuration

resource "dokploy_bootstrap" "init" {
  host       = "http://localhost:3000"
  email      = "admin@example.com"
  password   = "securepassword123"
  first_name = "Admin"
  last_name  = "User"
}

output "api_key" {
  description = "API key for provider configuration"
  value       = dokploy_bootstrap.init.api_key
  sensitive   = true
}
