# Redirect www to non-www
resource "dokploy_redirect" "www_to_apex" {
  regex          = "^https://www\\.example\\.com/(.*)"
  replacement    = "https://example.com/$1"
  permanent      = true
  application_id = dokploy_application.web.id
}

# Redirect old paths to new paths
resource "dokploy_redirect" "legacy_api" {
  regex          = "^/api/v1/(.*)"
  replacement    = "/api/v2/$1"
  permanent      = false
  application_id = dokploy_application.api.id
}

# Redirect all HTTP to HTTPS
resource "dokploy_redirect" "force_https" {
  regex          = "^http://(.*)$"
  replacement    = "https://$1"
  permanent      = true
  application_id = dokploy_application.web.id
}
