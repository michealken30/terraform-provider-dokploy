package project_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/acctest"
)

func TestAccProjectResource_basic(t *testing.T) {
	acctest.SkipIfNotAccTest(t)
	name := acctest.RandomName("project")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: acctest.ConfigProject(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_project.test", "name", name),
					resource.TestCheckResourceAttrSet("dokploy_project.test", "id"),
				),
			},
			// ImportState
			{
				ResourceName:      "dokploy_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccProjectResource_withDescription(t *testing.T) {
	acctest.SkipIfNotAccTest(t)
	name := acctest.RandomName("project")
	description := "Test project description"
	updatedDescription := "Updated description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with description
			{
				Config: acctest.ConfigProjectWithDescription(name, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_project.test", "name", name),
					resource.TestCheckResourceAttr("dokploy_project.test", "description", description),
				),
			},
			// Update description
			{
				Config: acctest.ConfigProjectWithDescription(name, updatedDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_project.test", "name", name),
					resource.TestCheckResourceAttr("dokploy_project.test", "description", updatedDescription),
				),
			},
		},
	})
}

func TestAccProjectResource_update(t *testing.T) {
	acctest.SkipIfNotAccTest(t)
	name := acctest.RandomName("project")
	updatedName := acctest.RandomName("project-updated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: acctest.ConfigProject(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_project.test", "name", name),
				),
			},
			// Update name
			{
				Config: acctest.ConfigProject(updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dokploy_project.test", "name", updatedName),
				),
			},
		},
	})
}

func TestAccProjectResource_import(t *testing.T) {
	acctest.SkipIfNotAccTest(t)
	name := acctest.RandomName("project")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigProject(name),
			},
			{
				ResourceName:      "dokploy_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
