package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccFullIntegration tests all 4 resource types together in a realistic
// setup: policies -> roles (with hierarchy) -> role-policy attachments -> collections.
// This validates cross-resource references, ordering, and full lifecycle.
func TestAccFullIntegration(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create the full setup
			{
				Config: testAccProviderConfig() + testAccFullIntegrationConfig_initial(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Policies
					resource.TestCheckResourceAttr("directus_policy.admin", "name", "Integration Admin"),
					resource.TestCheckResourceAttr("directus_policy.admin", "admin_access", "true"),
					resource.TestCheckResourceAttr("directus_policy.admin", "enforce_tfa", "true"),
					resource.TestCheckResourceAttr("directus_policy.editor", "name", "Integration Editor"),
					resource.TestCheckResourceAttr("directus_policy.editor", "app_access", "true"),
					resource.TestCheckResourceAttr("directus_policy.viewer", "name", "Integration Viewer"),

					// Roles with hierarchy
					resource.TestCheckResourceAttrSet("directus_role.admin", "id"),
					resource.TestCheckResourceAttrSet("directus_role.team_lead", "id"),
					resource.TestCheckResourceAttrPair("directus_role.team_lead", "parent", "directus_role.admin", "id"),
					resource.TestCheckResourceAttrSet("directus_role.member", "id"),
					resource.TestCheckResourceAttrPair("directus_role.member", "parent", "directus_role.team_lead", "id"),

					// Attachments
					resource.TestCheckResourceAttr("directus_role_policies_attachment.admin_attach", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr("directus_role_policies_attachment.team_lead_attach", "policy_ids.#", "2"),
					resource.TestCheckResourceAttr("directus_role_policies_attachment.member_attach", "policy_ids.#", "1"),

					// Collections
					resource.TestCheckResourceAttr("directus_collection.articles", "collection", "integ_articles"),
					resource.TestCheckResourceAttr("directus_collection.articles", "icon", "article"),
					resource.TestCheckResourceAttr("directus_collection.config", "singleton", "true"),
					resource.TestCheckResourceAttr("directus_collection.audit_log", "hidden", "true"),
				),
			},
			// Step 2: Update across all resource types
			{
				Config: testAccProviderConfig() + testAccFullIntegrationConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Updated policy
					resource.TestCheckResourceAttr("directus_policy.editor", "name", "Integration Editor v2"),
					resource.TestCheckResourceAttr("directus_policy.editor", "enforce_tfa", "true"),

					// Role re-parented: member now directly under admin
					resource.TestCheckResourceAttrPair("directus_role.member", "parent", "directus_role.admin", "id"),

					// Attachment changed: team_lead loses editor, gets viewer instead
					resource.TestCheckResourceAttr("directus_role_policies_attachment.team_lead_attach", "policy_ids.#", "2"),

					// Collection updated
					resource.TestCheckResourceAttr("directus_collection.articles", "note", "Updated articles"),
					resource.TestCheckResourceAttr("directus_collection.articles", "color", "#FF0000"),
					resource.TestCheckResourceAttr("directus_collection.audit_log", "hidden", "false"),
				),
			},
			// Step 3: Import all resource types
			{
				ResourceName:      "directus_policy.admin",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:            "directus_role.admin",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"children", "users"},
			},
			{
				ResourceName:            "directus_role.team_lead",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"children", "users"},
			},
			{
				ResourceName:      "directus_role_policies_attachment.admin_attach",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRolePoliciesAttachmentImportStateIdFunc("directus_role_policies_attachment.admin_attach"),
			},
			{
				ResourceName:                         "directus_collection.articles",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccCollectionImportStateIdFunc("directus_collection.articles"),
				ImportStateVerifyIdentifierAttribute: "collection",
			},
		},
	})
}

func testAccFullIntegrationConfig_initial() string {
	return `
# --- Policies ---
resource "directus_policy" "admin" {
  name         = "Integration Admin"
  description  = "Admin policy for integration test"
  icon         = "admin_panel_settings"
  admin_access = true
  app_access   = true
  enforce_tfa  = true
}

resource "directus_policy" "editor" {
  name         = "Integration Editor"
  description  = "Editor policy"
  icon         = "edit"
  admin_access = false
  app_access   = true
  enforce_tfa  = false
}

resource "directus_policy" "viewer" {
  name         = "Integration Viewer"
  description  = "Read-only"
  icon         = "visibility"
  admin_access = false
  app_access   = true
}

# --- Roles with hierarchy ---
resource "directus_role" "admin" {
  name        = "Integration Admin Role"
  description = "Top-level admin"
  icon        = "admin_panel_settings"
}

resource "directus_role" "team_lead" {
  name        = "Integration Team Lead"
  description = "Reports to admin"
  icon        = "supervisor_account"
  parent      = directus_role.admin.id
}

resource "directus_role" "member" {
  name        = "Integration Member"
  description = "Reports to team lead"
  icon        = "person"
  parent      = directus_role.team_lead.id
}

# --- Attachments ---
resource "directus_role_policies_attachment" "admin_attach" {
  role_id    = directus_role.admin.id
  policy_ids = [directus_policy.admin.id]
}

resource "directus_role_policies_attachment" "team_lead_attach" {
  role_id = directus_role.team_lead.id
  policy_ids = [
    directus_policy.editor.id,
    directus_policy.viewer.id,
  ]
}

resource "directus_role_policies_attachment" "member_attach" {
  role_id    = directus_role.member.id
  policy_ids = [directus_policy.viewer.id]
}

# --- Collections ---
resource "directus_collection" "articles" {
  collection = "integ_articles"
  icon       = "article"
  note       = "Blog articles"
}

resource "directus_collection" "config" {
  collection = "integ_config"
  icon       = "settings"
  note       = "Site configuration"
  singleton  = true
}

resource "directus_collection" "audit_log" {
  collection = "integ_audit_log"
  icon       = "history"
  note       = "Audit trail"
  hidden     = true
}
`
}

func testAccFullIntegrationConfig_updated() string {
	return `
# --- Policies (editor updated) ---
resource "directus_policy" "admin" {
  name         = "Integration Admin"
  description  = "Admin policy for integration test"
  icon         = "admin_panel_settings"
  admin_access = true
  app_access   = true
  enforce_tfa  = true
}

resource "directus_policy" "editor" {
  name         = "Integration Editor v2"
  description  = "Updated editor policy"
  icon         = "draw"
  admin_access = false
  app_access   = true
  enforce_tfa  = true
}

resource "directus_policy" "viewer" {
  name         = "Integration Viewer"
  description  = "Read-only"
  icon         = "visibility"
  admin_access = false
  app_access   = true
}

# --- Roles (member re-parented) ---
resource "directus_role" "admin" {
  name        = "Integration Admin Role"
  description = "Top-level admin"
  icon        = "admin_panel_settings"
}

resource "directus_role" "team_lead" {
  name        = "Integration Team Lead"
  description = "Reports to admin"
  icon        = "supervisor_account"
  parent      = directus_role.admin.id
}

resource "directus_role" "member" {
  name        = "Integration Member"
  description = "Now reports to admin directly"
  icon        = "person"
  parent      = directus_role.admin.id
}

# --- Attachments (team_lead loses editor, gets admin+viewer) ---
resource "directus_role_policies_attachment" "admin_attach" {
  role_id    = directus_role.admin.id
  policy_ids = [directus_policy.admin.id]
}

resource "directus_role_policies_attachment" "team_lead_attach" {
  role_id = directus_role.team_lead.id
  policy_ids = [
    directus_policy.admin.id,
    directus_policy.viewer.id,
  ]
}

resource "directus_role_policies_attachment" "member_attach" {
  role_id    = directus_role.member.id
  policy_ids = [directus_policy.viewer.id]
}

# --- Collections (articles updated, audit_log now visible) ---
resource "directus_collection" "articles" {
  collection = "integ_articles"
  icon       = "article"
  note       = "Updated articles"
  color      = "#FF0000"
}

resource "directus_collection" "config" {
  collection = "integ_config"
  icon       = "settings"
  note       = "Site configuration"
  singleton  = true
}

resource "directus_collection" "audit_log" {
  collection = "integ_audit_log"
  icon       = "history"
  note       = "Audit trail (now visible)"
  hidden     = false
}
`
}
