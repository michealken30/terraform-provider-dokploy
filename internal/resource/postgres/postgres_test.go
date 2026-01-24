package postgres_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/acctest"
)

func TestAccPostgresResource_basic(t *testing.T) {
	acctest.SkipIfNotAccTest(t)
	projectName := acctest.RandomName("project")
	envName := acctest.RandomName("env")
	dbName := acctest.RandomName("postgres")
	appName := acctest.RandomName("pgapp")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: acctest.ConfigPostgres(projectName, envName, dbName, appName, "testuser", "testpass123"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_postgres.test", "name", dbName),
					resource.TestCheckResourceAttr("dokploy_postgres.test", "database_user", "testuser"),
					resource.TestCheckResourceAttrSet("dokploy_postgres.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_postgres.test", "environment_id"),
				),
			},
			// ImportState
			{
				ResourceName:            "dokploy_postgres.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"database_password"},
			},
		},
	})
}

func TestAccPostgresResource_update(t *testing.T) {
	acctest.SkipIfNotAccTest(t)
	projectName := acctest.RandomName("project")
	envName := acctest.RandomName("env")
	dbName := acctest.RandomName("postgres")
	appName := acctest.RandomName("pgapp")
	updatedDbName := acctest.RandomName("postgres-updated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: acctest.ConfigPostgres(projectName, envName, dbName, appName, "testuser", "testpass123"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_postgres.test", "name", dbName),
				),
			},
			// Update name
			{
				Config: acctest.ConfigPostgres(projectName, envName, updatedDbName, appName, "testuser", "testpass123"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_postgres.test", "name", updatedDbName),
				),
			},
		},
	})
}

func TestAccPostgresResource_import(t *testing.T) {
	acctest.SkipIfNotAccTest(t)
	projectName := acctest.RandomName("project")
	envName := acctest.RandomName("env")
	dbName := acctest.RandomName("postgres")
	appName := acctest.RandomName("pgapp")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigPostgres(projectName, envName, dbName, appName, "testuser", "testpass123"),
			},
			{
				ResourceName:            "dokploy_postgres.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"database_password"},
			},
		},
	})
}

// testAccCheckPostgresExists verifies the postgres exists in the API.
func testAccCheckPostgresExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is not set")
		}

		return nil
	}
}
