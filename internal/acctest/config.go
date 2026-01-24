package acctest

import (
	"fmt"
	"math/rand/v2"
)

// RandomName generates a unique name with a prefix for test resources.
// Names are prefixed with "tf-test-" to make them easily identifiable for cleanup.
func RandomName(prefix string) string {
	return fmt.Sprintf("tf-test-%s-%d", prefix, rand.Int32())
}

// ConfigProject returns Terraform configuration for a project resource.
func ConfigProject(name string) string {
	return fmt.Sprintf(`
resource "dokploy_project" "test" {
  name = %q
}
`, name)
}

// ConfigProjectWithDescription returns Terraform configuration for a project with description.
func ConfigProjectWithDescription(name, description string) string {
	return fmt.Sprintf(`
resource "dokploy_project" "test" {
  name        = %q
  description = %q
}
`, name, description)
}

// ConfigEnvironment returns Terraform configuration for an environment resource.
func ConfigEnvironment(projectName, envName string) string {
	return fmt.Sprintf(`
resource "dokploy_project" "test" {
  name = %q
}

resource "dokploy_environment" "test" {
  name       = %q
  project_id = dokploy_project.test.id
}
`, projectName, envName)
}

// ConfigEnvironmentWithDescription returns Terraform configuration for an environment with description.
func ConfigEnvironmentWithDescription(projectName, envName, description string) string {
	return fmt.Sprintf(`
resource "dokploy_project" "test" {
  name = %q
}

resource "dokploy_environment" "test" {
  name        = %q
  description = %q
  project_id  = dokploy_project.test.id
}
`, projectName, envName, description)
}

// ConfigApplication returns Terraform configuration for an application resource.
func ConfigApplication(projectName, envName, appName string) string {
	return fmt.Sprintf(`
resource "dokploy_project" "test" {
  name = %q
}

resource "dokploy_environment" "test" {
  name       = %q
  project_id = dokploy_project.test.id
}

resource "dokploy_application" "test" {
  name           = %q
  environment_id = dokploy_environment.test.id
}
`, projectName, envName, appName)
}

// ConfigPostgres returns Terraform configuration for a postgres resource.
func ConfigPostgres(projectName, envName, dbName, appName, user, password string) string {
	return fmt.Sprintf(`
resource "dokploy_project" "test" {
  name = %q
}

resource "dokploy_environment" "test" {
  name       = %q
  project_id = dokploy_project.test.id
}

resource "dokploy_postgres" "test" {
  name              = %q
  app_name          = %q
  environment_id    = dokploy_environment.test.id
  database_name     = %q
  database_user     = %q
  database_password = %q
}
`, projectName, envName, dbName, appName, dbName, user, password)
}

// ConfigCompose returns Terraform configuration for a compose resource.
func ConfigCompose(projectName, envName, composeName, composeType string) string {
	return fmt.Sprintf(`
resource "dokploy_project" "test" {
  name = %q
}

resource "dokploy_environment" "test" {
  name       = %q
  project_id = dokploy_project.test.id
}

resource "dokploy_compose" "test" {
  name           = %q
  environment_id = dokploy_environment.test.id
  compose_type   = %q
}
`, projectName, envName, composeName, composeType)
}
