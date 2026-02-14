package provider

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccProtoV6ProviderFactories returns provider factories for acceptance tests.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"directus": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates the testing prerequisites.
// It reads DIRECTUS_ENDPOINT and DIRECTUS_TOKEN from environment variables,
// falling back to .env file (TEST_DIRECTUS_ENDPOINT / TEST_DIRECTUS_TOKEN) if not set.
func testAccPreCheck(t *testing.T) {
	t.Helper()

	// Load from .env if direct env vars are not set
	loadEnvFallback()

	endpoint := os.Getenv("DIRECTUS_ENDPOINT")
	if endpoint == "" {
		t.Fatal("DIRECTUS_ENDPOINT must be set for acceptance tests (or TEST_DIRECTUS_ENDPOINT in .env)")
	}

	token := os.Getenv("DIRECTUS_TOKEN")
	if token == "" {
		t.Fatal("DIRECTUS_TOKEN must be set for acceptance tests (or TEST_DIRECTUS_TOKEN in .env)")
	}

	// Verify Directus is reachable
	req, err := http.NewRequest("GET", endpoint+"/server/ping", nil)
	if err != nil {
		t.Fatalf("Failed to create ping request: %s", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Directus is not reachable at %s: %s", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Directus returned HTTP %d for ping", resp.StatusCode)
	}
}

// loadEnvFallback reads .env file and sets DIRECTUS_ENDPOINT / DIRECTUS_TOKEN
// from TEST_DIRECTUS_ENDPOINT / TEST_DIRECTUS_TOKEN if the direct env vars are not set.
// This allows `make test-acceptance` to work seamlessly after `make setup`.
func loadEnvFallback() {
	if os.Getenv("DIRECTUS_ENDPOINT") != "" && os.Getenv("DIRECTUS_TOKEN") != "" {
		return // already set, skip .env
	}

	envMap := readDotEnv(".env")
	if len(envMap) == 0 {
		return
	}

	if os.Getenv("DIRECTUS_ENDPOINT") == "" {
		if v, ok := envMap["TEST_DIRECTUS_ENDPOINT"]; ok {
			os.Setenv("DIRECTUS_ENDPOINT", v)
		}
		if v, ok := envMap["DIRECTUS_ENDPOINT"]; ok {
			os.Setenv("DIRECTUS_ENDPOINT", v)
		}
	}

	if os.Getenv("DIRECTUS_TOKEN") == "" {
		if v, ok := envMap["TEST_DIRECTUS_TOKEN"]; ok {
			os.Setenv("DIRECTUS_TOKEN", v)
		}
		if v, ok := envMap["DIRECTUS_TOKEN"]; ok {
			os.Setenv("DIRECTUS_TOKEN", v)
		}
	}
}

// readDotEnv reads a .env file and returns a key-value map.
func readDotEnv(path string) map[string]string {
	result := make(map[string]string)

	f, err := os.Open(path)
	if err != nil {
		return result
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return result
}

// testAccProviderConfig returns the provider configuration block for tests.
func testAccProviderConfig() string {
	return fmt.Sprintf(`
provider "directus" {
  endpoint = %q
  token    = %q
}
`, os.Getenv("DIRECTUS_ENDPOINT"), os.Getenv("DIRECTUS_TOKEN"))
}

// testAccCheckResourceDestroyed is a generic CheckDestroy helper that verifies
// a Directus API resource has been deleted (returns 403 or 404).
func testAccCheckResourceDestroyed(resourceType, collection string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		endpoint := os.Getenv("DIRECTUS_ENDPOINT")
		token := os.Getenv("DIRECTUS_TOKEN")

		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}

			// Determine the ID to check
			id := rs.Primary.ID
			if id == "" {
				continue
			}

			// Build the API path
			var apiURL string
			if collection == "collections" {
				// Collections use the collection name, not a UUID
				collName := rs.Primary.Attributes["collection"]
				if collName == "" {
					collName = id
				}
				apiURL = fmt.Sprintf("%s/collections/%s", endpoint, collName)
			} else {
				apiURL = fmt.Sprintf("%s/%s/%s", endpoint, collection, id)
			}

			req, err := http.NewRequest("GET", apiURL, nil)
			if err != nil {
				return fmt.Errorf("error creating request: %s", err)
			}
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				// Connection error means it's gone
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == 404 || resp.StatusCode == 403 {
				continue
			}

			return fmt.Errorf("%s %s still exists (HTTP %d)", resourceType, id, resp.StatusCode)
		}
		return nil
	}
}
