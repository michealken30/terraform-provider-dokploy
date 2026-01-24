# Basic domain with Let's Encrypt SSL
resource "dokploy_domain" "example" {
  host             = "app.example.com"
  application_id   = dokploy_application.web.id
  https            = true
  certificate_type = "letsencrypt"
}

# Domain with path prefix
resource "dokploy_domain" "api" {
  host             = "api.example.com"
  path             = "/v1"
  application_id   = dokploy_application.api.id
  https            = true
  certificate_type = "letsencrypt"
  strip_path       = true
}

# Domain for compose service
resource "dokploy_domain" "compose_web" {
  host             = "web.example.com"
  compose_id       = dokploy_compose.stack.id
  service_name     = "nginx"
  https            = true
  certificate_type = "letsencrypt"
}
