package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRole_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_role", "roles"),
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccProviderConfig() + `
resource "directus_role" "test" {
  name        = "AccTest Basic Role"
  description = "Created by acceptance test"
  icon        = "person"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("directus_role.test", "id"),
					resource.TestCheckResourceAttr("directus_role.test", "name", "AccTest Basic Role"),
					resource.TestCheckResourceAttr("directus_role.test", "description", "Created by acceptance test"),
					resource.TestCheckResourceAttr("directus_role.test", "icon", "person"),
				),
			},
			// ImportState
			{
				ResourceName:            "directus_role.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"children", "users"},
			},
		},
	})
}

func TestAccRole_allFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_role", "roles"),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "directus_role" "parent_full" {
  name        = "AccTest Full Parent Role"
  description = "Parent for all-fields test"
  icon        = "admin_panel_settings"
}

resource "directus_role" "full" {
  name        = "AccTest Full Role"
  description = "Role with all configurable fields"
  icon        = "verified_user"
  parent      = directus_role.parent_full.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("directus_role.full", "id"),
					resource.TestCheckResourceAttr("directus_role.full", "name", "AccTest Full Role"),
					resource.TestCheckResourceAttr("directus_role.full", "description", "Role with all configurable fields"),
					resource.TestCheckResourceAttr("directus_role.full", "icon", "verified_user"),
					resource.TestCheckResourceAttrPair("directus_role.full", "parent", "directus_role.parent_full", "id"),
				),
			},
			// Import the child role (has parent)
			{
				ResourceName:            "directus_role.full",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"children", "users"},
			},
			// Import the parent role
			{
				ResourceName:            "directus_role.parent_full",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"children", "users"},
			},
		},
	})
}

func TestAccRole_minimal(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_role", "roles"),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "directus_role" "minimal" {
  name = "AccTest Minimal Role"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("directus_role.minimal", "id"),
					resource.TestCheckResourceAttr("directus_role.minimal", "name", "AccTest Minimal Role"),
				),
			},
			// ImportState
			{
				ResourceName:            "directus_role.minimal",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"children", "users"},
			},
		},
	})
}

func TestAccRole_hierarchy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_role", "roles"),
		Steps: []resource.TestStep{
			// Create parent-child-grandchild hierarchy
			{
				Config: testAccProviderConfig() + `
resource "directus_role" "parent" {
  name        = "AccTest Parent Role"
  description = "Parent role"
  icon        = "admin_panel_settings"
}

resource "directus_role" "child" {
  name        = "AccTest Child Role"
  description = "Child of parent"
  icon        = "people"
  parent      = directus_role.parent.id
}

resource "directus_role" "grandchild" {
  name        = "AccTest Grandchild Role"
  description = "Child of child"
  icon        = "person"
  parent      = directus_role.child.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("directus_role.parent", "id"),
					resource.TestCheckResourceAttrSet("directus_role.child", "id"),
					resource.TestCheckResourceAttrSet("directus_role.grandchild", "id"),
					// Verify parent relationships
					resource.TestCheckResourceAttrPair("directus_role.child", "parent", "directus_role.parent", "id"),
					resource.TestCheckResourceAttrPair("directus_role.grandchild", "parent", "directus_role.child", "id"),
				),
			},
			// Import parent
			{
				ResourceName:            "directus_role.parent",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"children", "users"},
			},
			// Import child (has parent)
			{
				ResourceName:            "directus_role.child",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"children", "users"},
			},
		},
	})
}

func TestAccRole_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_role", "roles"),
		Steps: []resource.TestStep{
			// Create two standalone roles
			{
				Config: testAccProviderConfig() + `
resource "directus_role" "role_a" {
  name        = "AccTest Role A"
  description = "Original"
  icon        = "person"
}

resource "directus_role" "role_b" {
  name        = "AccTest Role B"
  description = "Standalone"
  icon        = "people"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_role.role_a", "name", "AccTest Role A"),
					resource.TestCheckResourceAttr("directus_role.role_b", "name", "AccTest Role B"),
				),
			},
			// Update: change fields + make role_b a child of role_a
			{
				Config: testAccProviderConfig() + `
resource "directus_role" "role_a" {
  name        = "AccTest Role A Updated"
  description = "Updated description"
  icon        = "shield"
}

resource "directus_role" "role_b" {
  name        = "AccTest Role B Updated"
  description = "Now a child"
  icon        = "child_care"
  parent      = directus_role.role_a.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_role.role_a", "name", "AccTest Role A Updated"),
					resource.TestCheckResourceAttr("directus_role.role_a", "icon", "shield"),
					resource.TestCheckResourceAttr("directus_role.role_b", "name", "AccTest Role B Updated"),
					resource.TestCheckResourceAttrPair("directus_role.role_b", "parent", "directus_role.role_a", "id"),
				),
			},
			// Update: remove parent (make role_b standalone again)
			{
				Config: testAccProviderConfig() + `
resource "directus_role" "role_a" {
  name        = "AccTest Role A Updated"
  description = "Updated description"
  icon        = "shield"
}

resource "directus_role" "role_b" {
  name        = "AccTest Role B Standalone"
  description = "Standalone again"
  icon        = "people"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_role.role_b", "name", "AccTest Role B Standalone"),
				),
			},
			// ImportState after update
			{
				ResourceName:            "directus_role.role_a",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"children", "users"},
			},
		},
	})
}
