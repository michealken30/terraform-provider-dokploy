# Terraform Provider for Dokploy

[![Tests](https://github.com/thefrozenfire/terraform-provider-dokploy/actions/workflows/test.yml/badge.svg)](https://github.com/thefrozenfire/terraform-provider-dokploy/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/thefrozenfire/terraform-provider-dokploy)](https://goreportcard.com/report/github.com/thefrozenfire/terraform-provider-dokploy)

The Dokploy provider allows you to manage [Dokploy](https://dokploy.com) infrastructure as code. Dokploy is an open-source, self-hostable Platform as a Service (PaaS) that simplifies deploying applications, databases, and Docker Compose stacks.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23 (to build the provider)
- A running Dokploy instance with API access

## Installation

### From Terraform Registry (Recommended)

```hcl
terraform {
  required_providers {
    dokploy = {
      source  = "thefrozenfire/dokploy"
      version = "~> 0.1"
    }
  }
}
```

### From Source

```bash
git clone https://github.com/thefrozenfire/terraform-provider-dokploy.git
cd terraform-provider-dokploy
make install
```

## Authentication

The provider requires a Dokploy host URL and API key:

```hcl
provider "dokploy" {
  host    = "https://dokploy.example.com"
  api_key = var.dokploy_api_key
}
```

Or via environment variables:

```bash
export DOKPLOY_HOST="https://dokploy.example.com"
export DOKPLOY_API_KEY="your-api-key"
```

## Quick Start

```hcl
# Create a project
resource "dokploy_project" "myapp" {
  name        = "my-application"
  description = "Production application"
}

# Create an environment
resource "dokploy_environment" "prod" {
  project_id  = dokploy_project.myapp.id
  name        = "production"
  description = "Production environment"
}

# Deploy an application
resource "dokploy_application" "web" {
  environment_id = dokploy_environment.prod.id
  name           = "web"
  description    = "Web frontend"
}

# Add a PostgreSQL database
resource "dokploy_postgres" "db" {
  environment_id    = dokploy_environment.prod.id
  name              = "postgres"
  database_name     = "myapp_prod"
  database_user     = "myapp"
  database_password = var.db_password
  docker_image      = "postgres:16-alpine"
}
```

## Resources

| Resource | Description |
|----------|-------------|
| `dokploy_bootstrap` | Bootstrap a new Dokploy instance |
| `dokploy_project` | Manage projects |
| `dokploy_environment` | Manage environments within projects |
| `dokploy_application` | Deploy applications |
| `dokploy_compose` | Deploy Docker Compose stacks |
| `dokploy_postgres` | PostgreSQL databases |
| `dokploy_mysql` | MySQL databases |
| `dokploy_mariadb` | MariaDB databases |
| `dokploy_mongo` | MongoDB databases |
| `dokploy_redis` | Redis instances |
| `dokploy_server` | Remote server configuration |
| `dokploy_sshkey` | SSH key management |
| `dokploy_registry` | Container registry credentials |
| `dokploy_certificate` | SSL/TLS certificates |
| `dokploy_destination` | Backup destinations (S3-compatible) |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `dokploy_projects` | List all projects |
| `dokploy_project` | Get a project by ID or name |
| `dokploy_servers` | List all servers |
| `dokploy_server` | Get a server by ID or name |
| `dokploy_sshkeys` | List all SSH keys |
| `dokploy_sshkey` | Get an SSH key by ID or name |
| `dokploy_environments` | List environments in a project |
| `dokploy_application` | Get an application by ID |

## Examples

See the [examples](./examples) directory for complete configurations:

- [Bootstrap](./examples/resources/bootstrap) - Initial Dokploy setup
- [Project & Environment](./examples/resources/project) - Basic project structure
- [Applications](./examples/resources/application) - Application deployment
- [Databases](./examples/resources/postgres) - Database provisioning
- [Docker Compose](./examples/resources/compose) - Multi-container deployments
- [Full Stack](./examples/scenarios/full-stack) - Complete production setup

## Development

### Building

```bash
make build
```

### Testing

```bash
# Unit tests
make test

# Acceptance tests (requires running Dokploy instance)
export DOKPLOY_HOST="http://localhost:3000"
export DOKPLOY_API_KEY="your-api-key"
make testacc
```

### Linting

```bash
make lint
```

### Local Installation

```bash
make install
```

This installs the provider to `~/.terraform.d/plugins/` for local development.

### Documentation

Generate provider documentation:

```bash
make docs
```

## Contributing

Contributions are welcome! Please see our [testing guide](TESTING.md) for information on running tests.

## License

MIT License - see [LICENSE](LICENSE) for details.
