package environment_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/acctest"
)

func TestAccEnvironmentResource_basic(t *testing.T) {
	acctest.SkipIfNotAccTest(t)
	projectName := acctest.RandomName("project")
	envName := acctest.RandomName("env")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: acctest.ConfigEnvironment(projectName, envName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_environment.test", "name", envName),
					resource.TestCheckResourceAttrSet("dokploy_environment.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_environment.test", "project_id"),
				),
			},
			// ImportState
			{
				ResourceName:      "dokploy_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEnvironmentResource_withDescription(t *testing.T) {
	acctest.SkipIfNotAccTest(t)
	projectName := acctest.RandomName("project")
	envName := acctest.RandomName("env")
	description := "Test environment description"
	updatedDescription := "Updated description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with description
			{
				Config: acctest.ConfigEnvironmentWithDescription(projectName, envName, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_environment.test", "name", envName),
					resource.TestCheckResourceAttr("dokploy_environment.test", "description", description),
				),
			},
			// Update description
			{
				Config: acctest.ConfigEnvironmentWithDescription(projectName, envName, updatedDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_environment.test", "name", envName),
					resource.TestCheckResourceAttr("dokploy_environment.test", "description", updatedDescription),
				),
			},
		},
	})
}

func TestAccEnvironmentResource_import(t *testing.T) {
	acctest.SkipIfNotAccTest(t)
	projectName := acctest.RandomName("project")
	envName := acctest.RandomName("env")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigEnvironment(projectName, envName),
			},
			{
				ResourceName:      "dokploy_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
