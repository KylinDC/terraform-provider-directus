package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCollection_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_collection", "collections"),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "directus_collection" "test" {
  collection = "acc_test_basic"
  icon       = "article"
  note       = "Basic acceptance test collection"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_collection.test", "collection", "acc_test_basic"),
					resource.TestCheckResourceAttr("directus_collection.test", "icon", "article"),
					resource.TestCheckResourceAttr("directus_collection.test", "note", "Basic acceptance test collection"),
					resource.TestCheckResourceAttr("directus_collection.test", "hidden", "false"),
					resource.TestCheckResourceAttr("directus_collection.test", "singleton", "false"),
				),
			},
			// ImportState
			{
				ResourceName:                         "directus_collection.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccCollectionImportStateIdFunc("directus_collection.test"),
				ImportStateVerifyIdentifierAttribute: "collection",
			},
		},
	})
}

func TestAccCollection_allFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_collection", "collections"),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "directus_collection" "full" {
  collection    = "acc_test_full"
  icon          = "shopping_cart"
  note          = "Full featured collection"
  hidden        = true
  singleton     = false
  sort_field    = "sort_order"
  archive_field = "status"
  color         = "#6644FF"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_collection.full", "collection", "acc_test_full"),
					resource.TestCheckResourceAttr("directus_collection.full", "icon", "shopping_cart"),
					resource.TestCheckResourceAttr("directus_collection.full", "note", "Full featured collection"),
					resource.TestCheckResourceAttr("directus_collection.full", "hidden", "true"),
					resource.TestCheckResourceAttr("directus_collection.full", "singleton", "false"),
					resource.TestCheckResourceAttr("directus_collection.full", "sort_field", "sort_order"),
					resource.TestCheckResourceAttr("directus_collection.full", "archive_field", "status"),
					resource.TestCheckResourceAttr("directus_collection.full", "color", "#6644FF"),
				),
			},
			// ImportState
			{
				ResourceName:                         "directus_collection.full",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccCollectionImportStateIdFunc("directus_collection.full"),
				ImportStateVerifyIdentifierAttribute: "collection",
			},
		},
	})
}

func TestAccCollection_singleton(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_collection", "collections"),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "directus_collection" "singleton" {
  collection = "acc_test_singleton"
  icon       = "settings"
  note       = "Singleton collection"
  singleton  = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_collection.singleton", "collection", "acc_test_singleton"),
					resource.TestCheckResourceAttr("directus_collection.singleton", "singleton", "true"),
				),
			},
			// ImportState
			{
				ResourceName:                         "directus_collection.singleton",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccCollectionImportStateIdFunc("directus_collection.singleton"),
				ImportStateVerifyIdentifierAttribute: "collection",
			},
		},
	})
}

func TestAccCollection_minimal(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_collection", "collections"),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "directus_collection" "minimal" {
  collection = "acc_test_minimal"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_collection.minimal", "collection", "acc_test_minimal"),
					resource.TestCheckResourceAttr("directus_collection.minimal", "hidden", "false"),
					resource.TestCheckResourceAttr("directus_collection.minimal", "singleton", "false"),
				),
			},
			// ImportState
			{
				ResourceName:                         "directus_collection.minimal",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccCollectionImportStateIdFunc("directus_collection.minimal"),
				ImportStateVerifyIdentifierAttribute: "collection",
			},
		},
	})
}

func TestAccCollection_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_collection", "collections"),
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccProviderConfig() + `
resource "directus_collection" "update" {
  collection = "acc_test_update"
  icon       = "article"
  note       = "Before update"
  hidden     = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_collection.update", "icon", "article"),
					resource.TestCheckResourceAttr("directus_collection.update", "note", "Before update"),
					resource.TestCheckResourceAttr("directus_collection.update", "hidden", "true"),
				),
			},
			// Update all mutable fields
			{
				Config: testAccProviderConfig() + `
resource "directus_collection" "update" {
  collection    = "acc_test_update"
  icon          = "newspaper"
  note          = "After update"
  hidden        = false
  color         = "#FF4444"
  sort_field    = "name"
  archive_field = "archived"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_collection.update", "icon", "newspaper"),
					resource.TestCheckResourceAttr("directus_collection.update", "note", "After update"),
					resource.TestCheckResourceAttr("directus_collection.update", "hidden", "false"),
					resource.TestCheckResourceAttr("directus_collection.update", "color", "#FF4444"),
					resource.TestCheckResourceAttr("directus_collection.update", "sort_field", "name"),
					resource.TestCheckResourceAttr("directus_collection.update", "archive_field", "archived"),
				),
			},
			// ImportState after update
			{
				ResourceName:                         "directus_collection.update",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccCollectionImportStateIdFunc("directus_collection.update"),
				ImportStateVerifyIdentifierAttribute: "collection",
			},
		},
	})
}

// TestAccCollection_requiresReplace verifies that changing the collection name
// forces recreation (RequiresReplace plan modifier on the "collection" attribute).
func TestAccCollection_requiresReplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_collection", "collections"),
		Steps: []resource.TestStep{
			// Create with initial name
			{
				Config: testAccProviderConfig() + `
resource "directus_collection" "replace_test" {
  collection = "acc_test_replace_v1"
  icon       = "swap_horiz"
  note       = "Will be replaced"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_collection.replace_test", "collection", "acc_test_replace_v1"),
				),
			},
			// Change collection name -> forces destroy + recreate
			{
				Config: testAccProviderConfig() + `
resource "directus_collection" "replace_test" {
  collection = "acc_test_replace_v2"
  icon       = "swap_horiz"
  note       = "Replaced collection"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_collection.replace_test", "collection", "acc_test_replace_v2"),
					resource.TestCheckResourceAttr("directus_collection.replace_test", "note", "Replaced collection"),
				),
			},
			// ImportState after replace
			{
				ResourceName:                         "directus_collection.replace_test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccCollectionImportStateIdFunc("directus_collection.replace_test"),
				ImportStateVerifyIdentifierAttribute: "collection",
			},
		},
	})
}

// TestAccCollection_hiddenSingleton verifies that a collection can be both
// hidden and singleton simultaneously, covering the combination of bool attributes.
func TestAccCollection_hiddenSingleton(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroyed("directus_collection", "collections"),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "directus_collection" "hidden_singleton" {
  collection = "acc_test_hidden_singleton"
  icon       = "lock"
  note       = "Hidden singleton config"
  hidden     = true
  singleton  = true
  color      = "#333333"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_collection.hidden_singleton", "collection", "acc_test_hidden_singleton"),
					resource.TestCheckResourceAttr("directus_collection.hidden_singleton", "hidden", "true"),
					resource.TestCheckResourceAttr("directus_collection.hidden_singleton", "singleton", "true"),
					resource.TestCheckResourceAttr("directus_collection.hidden_singleton", "color", "#333333"),
				),
			},
			// Toggle both booleans
			{
				Config: testAccProviderConfig() + `
resource "directus_collection" "hidden_singleton" {
  collection = "acc_test_hidden_singleton"
  icon       = "lock_open"
  note       = "Now visible and not singleton"
  hidden     = false
  singleton  = false
  color      = "#333333"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_collection.hidden_singleton", "hidden", "false"),
					resource.TestCheckResourceAttr("directus_collection.hidden_singleton", "singleton", "false"),
					resource.TestCheckResourceAttr("directus_collection.hidden_singleton", "icon", "lock_open"),
				),
			},
			// ImportState
			{
				ResourceName:                         "directus_collection.hidden_singleton",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccCollectionImportStateIdFunc("directus_collection.hidden_singleton"),
				ImportStateVerifyIdentifierAttribute: "collection",
			},
		},
	})
}

// testAccCollectionImportStateIdFunc returns the collection name for import.
// The collection resource uses the collection name as import ID.
func testAccCollectionImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		collName := rs.Primary.Attributes["collection"]
		if collName == "" {
			return "", fmt.Errorf("collection not set in state for %s", resourceName)
		}

		return collName, nil
	}
}
