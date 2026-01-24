# Basic authentication for an application
resource "dokploy_security" "admin" {
  username       = "admin"
  password       = var.admin_password
  application_id = dokploy_application.admin_panel.id
}

# Multiple users for the same application
resource "dokploy_security" "readonly" {
  username       = "viewer"
  password       = var.viewer_password
  application_id = dokploy_application.admin_panel.id
}
