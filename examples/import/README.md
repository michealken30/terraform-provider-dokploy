# Importing Existing Dokploy Resources

This example demonstrates how to import existing Dokploy infrastructure into Terraform management.

## Prerequisites

- Terraform 1.5+ (for import block syntax)
- Access to your Dokploy instance
- Dokploy API key with read permissions

## Configuration

Set your Dokploy credentials:

```bash
export DOKPLOY_HOST="https://your-dokploy-instance.com"
export DOKPLOY_API_KEY="your-api-key"
```

## Import Workflow

### Step 1: Initialize Terraform

```bash
terraform init
```

### Step 2: Discover Existing Resources

Run the data sources to see what resources exist:

```bash
terraform apply -target=data.dokploy_projects.all \
                -target=data.dokploy_servers.all \
                -target=data.dokploy_ssh_keys.all \
                -target=data.dokploy_registries.all \
                -target=data.dokploy_certificates.all \
                -target=data.dokploy_destinations.all
```

This outputs all available resource IDs organized by name.

### Step 3: Discover Environment Contents

For each project you want to import, uncomment and customize the `dokploy_environments` data source in `main.tf`:

```hcl
data "dokploy_environments" "project_envs" {
  project_id = "your-project-id-from-step-2"
}
```

Then run:

```bash
terraform apply -target=data.dokploy_environments.project_envs
```

This shows all applications, databases, and compose services within each environment.

### Step 4: Add Import Blocks

Edit `main.tf` and uncomment/customize the import blocks for resources you want to import:

```hcl
import {
  to = dokploy_project.my_project
  id = "abc123-actual-project-id"
}

import {
  to = dokploy_application.my_app
  id = "def456-actual-application-id"
}
```

### Step 5: Generate Configuration

Use Terraform to generate the resource configuration:

```bash
terraform plan -generate-config-out=generated.tf
```

This creates `generated.tf` with the full resource definitions.

### Step 6: Review and Customize

Review `generated.tf` and make any necessary adjustments:

- Remove computed-only attributes
- Add sensitive values (passwords, keys) that won't be imported
- Organize resources into logical files
- Add lifecycle blocks if needed

### Step 7: Import Resources

```bash
terraform apply
```

## Using the Import Helper Script

For bulk imports, use the provided helper script:

```bash
# Generate import blocks for all resources
./scripts/dokploy-import.sh > imports.tf

# Generate configuration
terraform plan -generate-config-out=generated.tf

# Apply
terraform apply
```

## Classic Import Command (Terraform < 1.5)

For older Terraform versions, use the import command:

```bash
# First, create empty resource blocks in your config
resource "dokploy_project" "my_project" {}

# Then import
terraform import dokploy_project.my_project abc123-project-id
```

## Tips

1. **Start small**: Import one project at a time to avoid overwhelming complexity
2. **Sensitive values**: Passwords and keys won't be imported - add them to your config after import
3. **Dependencies**: Import in order - projects first, then environments, then services
4. **State backup**: Always backup your state before bulk imports: `terraform state pull > backup.tfstate`

## Troubleshooting

### Resource not found during import

Verify the ID is correct by checking the data source outputs.

### Configuration drift after import

Run `terraform plan` to see differences between imported state and actual resource. Adjust your configuration to match.

### Sensitive values missing

Some attributes (passwords, private keys) are write-only and won't be read back during import. You must set these in your configuration.
