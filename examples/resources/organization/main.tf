# Basic organization
resource "dokploy_organization" "main" {
  name = "My Company"
}

# Organization with logo
resource "dokploy_organization" "with_logo" {
  name = "Acme Corp"
  logo = "https://example.com/logo.png"
}
