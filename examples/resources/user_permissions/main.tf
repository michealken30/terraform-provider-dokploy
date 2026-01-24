# Grant a user access to specific projects
resource "dokploy_user_permissions" "developer" {
  user_id = "user-abc123"

  # Project access
  accessed_projects = [
    dokploy_project.frontend.id,
    dokploy_project.backend.id,
  ]

  # Service permissions
  can_create_services = true
  can_delete_services = false

  # Environment permissions
  can_create_environments = true
  can_delete_environments = false

  # Infrastructure access
  can_access_to_docker       = true
  can_access_to_api          = true
  can_access_to_ssh_keys     = false
  can_access_to_git_providers = true
  can_access_to_traefik_files = false
}

# Full admin-like permissions for a team lead
resource "dokploy_user_permissions" "team_lead" {
  user_id = "user-xyz789"

  # Full project access
  can_create_projects = true
  can_delete_projects = true

  # Full service access
  can_create_services = true
  can_delete_services = true

  # Full environment access
  can_create_environments = true
  can_delete_environments = true

  # Full infrastructure access
  can_access_to_docker        = true
  can_access_to_api           = true
  can_access_to_ssh_keys      = true
  can_access_to_git_providers = true
  can_access_to_traefik_files = true
}

# Read-only access for an auditor
resource "dokploy_user_permissions" "auditor" {
  user_id = "user-audit456"

  # Can view all projects but not modify
  accessed_projects = [
    dokploy_project.frontend.id,
    dokploy_project.backend.id,
    dokploy_project.database.id,
  ]

  # No create/delete permissions (all default to false)
  can_access_to_api = true
}
