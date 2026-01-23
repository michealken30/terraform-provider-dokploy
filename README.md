# Terraform Provider for Dokploy

A Terraform provider for managing Dokploy infrastructure.

## Status

**Phase 1: Data Sources (Read-Only)** - Complete

The provider currently supports read-only data sources for querying Dokploy infrastructure:
- `dokploy_projects` - List all projects
- `dokploy_project` - Get a single project by ID or name
- `dokploy_servers` - List all servers
- `dokploy_server` - Get a single server by ID or name
- `dokploy_ssh_keys` - List all SSH keys
- `dokploy_ssh_key` - Get a single SSH key by ID or name
- `dokploy_environments` - List environments in a project
- `dokploy_application` - Get a single application by ID

**Phase 2: HCL Generator** - Complete

A tool to generate Terraform configuration from Dokploy backup JSON files.

**Phase 3: Resources (Future)**

Create/update/delete operations will be added after validation on a staging environment.

## Building

```bash
go build -o terraform-provider-dokploy
```

## Provider Configuration

```hcl
provider "dokploy" {
  host    = "https://dokploy.example.com"
  api_key = var.dokploy_api_key
}
```

Configuration can also be set via environment variables:
- `DOKPLOY_HOST` - Dokploy server URL
- `DOKPLOY_API_KEY` - API key

## Example Usage

```hcl
# List all projects
data "dokploy_projects" "all" {}

# Get a specific project by name
data "dokploy_project" "example" {
  name = "my-project"
}

# Get environments for a project
data "dokploy_environments" "example" {
  project_id = data.dokploy_project.example.id
}

# Output project environment names
output "project_environments" {
  value = [for e in data.dokploy_project.example.environments : e.name]
}
```

See `examples/data-sources/main.tf` for more examples.

## HCL Generator

Generate Terraform configuration from backup JSON files:

```bash
cd tools/hcl-generator
go build -o hcl-generator

./hcl-generator --backup-dir /path/to/backup --output ./generated
```

The generator creates:
- `provider.tf` - Provider configuration
- `ssh_keys.tf` - SSH key resources
- `servers.tf` - Server resources
- `projects.tf` - Projects and environments
- `applications.tf` - Application resources
- `compose.tf` - Docker Compose services
- `databases.tf` - Database resources (Postgres, MySQL, etc.)
- `registries.tf` - Container registries
- `imports.tf` - Import blocks for existing resources

**Security Notes:**
- Private SSH keys are NOT included in generated HCL
- Database passwords are NOT included
- Registry passwords are NOT included

## Local Development

1. Build the provider:
   ```bash
   go build -o terraform-provider-dokploy
   ```

2. Create a `.terraformrc` file in your home directory:
   ```hcl
   provider_installation {
     dev_overrides {
       "registry.terraform.io/reserve-protocol/dokploy" = "/path/to/terraform-provider-dokploy"
     }
     direct {}
   }
   ```

3. Use the provider in your Terraform configuration.

## Architecture

```
terraform-provider-dokploy/
├── main.go                        # Provider entry point
├── internal/
│   ├── provider/
│   │   └── provider.go            # Provider configuration
│   ├── client/
│   │   ├── client.go              # API client
│   │   └── models.go              # Data models
│   └── datasource/
│       ├── projects/              # dokploy_projects data source
│       ├── project/               # dokploy_project data source
│       ├── servers/               # dokploy_servers data source
│       ├── server/                # dokploy_server data source
│       ├── sshkeys/               # dokploy_ssh_keys data source
│       ├── sshkey/                # dokploy_ssh_key data source
│       ├── environments/          # dokploy_environments data source
│       └── application/           # dokploy_application data source
├── tools/
│   └── hcl-generator/             # HCL generation tool
└── examples/
    └── data-sources/              # Example configurations
```
