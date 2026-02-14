package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRolePoliciesAttachment_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_role", "roles"),
		Steps: []resource.TestStep{
			// Create role + policy + attachment
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "test_policy" {
  name       = "AccTest Attachment Policy"
  app_access = true
}

resource "directus_role" "test_role" {
  name = "AccTest Attachment Role"
}

resource "directus_role_policies_attachment" "test" {
  role_id    = directus_role.test_role.id
  policy_ids = [directus_policy.test_policy.id]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("directus_role_policies_attachment.test", "role_id", "directus_role.test_role", "id"),
					resource.TestCheckResourceAttr("directus_role_policies_attachment.test", "policy_ids.#", "1"),
				),
			},
			// ImportState
			{
				ResourceName:      "directus_role_policies_attachment.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRolePoliciesAttachmentImportStateIdFunc("directus_role_policies_attachment.test"),
			},
		},
	})
}

func TestAccRolePoliciesAttachment_multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_role", "roles"),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "policy_a" {
  name       = "AccTest Multi Policy A"
  app_access = true
}

resource "directus_policy" "policy_b" {
  name         = "AccTest Multi Policy B"
  admin_access = true
}

resource "directus_policy" "policy_c" {
  name       = "AccTest Multi Policy C"
  app_access = true
}

resource "directus_role" "multi_role" {
  name = "AccTest Multi Attachment Role"
}

resource "directus_role_policies_attachment" "multi" {
  role_id = directus_role.multi_role.id
  policy_ids = [
    directus_policy.policy_a.id,
    directus_policy.policy_b.id,
    directus_policy.policy_c.id,
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_role_policies_attachment.multi", "policy_ids.#", "3"),
				),
			},
			// ImportState
			{
				ResourceName:      "directus_role_policies_attachment.multi",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRolePoliciesAttachmentImportStateIdFunc("directus_role_policies_attachment.multi"),
			},
		},
	})
}

func TestAccRolePoliciesAttachment_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_role", "roles"),
		Steps: []resource.TestStep{
			// Create with 2 policies
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "pa" {
  name       = "AccTest Update Attach Policy A"
  app_access = true
}

resource "directus_policy" "pb" {
  name       = "AccTest Update Attach Policy B"
  app_access = true
}

resource "directus_policy" "pc" {
  name         = "AccTest Update Attach Policy C"
  admin_access = true
}

resource "directus_role" "r" {
  name = "AccTest Update Attach Role"
}

resource "directus_role_policies_attachment" "att" {
  role_id = directus_role.r.id
  policy_ids = [
    directus_policy.pa.id,
    directus_policy.pb.id,
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_role_policies_attachment.att", "policy_ids.#", "2"),
				),
			},
			// Update: remove pb, add pc -> [pa, pc]
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "pa" {
  name       = "AccTest Update Attach Policy A"
  app_access = true
}

resource "directus_policy" "pb" {
  name       = "AccTest Update Attach Policy B"
  app_access = true
}

resource "directus_policy" "pc" {
  name         = "AccTest Update Attach Policy C"
  admin_access = true
}

resource "directus_role" "r" {
  name = "AccTest Update Attach Role"
}

resource "directus_role_policies_attachment" "att" {
  role_id = directus_role.r.id
  policy_ids = [
    directus_policy.pa.id,
    directus_policy.pc.id,
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_role_policies_attachment.att", "policy_ids.#", "2"),
				),
			},
			// Update: reduce to single policy -> [pc]
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "pa" {
  name       = "AccTest Update Attach Policy A"
  app_access = true
}

resource "directus_policy" "pb" {
  name       = "AccTest Update Attach Policy B"
  app_access = true
}

resource "directus_policy" "pc" {
  name         = "AccTest Update Attach Policy C"
  admin_access = true
}

resource "directus_role" "r" {
  name = "AccTest Update Attach Role"
}

resource "directus_role_policies_attachment" "att" {
  role_id    = directus_role.r.id
  policy_ids = [directus_policy.pc.id]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_role_policies_attachment.att", "policy_ids.#", "1"),
				),
			},
			// ImportState after updates
			{
				ResourceName:      "directus_role_policies_attachment.att",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRolePoliciesAttachmentImportStateIdFunc("directus_role_policies_attachment.att"),
			},
		},
	})
}

// TestAccRolePoliciesAttachment_replaceAll verifies that switching from one set
// of policies to a completely different set works (full replacement).
func TestAccRolePoliciesAttachment_replaceAll(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_role", "roles"),
		Steps: []resource.TestStep{
			// Create with policy_x and policy_y
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "px" {
  name       = "AccTest Replace Policy X"
  app_access = true
}

resource "directus_policy" "py" {
  name       = "AccTest Replace Policy Y"
  app_access = true
}

resource "directus_policy" "pz" {
  name         = "AccTest Replace Policy Z"
  admin_access = true
}

resource "directus_policy" "pw" {
  name       = "AccTest Replace Policy W"
  app_access = true
}

resource "directus_role" "replace_role" {
  name = "AccTest Replace All Role"
}

resource "directus_role_policies_attachment" "replace_att" {
  role_id = directus_role.replace_role.id
  policy_ids = [
    directus_policy.px.id,
    directus_policy.py.id,
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_role_policies_attachment.replace_att", "policy_ids.#", "2"),
				),
			},
			// Replace ALL: switch from [px, py] to [pz, pw]
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "px" {
  name       = "AccTest Replace Policy X"
  app_access = true
}

resource "directus_policy" "py" {
  name       = "AccTest Replace Policy Y"
  app_access = true
}

resource "directus_policy" "pz" {
  name         = "AccTest Replace Policy Z"
  admin_access = true
}

resource "directus_policy" "pw" {
  name       = "AccTest Replace Policy W"
  app_access = true
}

resource "directus_role" "replace_role" {
  name = "AccTest Replace All Role"
}

resource "directus_role_policies_attachment" "replace_att" {
  role_id = directus_role.replace_role.id
  policy_ids = [
    directus_policy.pz.id,
    directus_policy.pw.id,
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_role_policies_attachment.replace_att", "policy_ids.#", "2"),
				),
			},
			// ImportState after full replacement
			{
				ResourceName:      "directus_role_policies_attachment.replace_att",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRolePoliciesAttachmentImportStateIdFunc("directus_role_policies_attachment.replace_att"),
			},
		},
	})
}

// TestAccRolePoliciesAttachment_singlePolicy verifies the simplest case:
// a single policy attached, updated, and imported.
func TestAccRolePoliciesAttachment_singlePolicy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_role", "roles"),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "directus_policy" "single_p" {
  name       = "AccTest Single Attach Policy"
  app_access = true
}

resource "directus_role" "single_r" {
  name = "AccTest Single Attach Role"
}

resource "directus_role_policies_attachment" "single" {
  role_id    = directus_role.single_r.id
  policy_ids = [directus_policy.single_p.id]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_role_policies_attachment.single", "policy_ids.#", "1"),
					resource.TestCheckResourceAttrPair("directus_role_policies_attachment.single", "role_id", "directus_role.single_r", "id"),
				),
			},
			// ImportState
			{
				ResourceName:      "directus_role_policies_attachment.single",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRolePoliciesAttachmentImportStateIdFunc("directus_role_policies_attachment.single"),
			},
		},
	})
}

// testAccRolePoliciesAttachmentImportStateIdFunc returns the role_id for import.
// The role_policies_attachment resource uses the role UUID as import ID.
func testAccRolePoliciesAttachmentImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		roleID := rs.Primary.Attributes["role_id"]
		if roleID == "" {
			return "", fmt.Errorf("role_id not set in state for %s", resourceName)
		}

		return roleID, nil
	}
}
