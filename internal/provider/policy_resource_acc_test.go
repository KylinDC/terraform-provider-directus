package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_policy", "policies"),
		Steps: []resource.TestStep{
			// Create and verify
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "test" {
  name         = "AccTest Basic Policy"
  description  = "Created by acceptance test"
  icon         = "check_circle"
  app_access   = true
  admin_access = false
  enforce_tfa  = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("directus_policy.test", "id"),
					resource.TestCheckResourceAttr("directus_policy.test", "name", "AccTest Basic Policy"),
					resource.TestCheckResourceAttr("directus_policy.test", "description", "Created by acceptance test"),
					resource.TestCheckResourceAttr("directus_policy.test", "icon", "check_circle"),
					resource.TestCheckResourceAttr("directus_policy.test", "app_access", "true"),
					resource.TestCheckResourceAttr("directus_policy.test", "admin_access", "false"),
					resource.TestCheckResourceAttr("directus_policy.test", "enforce_tfa", "false"),
				),
			},
			// ImportState
			{
				ResourceName:      "directus_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPolicy_allFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_policy", "policies"),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "full" {
  name         = "AccTest Full Policy"
  description  = "Policy with all fields"
  icon         = "admin_panel_settings"
  admin_access = true
  app_access   = true
  enforce_tfa  = true
  ip_access    = "10.0.0.0/8,192.168.1.0/24"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("directus_policy.full", "id"),
					resource.TestCheckResourceAttr("directus_policy.full", "name", "AccTest Full Policy"),
					resource.TestCheckResourceAttr("directus_policy.full", "description", "Policy with all fields"),
					resource.TestCheckResourceAttr("directus_policy.full", "icon", "admin_panel_settings"),
					resource.TestCheckResourceAttr("directus_policy.full", "admin_access", "true"),
					resource.TestCheckResourceAttr("directus_policy.full", "app_access", "true"),
					resource.TestCheckResourceAttr("directus_policy.full", "enforce_tfa", "true"),
					resource.TestCheckResourceAttr("directus_policy.full", "ip_access", "10.0.0.0/8,192.168.1.0/24"),
				),
			},
			// ImportState
			{
				ResourceName:      "directus_policy.full",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPolicy_minimal(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_policy", "policies"),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "minimal" {
  name = "AccTest Minimal Policy"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("directus_policy.minimal", "id"),
					resource.TestCheckResourceAttr("directus_policy.minimal", "name", "AccTest Minimal Policy"),
					resource.TestCheckResourceAttr("directus_policy.minimal", "admin_access", "false"),
					resource.TestCheckResourceAttr("directus_policy.minimal", "app_access", "false"),
					resource.TestCheckResourceAttr("directus_policy.minimal", "enforce_tfa", "false"),
				),
			},
			// ImportState
			{
				ResourceName:      "directus_policy.minimal",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccPolicy_ipAccessLifecycle verifies that ip_access can be set, updated, and removed.
func TestAccPolicy_ipAccessLifecycle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_policy", "policies"),
		Steps: []resource.TestStep{
			// Create with ip_access
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "ip_test" {
  name       = "AccTest IP Policy"
  ip_access  = "10.0.0.0/8"
  app_access = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_policy.ip_test", "ip_access", "10.0.0.0/8"),
				),
			},
			// Update ip_access to multiple ranges
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "ip_test" {
  name       = "AccTest IP Policy"
  ip_access  = "10.0.0.0/8,192.168.0.0/16,172.16.0.0/12"
  app_access = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_policy.ip_test", "ip_access", "10.0.0.0/8,192.168.0.0/16,172.16.0.0/12"),
				),
			},
			// Import with ip_access set
			{
				ResourceName:      "directus_policy.ip_test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccPolicy_booleanDefaults verifies that all boolean fields default correctly
// and can be individually toggled.
func TestAccPolicy_booleanDefaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_policy", "policies"),
		Steps: []resource.TestStep{
			// Create with defaults (all false)
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "bool_test" {
  name = "AccTest Bool Defaults"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_policy.bool_test", "admin_access", "false"),
					resource.TestCheckResourceAttr("directus_policy.bool_test", "app_access", "false"),
					resource.TestCheckResourceAttr("directus_policy.bool_test", "enforce_tfa", "false"),
				),
			},
			// Toggle only admin_access
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "bool_test" {
  name         = "AccTest Bool Defaults"
  admin_access = true
  app_access   = false
  enforce_tfa  = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_policy.bool_test", "admin_access", "true"),
					resource.TestCheckResourceAttr("directus_policy.bool_test", "app_access", "false"),
					resource.TestCheckResourceAttr("directus_policy.bool_test", "enforce_tfa", "false"),
				),
			},
			// Toggle only enforce_tfa
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "bool_test" {
  name         = "AccTest Bool Defaults"
  admin_access = false
  app_access   = false
  enforce_tfa  = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_policy.bool_test", "admin_access", "false"),
					resource.TestCheckResourceAttr("directus_policy.bool_test", "app_access", "false"),
					resource.TestCheckResourceAttr("directus_policy.bool_test", "enforce_tfa", "true"),
				),
			},
			// All true
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "bool_test" {
  name         = "AccTest Bool Defaults"
  admin_access = true
  app_access   = true
  enforce_tfa  = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_policy.bool_test", "admin_access", "true"),
					resource.TestCheckResourceAttr("directus_policy.bool_test", "app_access", "true"),
					resource.TestCheckResourceAttr("directus_policy.bool_test", "enforce_tfa", "true"),
				),
			},
			// ImportState
			{
				ResourceName:      "directus_policy.bool_test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPolicy_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_policy", "policies"),
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "update" {
  name         = "AccTest Update Policy"
  description  = "Before update"
  icon         = "edit"
  app_access   = true
  admin_access = false
  enforce_tfa  = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_policy.update", "name", "AccTest Update Policy"),
					resource.TestCheckResourceAttr("directus_policy.update", "description", "Before update"),
					resource.TestCheckResourceAttr("directus_policy.update", "enforce_tfa", "false"),
				),
			},
			// Update all mutable fields
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "update" {
  name         = "AccTest Update Policy Modified"
  description  = "After update"
  icon         = "security"
  app_access   = true
  admin_access = true
  enforce_tfa  = true
  ip_access    = "172.16.0.0/12"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_policy.update", "name", "AccTest Update Policy Modified"),
					resource.TestCheckResourceAttr("directus_policy.update", "description", "After update"),
					resource.TestCheckResourceAttr("directus_policy.update", "icon", "security"),
					resource.TestCheckResourceAttr("directus_policy.update", "admin_access", "true"),
					resource.TestCheckResourceAttr("directus_policy.update", "enforce_tfa", "true"),
					resource.TestCheckResourceAttr("directus_policy.update", "ip_access", "172.16.0.0/12"),
				),
			},
			// ImportState after update
			{
				ResourceName:      "directus_policy.update",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
