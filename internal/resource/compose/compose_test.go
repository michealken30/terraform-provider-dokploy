package compose_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/acctest"
)

func TestAccComposeResource_basic(t *testing.T) {
	acctest.SkipIfNotAccTest(t)
	projectName := acctest.RandomName("project")
	envName := acctest.RandomName("env")
	composeName := acctest.RandomName("compose")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: acctest.ConfigCompose(projectName, envName, composeName, "docker-compose"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_compose.test", "name", composeName),
					resource.TestCheckResourceAttr("dokploy_compose.test", "compose_type", "docker-compose"),
					resource.TestCheckResourceAttrSet("dokploy_compose.test", "id"),
					resource.TestCheckResourceAttrSet("dokploy_compose.test", "environment_id"),
				),
			},
			// ImportState
			{
				ResourceName:      "dokploy_compose.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccComposeResource_update(t *testing.T) {
	acctest.SkipIfNotAccTest(t)
	projectName := acctest.RandomName("project")
	envName := acctest.RandomName("env")
	composeName := acctest.RandomName("compose")
	updatedComposeName := acctest.RandomName("compose-updated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: acctest.ConfigCompose(projectName, envName, composeName, "docker-compose"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_compose.test", "name", composeName),
				),
			},
			// Update name
			{
				Config: acctest.ConfigCompose(projectName, envName, updatedComposeName, "docker-compose"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_compose.test", "name", updatedComposeName),
				),
			},
		},
	})
}

func TestAccComposeResource_import(t *testing.T) {
	acctest.SkipIfNotAccTest(t)
	projectName := acctest.RandomName("project")
	envName := acctest.RandomName("env")
	composeName := acctest.RandomName("compose")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(projectName, envName, composeName, "docker-compose"),
			},
			{
				ResourceName:      "dokploy_compose.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
