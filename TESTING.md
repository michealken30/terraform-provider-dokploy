# Testing Guide

This document describes how to run tests for the Dokploy Terraform Provider.

## Prerequisites

- Go 1.23 or later
- Docker and Docker Compose
- A Dokploy instance for acceptance tests

## Unit Tests

Unit tests don't require a running Dokploy instance:

```bash
make test
```

## Acceptance Tests

Acceptance tests require a running Dokploy instance and valid credentials.

### Using Local Dokploy (Docker Compose)

1. Start Dokploy:
   ```bash
   docker compose up -d
   ```

2. Wait for Dokploy to be ready and complete the bootstrap process through the UI or API.

3. Set environment variables:
   ```bash
   export DOKPLOY_HOST="http://localhost:3000"
   export DOKPLOY_API_KEY="your-api-key"
   ```

4. Run acceptance tests:
   ```bash
   make testacc
   ```

### Using Remote Dokploy Instance

1. Set environment variables:
   ```bash
   export DOKPLOY_HOST="https://your-dokploy-instance.com"
   export DOKPLOY_API_KEY="your-api-key"
   ```

2. Run acceptance tests:
   ```bash
   make testacc
   ```

## Test Sweepers

Test sweepers clean up resources left behind by failed tests. Resources with the `tf-test-` prefix are removed.

**Warning:** This is destructive! Only run on development instances.

```bash
make sweep
```

## Running Specific Tests

Run a specific test:

```bash
TF_ACC=1 go test -v -run TestAccProjectResource_basic ./internal/resource/project/...
```

Run tests for a specific resource:

```bash
TF_ACC=1 go test -v ./internal/resource/project/...
```

## Test Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `TF_ACC` | Yes | Set to `1` to enable acceptance tests |
| `DOKPLOY_HOST` | Yes | Dokploy server URL (e.g., `http://localhost:3000`) |
| `DOKPLOY_API_KEY` | Yes | Dokploy API key for authentication |

## Writing Tests

### Test Naming Convention

- `TestAccXxxResource_basic` - Basic create/read/update/delete test
- `TestAccXxxResource_withYyy` - Test with specific feature (e.g., `withDescription`)
- `TestAccXxxResource_update` - Test update functionality
- `TestAccXxxResource_import` - Test import state functionality

### Test Utilities

The `internal/acctest` package provides:

- `ProtoV6ProviderFactories` - Provider factories for acceptance tests
- `TestAccPreCheck(t)` - Validates environment before running tests
- `RandomName(prefix)` - Generates unique names with `tf-test-` prefix
- `ConfigXxx()` - Helper functions for generating Terraform configs

### Example Test

```go
func TestAccProjectResource_basic(t *testing.T) {
    acctest.SkipIfNotAccTest(t)
    name := acctest.RandomName("project")

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { acctest.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: acctest.ConfigProject(name),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("dokploy_project.test", "name", name),
                    resource.TestCheckResourceAttrSet("dokploy_project.test", "id"),
                ),
            },
            {
                ResourceName:      "dokploy_project.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}
```

## CI/CD

The GitHub Actions workflow automatically runs:

1. **Build** - Verifies the code compiles
2. **Lint** - Runs golangci-lint
3. **Unit Tests** - Runs tests in `internal/client/`
4. **Acceptance Tests** - Runs full acceptance tests against a Docker Compose Dokploy instance

Acceptance tests only run on:
- Push to main/master
- PRs from the same repository (not forks)
