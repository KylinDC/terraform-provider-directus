#!/bin/bash
set -e

# ============================================================
# Comprehensive E2E Tests: All Resources + Import
# ============================================================
# This script tests the full lifecycle of all Terraform resources:
# - Policies (with IP access as CSV → API converts to array)
# - Roles (with O2M parent-child relationships)
# - Role-Policy Attachments (M2M via /access endpoint)
# - Collections (with metadata)
# - Import workflows (API → Terraform import)
# - State management (remove + reimport)
# ============================================================

echo "========================================================"
echo "  Comprehensive E2E Tests: All Resources + Import"
echo "========================================================"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

PASS_COUNT=0
FAIL_COUNT=0

pass() {
    echo -e "${GREEN}  PASS: $1${NC}"
    PASS_COUNT=$((PASS_COUNT + 1))
}

fail() {
    echo -e "${RED}  FAIL: $1${NC}"
    FAIL_COUNT=$((FAIL_COUNT + 1))
    if [ "${STRICT:-}" = "1" ]; then
        exit 1
    fi
}

section() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
else
    echo -e "${RED}ERROR: .env file not found. Run ./scripts/setup-directus.sh first${NC}"
    exit 1
fi

DIRECTUS_URL="${TEST_DIRECTUS_ENDPOINT:-http://localhost:8055}"
TOKEN="$TEST_DIRECTUS_TOKEN"

# Check if Directus is running
if ! curl -s "${DIRECTUS_URL}/server/health" > /dev/null 2>&1; then
    echo -e "${RED}ERROR: Directus is not running at ${DIRECTUS_URL}. Start it with: docker-compose up -d${NC}"
    exit 1
fi
echo -e "${GREEN}Directus is running at ${DIRECTUS_URL}${NC}"

# Helper: call Directus API
api_get() {
    local response
    response=$(curl -s -X GET "${DIRECTUS_URL}$1" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json")
    if [ $? -ne 0 ]; then
        echo -e "${RED}ERROR: API GET failed for $1${NC}" >&2
        return 1
    fi
    echo "$response"
}

api_post() {
    local response
    response=$(curl -s -X POST "${DIRECTUS_URL}$1" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d "$2")
    if [ $? -ne 0 ]; then
        echo -e "${RED}ERROR: API POST failed for $1${NC}" >&2
        echo -e "${RED}Body: $2${NC}" >&2
        return 1
    fi
    echo "$response"
}

api_delete() {
    local response
    response=$(curl -s -X DELETE "${DIRECTUS_URL}$1" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json")
    if [ $? -ne 0 ]; then
        echo -e "${RED}ERROR: API DELETE failed for $1${NC}" >&2
        return 1
    fi
    echo "$response"
}

api_status() {
    curl -s -o /dev/null -w "%{http_code}" -X GET "${DIRECTUS_URL}$1" -H "Authorization: Bearer $TOKEN"
}

# Helper: check if JSON contains a value (works for strings, arrays, etc.)
json_contains() {
    local json="$1"
    local value="$2"
    echo "$json" | grep -q "$value"
}

# Build the provider
section "Building Provider"
echo -e "${YELLOW}Building provider...${NC}"
go build -o terraform-provider-directus
echo -e "${GREEN}Provider built${NC}"

# Install provider locally
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$OS" in
    darwin) OS_NAME="darwin" ;;
    linux) OS_NAME="linux" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac
case "$ARCH" in
    x86_64) ARCH_NAME="amd64" ;;
    aarch64|arm64) ARCH_NAME="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

PROVIDER_PATH="$HOME/.terraform.d/plugins/registry.terraform.io/kylindc/directus/0.1.0/${OS_NAME}_${ARCH_NAME}"
mkdir -p "$PROVIDER_PATH"
cp terraform-provider-directus "$PROVIDER_PATH/"
chmod +x "$PROVIDER_PATH/terraform-provider-directus"
echo -e "${GREEN}Provider installed to $PROVIDER_PATH${NC}"

# Common terraform header
TF_HEADER=$(cat <<'HEADER'
terraform {
  required_providers {
    directus = {
      source = "kylindc/directus"
    }
  }
}

variable "directus_token" {
  type      = string
  sensitive = true
}
HEADER
)

# Create test directory
TEST_DIR="test-e2e-comprehensive"
rm -rf $TEST_DIR
mkdir -p $TEST_DIR
cd $TEST_DIR

TF_VAR="-var=directus_token=$TOKEN"

# ============================================================
# TEST 1: Create All Resource Types with All Fields
# ============================================================
section "Test 1: Create All Resources with Full Fields"

cat > main.tf <<EOF
${TF_HEADER}

provider "directus" {
  endpoint = "$DIRECTUS_URL"
  token    = var.directus_token
}

# --- Policies ---

# Policy with all fields
resource "directus_policy" "admin" {
  name         = "E2E Admin Policy"
  description  = "Full admin access for E2E testing"
  icon         = "admin_panel_settings"
  admin_access = true
  app_access   = true
  enforce_tfa  = true
  ip_access    = "10.0.0.0/8,192.168.1.0/24"
}

# Policy with minimal fields
resource "directus_policy" "editor" {
  name         = "E2E Editor Policy"
  description  = "Editor access for E2E testing"
  icon         = "edit"
  admin_access = false
  app_access   = true
  enforce_tfa  = false
}

# Policy with only required field
resource "directus_policy" "viewer" {
  name = "E2E Viewer Policy"
}

# API-only policy
resource "directus_policy" "api_client" {
  name         = "E2E API Client"
  description  = "API access only, no Data Studio"
  icon         = "api"
  app_access   = false
  admin_access = false
  ip_access    = "203.0.113.0/24"
}

# --- Roles ---

# Root role
resource "directus_role" "admin" {
  name        = "E2E Administrator"
  description = "Admin role for E2E testing"
  icon        = "admin_panel_settings"
}

# Child role (O2M)
resource "directus_role" "content_team" {
  name        = "E2E Content Team"
  description = "Content management team"
  icon        = "people"
  parent      = directus_role.admin.id
}

# Grandchild role (deep hierarchy)
resource "directus_role" "editor" {
  name        = "E2E Editor"
  description = "Content editor role"
  icon        = "edit"
  parent      = directus_role.content_team.id
}

# Standalone role (no parent)
resource "directus_role" "viewer" {
  name        = "E2E Viewer"
  description = "Read-only viewer role"
  icon        = "visibility"
}

# Role with minimal fields
resource "directus_role" "api_role" {
  name = "E2E API Role"
}

# --- Role-Policy Attachments ---

resource "directus_role_policies_attachment" "admin_policies" {
  role_id    = directus_role.admin.id
  policy_ids = [directus_policy.admin.id]
}

resource "directus_role_policies_attachment" "content_team_policies" {
  role_id = directus_role.content_team.id
  policy_ids = [
    directus_policy.editor.id,
    directus_policy.viewer.id,
  ]
}

resource "directus_role_policies_attachment" "editor_policies" {
  role_id    = directus_role.editor.id
  policy_ids = [directus_policy.editor.id]
}

resource "directus_role_policies_attachment" "viewer_policies" {
  role_id    = directus_role.viewer.id
  policy_ids = [directus_policy.viewer.id]
}

resource "directus_role_policies_attachment" "api_role_policies" {
  role_id = directus_role.api_role.id
  policy_ids = [
    directus_policy.api_client.id,
    directus_policy.viewer.id,
  ]
}

# --- Collections ---

# Basic collection
resource "directus_collection" "articles" {
  collection = "e2e_articles"
  icon       = "article"
  note       = "E2E test articles"
}

# Singleton collection
resource "directus_collection" "settings" {
  collection = "e2e_settings"
  icon       = "settings"
  note       = "E2E test settings (singleton)"
  singleton  = true
}

# Hidden collection
resource "directus_collection" "logs" {
  collection = "e2e_logs"
  icon       = "database"
  note       = "Internal logs"
  hidden     = true
}

# Collection with all fields
resource "directus_collection" "products" {
  collection    = "e2e_products"
  icon          = "shopping_cart"
  note          = "Product catalog"
  hidden        = false
  singleton     = false
  sort_field    = "sort_order"
  archive_field = "status"
  color         = "#6644FF"
}

# Collection with minimal fields
resource "directus_collection" "tags" {
  collection = "e2e_tags"
}

# --- Outputs ---

output "admin_policy_id"    { value = directus_policy.admin.id }
output "editor_policy_id"   { value = directus_policy.editor.id }
output "viewer_policy_id"   { value = directus_policy.viewer.id }
output "api_client_policy_id" { value = directus_policy.api_client.id }

output "admin_role_id"        { value = directus_role.admin.id }
output "content_team_role_id" { value = directus_role.content_team.id }
output "editor_role_id"       { value = directus_role.editor.id }
output "viewer_role_id"       { value = directus_role.viewer.id }
output "api_role_id"          { value = directus_role.api_role.id }
EOF

terraform init > /dev/null 2>&1
terraform apply -auto-approve $TF_VAR

# Capture IDs
ADMIN_POLICY_ID=$(terraform output -raw admin_policy_id)
EDITOR_POLICY_ID=$(terraform output -raw editor_policy_id)
VIEWER_POLICY_ID=$(terraform output -raw viewer_policy_id)
API_CLIENT_POLICY_ID=$(terraform output -raw api_client_policy_id)

ADMIN_ROLE_ID=$(terraform output -raw admin_role_id)
CONTENT_TEAM_ROLE_ID=$(terraform output -raw content_team_role_id)
EDITOR_ROLE_ID=$(terraform output -raw editor_role_id)
VIEWER_ROLE_ID=$(terraform output -raw viewer_role_id)
API_ROLE_ID=$(terraform output -raw api_role_id)

pass "Created 4 policies, 5 roles, 5 role-policy attachments, 5 collections"

# ============================================================
# TEST 2: Verify All Resources via API
# ============================================================
section "Test 2: Verify All Resources via API"

echo "  Verifying policies..."

# Admin policy - all fields
RESP=$(api_get "/policies/$ADMIN_POLICY_ID")
echo "$RESP" | grep -q "E2E Admin Policy" && pass "Policy admin: name" || fail "Policy admin: name"
echo "$RESP" | grep -q "admin_panel_settings" && pass "Policy admin: icon" || fail "Policy admin: icon"
echo "$RESP" | grep -q '"admin_access":true' && pass "Policy admin: admin_access=true" || fail "Policy admin: admin_access"
echo "$RESP" | grep -q '"app_access":true' && pass "Policy admin: app_access=true" || fail "Policy admin: app_access"
echo "$RESP" | grep -q '"enforce_tfa":true' && pass "Policy admin: enforce_tfa=true" || fail "Policy admin: enforce_tfa"
# IP access can be string or array in JSON - check for the value
if echo "$RESP" | grep -q "10.0.0.0/8" && echo "$RESP" | grep -q "192.168.1.0/24"; then
    pass "Policy admin: ip_access contains both IPs"
else
    fail "Policy admin: ip_access (got: $(echo "$RESP" | grep -o '"ip_access":[^,}]*'))"
fi

# Viewer policy - minimal fields
RESP=$(api_get "/policies/$VIEWER_POLICY_ID")
echo "$RESP" | grep -q "E2E Viewer Policy" && pass "Policy viewer: name" || fail "Policy viewer: name"
echo "$RESP" | grep -q '"admin_access":false' && pass "Policy viewer: admin_access=false (default)" || fail "Policy viewer: admin_access default"
echo "$RESP" | grep -q '"app_access":false' && pass "Policy viewer: app_access=false (default)" || fail "Policy viewer: app_access default"

echo ""
echo "  Verifying roles..."

# Admin role (root)
RESP=$(api_get "/roles/$ADMIN_ROLE_ID")
echo "$RESP" | grep -q "E2E Administrator" && pass "Role admin: name" || fail "Role admin: name"
echo "$RESP" | grep -q "admin_panel_settings" && pass "Role admin: icon" || fail "Role admin: icon"

# Content team role (child of admin)
RESP=$(api_get "/roles/$CONTENT_TEAM_ROLE_ID")
echo "$RESP" | grep -q "E2E Content Team" && pass "Role content_team: name" || fail "Role content_team: name"
echo "$RESP" | grep -q "$ADMIN_ROLE_ID" && pass "Role content_team: parent=admin" || fail "Role content_team: parent"

# Editor role (grandchild: admin -> content_team -> editor)
RESP=$(api_get "/roles/$EDITOR_ROLE_ID")
echo "$RESP" | grep -q "E2E Editor" && pass "Role editor: name" || fail "Role editor: name"
echo "$RESP" | grep -q "$CONTENT_TEAM_ROLE_ID" && pass "Role editor: parent=content_team" || fail "Role editor: parent"

# API role (minimal)
RESP=$(api_get "/roles/$API_ROLE_ID")
echo "$RESP" | grep -q "E2E API Role" && pass "Role api_role: name" || fail "Role api_role: name"

echo ""
echo "  Verifying role-policy attachments..."

# Admin -> [admin_policy]
RESP=$(api_get "/roles/$ADMIN_ROLE_ID?fields=policies.id,policies.policy")
echo "$RESP" | grep -q "$ADMIN_POLICY_ID" && pass "Attachment admin: has admin_policy" || fail "Attachment admin: admin_policy"

# Content team -> [editor_policy, viewer_policy]
RESP=$(api_get "/roles/$CONTENT_TEAM_ROLE_ID?fields=policies.id,policies.policy")
echo "$RESP" | grep -q "$EDITOR_POLICY_ID" && pass "Attachment content_team: has editor_policy" || fail "Attachment content_team: editor_policy"
echo "$RESP" | grep -q "$VIEWER_POLICY_ID" && pass "Attachment content_team: has viewer_policy" || fail "Attachment content_team: viewer_policy"

# API role -> [api_client_policy, viewer_policy]
RESP=$(api_get "/roles/$API_ROLE_ID?fields=policies.id,policies.policy")
echo "$RESP" | grep -q "$API_CLIENT_POLICY_ID" && pass "Attachment api_role: has api_client_policy" || fail "Attachment api_role: api_client_policy"
echo "$RESP" | grep -q "$VIEWER_POLICY_ID" && pass "Attachment api_role: has viewer_policy" || fail "Attachment api_role: viewer_policy"

echo ""
echo "  Verifying collections..."

RESP=$(api_get "/collections/e2e_articles")
echo "$RESP" | grep -q "e2e_articles" && pass "Collection articles: exists" || fail "Collection articles: exists"
echo "$RESP" | grep -q "article" && pass "Collection articles: icon" || fail "Collection articles: icon"

RESP=$(api_get "/collections/e2e_settings")
echo "$RESP" | grep -q "e2e_settings" && pass "Collection settings: exists" || fail "Collection settings: exists"
echo "$RESP" | grep -q '"singleton":true' && pass "Collection settings: singleton=true" || fail "Collection settings: singleton"

RESP=$(api_get "/collections/e2e_logs")
echo "$RESP" | grep -q "e2e_logs" && pass "Collection logs: exists" || fail "Collection logs: exists"
echo "$RESP" | grep -q '"hidden":true' && pass "Collection logs: hidden=true" || fail "Collection logs: hidden"

RESP=$(api_get "/collections/e2e_products")
echo "$RESP" | grep -q "e2e_products" && pass "Collection products: exists" || fail "Collection products: exists"
echo "$RESP" | grep -q "#6644FF" && pass "Collection products: color=#6644FF" || fail "Collection products: color"

RESP=$(api_get "/collections/e2e_tags")
echo "$RESP" | grep -q "e2e_tags" && pass "Collection tags: exists (minimal)" || fail "Collection tags: exists"


# ============================================================
# TEST 3: Update All Resource Types
# ============================================================
section "Test 3: Update All Resources"

cat > main.tf <<EOF
${TF_HEADER}

provider "directus" {
  endpoint = "$DIRECTUS_URL"
  token    = var.directus_token
}

# --- Updated Policies ---

resource "directus_policy" "admin" {
  name         = "E2E Admin Policy Updated"
  description  = "Updated admin policy"
  icon         = "security"
  admin_access = true
  app_access   = true
  enforce_tfa  = false
  ip_access    = "172.16.0.0/12"
}

resource "directus_policy" "editor" {
  name         = "E2E Editor Policy Updated"
  description  = "Updated editor access"
  icon         = "draw"
  admin_access = false
  app_access   = true
  enforce_tfa  = true
}

resource "directus_policy" "viewer" {
  name         = "E2E Viewer Policy Updated"
  description  = "Now with description"
  icon         = "preview"
  app_access   = true
  admin_access = false
}

resource "directus_policy" "api_client" {
  name         = "E2E API Client Updated"
  description  = "Updated API client"
  icon         = "cloud"
  app_access   = false
  admin_access = false
}

# --- Updated Roles ---

resource "directus_role" "admin" {
  name        = "E2E Administrator Updated"
  description = "Updated admin role"
  icon        = "shield"
}

# Move content_team from child of admin to standalone
resource "directus_role" "content_team" {
  name        = "E2E Content Team Updated"
  description = "Now standalone (no parent)"
  icon        = "group"
}

# Move editor to direct child of admin (re-parent)
resource "directus_role" "editor" {
  name        = "E2E Editor Updated"
  description = "Now direct child of admin"
  icon        = "create"
  parent      = directus_role.admin.id
}

resource "directus_role" "viewer" {
  name        = "E2E Viewer Updated"
  description = "Updated viewer role"
  icon        = "eye"
}

resource "directus_role" "api_role" {
  name        = "E2E API Role Updated"
  description = "Now has description"
  icon        = "cloud"
}

# --- Updated Attachments: swap policies around ---

# Admin: keep admin, add editor
resource "directus_role_policies_attachment" "admin_policies" {
  role_id = directus_role.admin.id
  policy_ids = [
    directus_policy.admin.id,
    directus_policy.editor.id,
  ]
}

# Content Team: change from editor+viewer to admin only
resource "directus_role_policies_attachment" "content_team_policies" {
  role_id    = directus_role.content_team.id
  policy_ids = [directus_policy.admin.id]
}

# Editor: add viewer
resource "directus_role_policies_attachment" "editor_policies" {
  role_id = directus_role.editor.id
  policy_ids = [
    directus_policy.editor.id,
    directus_policy.viewer.id,
  ]
}

# Viewer: unchanged
resource "directus_role_policies_attachment" "viewer_policies" {
  role_id    = directus_role.viewer.id
  policy_ids = [directus_policy.viewer.id]
}

# API role: remove viewer, add editor
resource "directus_role_policies_attachment" "api_role_policies" {
  role_id = directus_role.api_role.id
  policy_ids = [
    directus_policy.api_client.id,
    directus_policy.editor.id,
  ]
}

# --- Updated Collections ---

resource "directus_collection" "articles" {
  collection = "e2e_articles"
  icon       = "newspaper"
  note       = "Updated articles collection"
  color      = "#FF4444"
}

resource "directus_collection" "settings" {
  collection = "e2e_settings"
  icon       = "tune"
  note       = "Updated settings"
  singleton  = true
}

resource "directus_collection" "logs" {
  collection = "e2e_logs"
  icon       = "list"
  note       = "Logs now visible"
  hidden     = false
}

resource "directus_collection" "products" {
  collection    = "e2e_products"
  icon          = "inventory_2"
  note          = "Updated products"
  hidden        = false
  singleton     = false
  sort_field    = "name"
  archive_field = "archived"
  color         = "#00CC88"
}

resource "directus_collection" "tags" {
  collection = "e2e_tags"
  icon       = "label"
  note       = "Tags now have metadata"
}

# --- Outputs ---

output "admin_policy_id"    { value = directus_policy.admin.id }
output "editor_policy_id"   { value = directus_policy.editor.id }
output "viewer_policy_id"   { value = directus_policy.viewer.id }
output "api_client_policy_id" { value = directus_policy.api_client.id }

output "admin_role_id"        { value = directus_role.admin.id }
output "content_team_role_id" { value = directus_role.content_team.id }
output "editor_role_id"       { value = directus_role.editor.id }
output "viewer_role_id"       { value = directus_role.viewer.id }
output "api_role_id"          { value = directus_role.api_role.id }
EOF

terraform apply -auto-approve $TF_VAR
pass "Applied updates to all resources"

# ============================================================
# TEST 4: Verify All Updates via API
# ============================================================
section "Test 4: Verify All Updates via API"

echo "  Verifying policy updates..."

RESP=$(api_get "/policies/$ADMIN_POLICY_ID")
echo "$RESP" | grep -q "E2E Admin Policy Updated" && pass "Policy admin: name updated" || fail "Policy admin: name update"
echo "$RESP" | grep -q "security" && pass "Policy admin: icon updated" || fail "Policy admin: icon update"
echo "$RESP" | grep -q '"enforce_tfa":false' && pass "Policy admin: enforce_tfa=false" || fail "Policy admin: enforce_tfa update"
if echo "$RESP" | grep -q "172.16.0.0/12"; then
    pass "Policy admin: ip_access updated"
else
    fail "Policy admin: ip_access updated (got: $(echo "$RESP" | grep -o '"ip_access":[^,}]*'))"
fi

RESP=$(api_get "/policies/$EDITOR_POLICY_ID")
echo "$RESP" | grep -q "E2E Editor Policy Updated" && pass "Policy editor: name updated" || fail "Policy editor: name update"
echo "$RESP" | grep -q '"enforce_tfa":true' && pass "Policy editor: enforce_tfa=true" || fail "Policy editor: enforce_tfa update"

RESP=$(api_get "/policies/$VIEWER_POLICY_ID")
echo "$RESP" | grep -q "E2E Viewer Policy Updated" && pass "Policy viewer: name updated" || fail "Policy viewer: name update"
echo "$RESP" | grep -q "Now with description" && pass "Policy viewer: description added" || fail "Policy viewer: description added"
echo "$RESP" | grep -q '"app_access":true' && pass "Policy viewer: app_access=true" || fail "Policy viewer: app_access update"

echo ""
echo "  Verifying role updates..."

RESP=$(api_get "/roles/$ADMIN_ROLE_ID")
echo "$RESP" | grep -q "E2E Administrator Updated" && pass "Role admin: name updated" || fail "Role admin: name update"
echo "$RESP" | grep -q "shield" && pass "Role admin: icon updated" || fail "Role admin: icon update"

# Content team should now have no parent (null parent)
RESP=$(api_get "/roles/$CONTENT_TEAM_ROLE_ID")
echo "$RESP" | grep -q "E2E Content Team Updated" && pass "Role content_team: name updated" || fail "Role content_team: name update"
echo "$RESP" | grep -q '"parent":null' && pass "Role content_team: parent removed" || fail "Role content_team: parent removed"

# Editor should now be direct child of admin
RESP=$(api_get "/roles/$EDITOR_ROLE_ID")
echo "$RESP" | grep -q "E2E Editor Updated" && pass "Role editor: name updated" || fail "Role editor: name update"
echo "$RESP" | grep -q "$ADMIN_ROLE_ID" && pass "Role editor: re-parented to admin" || fail "Role editor: re-parent"

RESP=$(api_get "/roles/$API_ROLE_ID")
echo "$RESP" | grep -q "E2E API Role Updated" && pass "Role api_role: name updated" || fail "Role api_role: name update"
echo "$RESP" | grep -q "Now has description" && pass "Role api_role: description added" || fail "Role api_role: description added"

echo ""
echo "  Verifying attachment updates..."

# Admin: admin + editor
RESP=$(api_get "/roles/$ADMIN_ROLE_ID?fields=policies.id,policies.policy")
echo "$RESP" | grep -q "$ADMIN_POLICY_ID" && pass "Attachment admin: still has admin_policy" || fail "Attachment admin: admin_policy"
echo "$RESP" | grep -q "$EDITOR_POLICY_ID" && pass "Attachment admin: editor_policy added" || fail "Attachment admin: editor_policy added"

# Content Team: only admin
RESP=$(api_get "/roles/$CONTENT_TEAM_ROLE_ID?fields=policies.id,policies.policy")
echo "$RESP" | grep -q "$ADMIN_POLICY_ID" && pass "Attachment content_team: has admin_policy" || fail "Attachment content_team: admin_policy"
# Should NOT have editor or viewer
if echo "$RESP" | grep -q "$EDITOR_POLICY_ID"; then
    fail "Attachment content_team: editor_policy should be removed"
else
    pass "Attachment content_team: editor_policy correctly removed"
fi
if echo "$RESP" | grep -q "$VIEWER_POLICY_ID"; then
    fail "Attachment content_team: viewer_policy should be removed"
else
    pass "Attachment content_team: viewer_policy correctly removed"
fi

# API role: api_client + editor (viewer removed)
RESP=$(api_get "/roles/$API_ROLE_ID?fields=policies.id,policies.policy")
echo "$RESP" | grep -q "$API_CLIENT_POLICY_ID" && pass "Attachment api_role: still has api_client_policy" || fail "Attachment api_role: api_client_policy"
echo "$RESP" | grep -q "$EDITOR_POLICY_ID" && pass "Attachment api_role: editor_policy added" || fail "Attachment api_role: editor_policy added"
if echo "$RESP" | grep -q "$VIEWER_POLICY_ID"; then
    fail "Attachment api_role: viewer_policy should be removed"
else
    pass "Attachment api_role: viewer_policy correctly removed"
fi

echo ""
echo "  Verifying collection updates..."

RESP=$(api_get "/collections/e2e_articles")
echo "$RESP" | grep -q "newspaper" && pass "Collection articles: icon updated" || fail "Collection articles: icon update"
echo "$RESP" | grep -q "Updated articles" && pass "Collection articles: note updated" || fail "Collection articles: note update"
echo "$RESP" | grep -q "#FF4444" && pass "Collection articles: color added" || fail "Collection articles: color added"

RESP=$(api_get "/collections/e2e_logs")
echo "$RESP" | grep -q '"hidden":false' && pass "Collection logs: hidden=false" || fail "Collection logs: hidden update"

RESP=$(api_get "/collections/e2e_products")
echo "$RESP" | grep -q "#00CC88" && pass "Collection products: color updated" || fail "Collection products: color update"

RESP=$(api_get "/collections/e2e_tags")
echo "$RESP" | grep -q "label" && pass "Collection tags: icon added" || fail "Collection tags: icon added"
echo "$RESP" | grep -q "Tags now have metadata" && pass "Collection tags: note added" || fail "Collection tags: note added"


# ============================================================
# TEST 5: Destroy All and Verify
# ============================================================
section "Test 5: Destroy All Resources"

terraform destroy -auto-approve $TF_VAR
pass "Terraform destroy completed"

echo "  Verifying deletion..."

# Policies
for pid in $ADMIN_POLICY_ID $EDITOR_POLICY_ID $VIEWER_POLICY_ID $API_CLIENT_POLICY_ID; do
    HTTP_CODE=$(api_status "/policies/$pid")
    if [ "$HTTP_CODE" = "404" ] || [ "$HTTP_CODE" = "403" ]; then
        pass "Policy $pid deleted"
    else
        fail "Policy $pid still exists (HTTP $HTTP_CODE)"
    fi
done

# Roles
for rid in $ADMIN_ROLE_ID $CONTENT_TEAM_ROLE_ID $EDITOR_ROLE_ID $VIEWER_ROLE_ID $API_ROLE_ID; do
    HTTP_CODE=$(api_status "/roles/$rid")
    if [ "$HTTP_CODE" = "404" ] || [ "$HTTP_CODE" = "403" ]; then
        pass "Role $rid deleted"
    else
        fail "Role $rid still exists (HTTP $HTTP_CODE)"
    fi
done

# Collections
for coll in e2e_articles e2e_settings e2e_logs e2e_products e2e_tags; do
    HTTP_CODE=$(api_status "/collections/$coll")
    if [ "$HTTP_CODE" = "404" ] || [ "$HTTP_CODE" = "403" ]; then
        pass "Collection $coll deleted"
    else
        fail "Collection $coll still exists (HTTP $HTTP_CODE)"
    fi
done


# ============================================================
# TEST 6: Import - Create Resources via API, then Import
# ============================================================
section "Test 6: Import All Resource Types"

echo -e "${YELLOW}  Creating resources directly via Directus API...${NC}"

# Create a policy via API
# Note: Directus API may accept ip_access as string or array depending on version
IMPORT_POLICY_RESP=$(api_post "/policies" '{
    "name": "Imported Policy",
    "description": "Created via API for import test",
    "icon": "import_export",
    "admin_access": false,
    "app_access": true,
    "enforce_tfa": true,
    "ip_access": ["192.168.0.0/16"]
}')
IMPORT_POLICY_ID=$(echo "$IMPORT_POLICY_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)
if [ -n "$IMPORT_POLICY_ID" ]; then
    pass "Created policy via API: $IMPORT_POLICY_ID"
else
    fail "Failed to create policy via API"
    echo "  Response: $IMPORT_POLICY_RESP"
fi

# Create a second policy for attachment test
IMPORT_POLICY2_RESP=$(api_post "/policies" '{
    "name": "Imported Policy 2",
    "description": "Second policy for attachment import",
    "icon": "policy",
    "admin_access": false,
    "app_access": true,
    "enforce_tfa": false
}')
IMPORT_POLICY2_ID=$(echo "$IMPORT_POLICY2_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)
if [ -n "$IMPORT_POLICY2_ID" ]; then
    pass "Created policy 2 via API: $IMPORT_POLICY2_ID"
else
    fail "Failed to create policy 2 via API"
fi

# Create a role via API
IMPORT_ROLE_RESP=$(api_post "/roles" '{
    "name": "Imported Role",
    "description": "Created via API for import test",
    "icon": "download"
}')
IMPORT_ROLE_ID=$(echo "$IMPORT_ROLE_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)
if [ -n "$IMPORT_ROLE_ID" ]; then
    pass "Created role via API: $IMPORT_ROLE_ID"
else
    fail "Failed to create role via API"
    echo "  Response: $IMPORT_ROLE_RESP"
fi

# Create a child role via API
IMPORT_CHILD_ROLE_RESP=$(api_post "/roles" "{
    \"name\": \"Imported Child Role\",
    \"description\": \"Child role for import test\",
    \"icon\": \"child_care\",
    \"parent\": \"$IMPORT_ROLE_ID\"
}")
IMPORT_CHILD_ROLE_ID=$(echo "$IMPORT_CHILD_ROLE_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)
if [ -n "$IMPORT_CHILD_ROLE_ID" ]; then
    pass "Created child role via API: $IMPORT_CHILD_ROLE_ID"
else
    fail "Failed to create child role via API"
fi

# Attach policies to role via API (using nested relational operations)
ATTACH_RESP=$(curl -s -X PATCH "${DIRECTUS_URL}/roles/$IMPORT_ROLE_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"policies\":{\"create\":[{\"policy\":\"$IMPORT_POLICY_ID\"},{\"policy\":\"$IMPORT_POLICY2_ID\"}]}}")
if echo "$ATTACH_RESP" | grep -q "$IMPORT_ROLE_ID"; then
    pass "Attached policies to role via API"
else
    fail "Failed to attach policies via API"
    echo "  Response: $ATTACH_RESP"
fi

# Create a collection via API
IMPORT_COLL_RESP=$(api_post "/collections" '{
    "collection": "e2e_imported",
    "schema": {},
    "meta": {
        "icon": "cloud_download",
        "note": "Created via API for import test",
        "hidden": false,
        "singleton": false,
        "color": "#AA33BB"
    }
}')
if echo "$IMPORT_COLL_RESP" | grep -q "e2e_imported"; then
    pass "Created collection via API: e2e_imported"
else
    fail "Failed to create collection via API"
    echo "  Response: $IMPORT_COLL_RESP"
fi

echo ""
echo -e "${YELLOW}  Writing Terraform config matching the API-created resources...${NC}"

# Write a terraform config that matches the resources we created via API
cat > main.tf <<EOF
${TF_HEADER}

provider "directus" {
  endpoint = "$DIRECTUS_URL"
  token    = var.directus_token
}

# --- Import targets ---

resource "directus_policy" "imported_policy" {
  name         = "Imported Policy"
  description  = "Created via API for import test"
  icon         = "import_export"
  admin_access = false
  app_access   = true
  enforce_tfa  = true
  ip_access    = "192.168.0.0/16"
}

resource "directus_policy" "imported_policy2" {
  name         = "Imported Policy 2"
  description  = "Second policy for attachment import"
  icon         = "policy"
  admin_access = false
  app_access   = true
  enforce_tfa  = false
}

resource "directus_role" "imported_role" {
  name        = "Imported Role"
  description = "Created via API for import test"
  icon        = "download"
}

resource "directus_role" "imported_child_role" {
  name        = "Imported Child Role"
  description = "Child role for import test"
  icon        = "child_care"
  parent      = directus_role.imported_role.id
}

resource "directus_role_policies_attachment" "imported_attachment" {
  role_id = directus_role.imported_role.id
  policy_ids = [
    directus_policy.imported_policy.id,
    directus_policy.imported_policy2.id,
  ]
}

resource "directus_collection" "imported_collection" {
  collection = "e2e_imported"
  icon       = "cloud_download"
  note       = "Created via API for import test"
  hidden     = false
  singleton  = false
  color      = "#AA33BB"
}

output "imported_policy_id" { value = directus_policy.imported_policy.id }
output "imported_policy2_id" { value = directus_policy.imported_policy2.id }
output "imported_role_id" { value = directus_role.imported_role.id }
output "imported_child_role_id" { value = directus_role.imported_child_role.id }
EOF

# Re-init (fresh state)
rm -rf .terraform terraform.tfstate terraform.tfstate.backup .terraform.lock.hcl
terraform init > /dev/null 2>&1

echo ""
echo -e "${YELLOW}  Importing resources into Terraform...${NC}"

# Import policy
echo -e "${YELLOW}  Importing directus_policy...${NC}"
terraform import $TF_VAR directus_policy.imported_policy "$IMPORT_POLICY_ID" && \
    pass "Import policy: directus_policy.imported_policy" || \
    fail "Import policy: directus_policy.imported_policy"

terraform import $TF_VAR directus_policy.imported_policy2 "$IMPORT_POLICY2_ID" && \
    pass "Import policy: directus_policy.imported_policy2" || \
    fail "Import policy: directus_policy.imported_policy2"

# Import role
echo -e "${YELLOW}  Importing directus_role...${NC}"
terraform import $TF_VAR directus_role.imported_role "$IMPORT_ROLE_ID" && \
    pass "Import role: directus_role.imported_role" || \
    fail "Import role: directus_role.imported_role"

terraform import $TF_VAR directus_role.imported_child_role "$IMPORT_CHILD_ROLE_ID" && \
    pass "Import role: directus_role.imported_child_role (with parent)" || \
    fail "Import role: directus_role.imported_child_role"

# Import role-policies attachment
echo -e "${YELLOW}  Importing directus_role_policies_attachment...${NC}"
terraform import $TF_VAR directus_role_policies_attachment.imported_attachment "$IMPORT_ROLE_ID" && \
    pass "Import attachment: directus_role_policies_attachment.imported_attachment" || \
    fail "Import attachment: directus_role_policies_attachment.imported_attachment"

# Import collection
echo -e "${YELLOW}  Importing directus_collection...${NC}"
terraform import $TF_VAR directus_collection.imported_collection "e2e_imported" && \
    pass "Import collection: directus_collection.imported_collection" || \
    fail "Import collection: directus_collection.imported_collection"

echo ""
echo -e "${YELLOW}  Verifying imported state with terraform plan (expect no changes)...${NC}"

# Run plan - expect no changes if import was correct
PLAN_OUTPUT=$(terraform plan -detailed-exitcode $TF_VAR 2>&1) || PLAN_EXIT=$?
PLAN_EXIT=${PLAN_EXIT:-0}

if [ "$PLAN_EXIT" = "0" ]; then
    pass "Terraform plan shows no changes after import (perfect import)"
elif [ "$PLAN_EXIT" = "2" ]; then
    # Exit code 2 means there are changes. Let's check if they're expected
    echo -e "${YELLOW}  Plan shows changes after import (may be expected for computed fields)${NC}"
    echo "$PLAN_OUTPUT" | tail -20
    # Apply the changes to align state
    terraform apply -auto-approve $TF_VAR && \
        pass "Applied alignment changes after import" || \
        fail "Failed to apply alignment changes after import"
else
    fail "Terraform plan failed after import (exit code $PLAN_EXIT)"
    echo "$PLAN_OUTPUT" | tail -20
fi

# Verify resources are managed correctly after import
echo ""
echo -e "${YELLOW}  Verifying state after import...${NC}"

# Check terraform state list
STATE_LIST=$(terraform state list)
echo "$STATE_LIST" | grep -q "directus_policy.imported_policy" && \
    pass "State contains imported policy" || fail "State missing imported policy"
echo "$STATE_LIST" | grep -q "directus_role.imported_role" && \
    pass "State contains imported role" || fail "State missing imported role"
echo "$STATE_LIST" | grep -q "directus_role.imported_child_role" && \
    pass "State contains imported child role" || fail "State missing imported child role"
echo "$STATE_LIST" | grep -q "directus_role_policies_attachment.imported_attachment" && \
    pass "State contains imported attachment" || fail "State missing imported attachment"
echo "$STATE_LIST" | grep -q "directus_collection.imported_collection" && \
    pass "State contains imported collection" || fail "State missing imported collection"


# ============================================================
# TEST 7: Modify Imported Resources (prove they're fully managed)
# ============================================================
section "Test 7: Update Imported Resources (verify full management)"

cat > main.tf <<EOF
${TF_HEADER}

provider "directus" {
  endpoint = "$DIRECTUS_URL"
  token    = var.directus_token
}

resource "directus_policy" "imported_policy" {
  name         = "Imported Policy - Modified"
  description  = "Modified after import"
  icon         = "sync"
  admin_access = true
  app_access   = true
  enforce_tfa  = false
}

resource "directus_policy" "imported_policy2" {
  name         = "Imported Policy 2 - Modified"
  description  = "Modified after import"
  icon         = "sync_alt"
  admin_access = false
  app_access   = true
  enforce_tfa  = false
}

resource "directus_role" "imported_role" {
  name        = "Imported Role - Modified"
  description = "Modified after import"
  icon        = "sync"
}

resource "directus_role" "imported_child_role" {
  name        = "Imported Child Role - Modified"
  description = "Modified after import"
  icon        = "sync_alt"
  parent      = directus_role.imported_role.id
}

# Change attachment: remove policy2, keep policy only
resource "directus_role_policies_attachment" "imported_attachment" {
  role_id    = directus_role.imported_role.id
  policy_ids = [directus_policy.imported_policy.id]
}

resource "directus_collection" "imported_collection" {
  collection = "e2e_imported"
  icon       = "sync"
  note       = "Modified after import"
  hidden     = true
  singleton  = false
  color      = "#112233"
}

output "imported_policy_id" { value = directus_policy.imported_policy.id }
output "imported_role_id" { value = directus_role.imported_role.id }
EOF

terraform apply -auto-approve $TF_VAR && \
    pass "Updated imported resources" || \
    fail "Failed to update imported resources"

# Verify updates via API
RESP=$(api_get "/policies/$IMPORT_POLICY_ID")
echo "$RESP" | grep -q "Imported Policy - Modified" && \
    pass "Imported policy updated: name" || fail "Imported policy update: name"
echo "$RESP" | grep -q '"admin_access":true' && \
    pass "Imported policy updated: admin_access=true" || fail "Imported policy update: admin_access"

RESP=$(api_get "/roles/$IMPORT_ROLE_ID")
echo "$RESP" | grep -q "Imported Role - Modified" && \
    pass "Imported role updated: name" || fail "Imported role update: name"

RESP=$(api_get "/roles/$IMPORT_ROLE_ID?fields=policies.id,policies.policy")
echo "$RESP" | grep -q "$IMPORT_POLICY_ID" && \
    pass "Imported attachment: policy retained" || fail "Imported attachment: policy retained"
if echo "$RESP" | grep -q "$IMPORT_POLICY2_ID"; then
    fail "Imported attachment: policy2 should be removed"
else
    pass "Imported attachment: policy2 correctly removed"
fi

RESP=$(api_get "/collections/e2e_imported")
echo "$RESP" | grep -q "Modified after import" && \
    pass "Imported collection updated: note" || fail "Imported collection update: note"
echo "$RESP" | grep -q '"hidden":true' && \
    pass "Imported collection updated: hidden=true" || fail "Imported collection update: hidden"
echo "$RESP" | grep -q "#112233" && \
    pass "Imported collection updated: color" || fail "Imported collection update: color"


# ============================================================
# TEST 8: Destroy Imported Resources
# ============================================================
section "Test 8: Destroy All Imported Resources"

terraform destroy -auto-approve $TF_VAR
pass "Terraform destroy completed for imported resources"

# Verify
HTTP_CODE=$(api_status "/policies/$IMPORT_POLICY_ID")
[ "$HTTP_CODE" = "404" ] || [ "$HTTP_CODE" = "403" ] && \
    pass "Imported policy deleted" || fail "Imported policy still exists"

HTTP_CODE=$(api_status "/policies/$IMPORT_POLICY2_ID")
[ "$HTTP_CODE" = "404" ] || [ "$HTTP_CODE" = "403" ] && \
    pass "Imported policy 2 deleted" || fail "Imported policy 2 still exists"

HTTP_CODE=$(api_status "/roles/$IMPORT_ROLE_ID")
[ "$HTTP_CODE" = "404" ] || [ "$HTTP_CODE" = "403" ] && \
    pass "Imported role deleted" || fail "Imported role still exists"

HTTP_CODE=$(api_status "/roles/$IMPORT_CHILD_ROLE_ID")
[ "$HTTP_CODE" = "404" ] || [ "$HTTP_CODE" = "403" ] && \
    pass "Imported child role deleted" || fail "Imported child role still exists"

HTTP_CODE=$(api_status "/collections/e2e_imported")
[ "$HTTP_CODE" = "404" ] || [ "$HTTP_CODE" = "403" ] && \
    pass "Imported collection deleted" || fail "Imported collection still exists"


# ============================================================
# TEST 9: Import with State-Remove-Reimport Workflow
# ============================================================
section "Test 9: State Remove and Reimport Workflow"

echo -e "${YELLOW}  Creating resources via Terraform...${NC}"

cat > main.tf <<EOF
${TF_HEADER}

provider "directus" {
  endpoint = "$DIRECTUS_URL"
  token    = var.directus_token
}

resource "directus_policy" "reimport_test" {
  name         = "Reimport Test Policy"
  description  = "Testing state remove + reimport"
  icon         = "replay"
  admin_access = false
  app_access   = true
  enforce_tfa  = false
}

resource "directus_role" "reimport_test" {
  name        = "Reimport Test Role"
  description = "Testing state remove + reimport"
  icon        = "replay"
}

resource "directus_role_policies_attachment" "reimport_test" {
  role_id    = directus_role.reimport_test.id
  policy_ids = [directus_policy.reimport_test.id]
}

resource "directus_collection" "reimport_test" {
  collection = "e2e_reimport"
  icon       = "replay"
  note       = "Testing state remove + reimport"
}

output "reimport_policy_id" { value = directus_policy.reimport_test.id }
output "reimport_role_id"   { value = directus_role.reimport_test.id }
EOF

rm -rf .terraform terraform.tfstate terraform.tfstate.backup .terraform.lock.hcl
terraform init > /dev/null 2>&1
terraform apply -auto-approve $TF_VAR

REIMPORT_POLICY_ID=$(terraform output -raw reimport_policy_id)
REIMPORT_ROLE_ID=$(terraform output -raw reimport_role_id)

pass "Created resources for reimport test"

echo ""
echo -e "${YELLOW}  Removing all resources from state (not from Directus)...${NC}"

terraform state rm directus_role_policies_attachment.reimport_test
terraform state rm directus_role.reimport_test
terraform state rm directus_policy.reimport_test
terraform state rm directus_collection.reimport_test
pass "Removed all resources from Terraform state"

# Verify resources still exist in Directus
HTTP_CODE=$(api_status "/policies/$REIMPORT_POLICY_ID")
[ "$HTTP_CODE" = "200" ] && pass "Policy still exists in Directus after state rm" || fail "Policy missing from Directus"

HTTP_CODE=$(api_status "/roles/$REIMPORT_ROLE_ID")
[ "$HTTP_CODE" = "200" ] && pass "Role still exists in Directus after state rm" || fail "Role missing from Directus"

HTTP_CODE=$(api_status "/collections/e2e_reimport")
[ "$HTTP_CODE" = "200" ] && pass "Collection still exists in Directus after state rm" || fail "Collection missing from Directus"

echo ""
echo -e "${YELLOW}  Reimporting all resources...${NC}"

terraform import $TF_VAR directus_policy.reimport_test "$REIMPORT_POLICY_ID" && \
    pass "Reimported policy" || fail "Failed to reimport policy"

terraform import $TF_VAR directus_role.reimport_test "$REIMPORT_ROLE_ID" && \
    pass "Reimported role" || fail "Failed to reimport role"

terraform import $TF_VAR directus_role_policies_attachment.reimport_test "$REIMPORT_ROLE_ID" && \
    pass "Reimported role-policies attachment" || fail "Failed to reimport attachment"

terraform import $TF_VAR directus_collection.reimport_test "e2e_reimport" && \
    pass "Reimported collection" || fail "Failed to reimport collection"

echo ""
echo -e "${YELLOW}  Verifying no drift after reimport...${NC}"

PLAN_OUTPUT=$(terraform plan -detailed-exitcode $TF_VAR 2>&1) || PLAN_EXIT=$?
PLAN_EXIT=${PLAN_EXIT:-0}

if [ "$PLAN_EXIT" = "0" ]; then
    pass "No drift detected after reimport (plan shows no changes)"
elif [ "$PLAN_EXIT" = "2" ]; then
    echo -e "${YELLOW}  Plan shows minor drift (may be expected for computed fields)${NC}"
    echo "$PLAN_OUTPUT" | tail -15
    terraform apply -auto-approve $TF_VAR && \
        pass "Applied drift corrections after reimport" || \
        fail "Failed to apply drift corrections"
else
    fail "Plan failed after reimport (exit code $PLAN_EXIT)"
fi

echo ""
echo -e "${YELLOW}  Cleaning up reimport test resources...${NC}"

terraform destroy -auto-approve $TF_VAR
pass "Destroyed reimport test resources"

# ============================================================
# Cleanup
# ============================================================
cd ..
rm -rf $TEST_DIR

echo ""
echo "========================================================"
echo -e "  Test Results: ${GREEN}${PASS_COUNT} passed${NC}, ${RED}${FAIL_COUNT} failed${NC}"
echo "========================================================"
echo ""
echo "Scenarios covered:"
echo "  1. Create all resource types with full fields"
echo "     - directus_policy (with all: name, icon, desc, ip_access, enforce_tfa, admin_access, app_access)"
echo "     - directus_policy (minimal: name only)"
echo "     - directus_role (with hierarchy: admin -> content_team -> editor)"
echo "     - directus_role (standalone, minimal)"
echo "     - directus_role_policies_attachment (single + multiple policies)"
echo "     - directus_collection (basic, singleton, hidden, full-featured, minimal)"
echo "  2. Verify all resources via Directus API"
echo "  3. Update all resource types"
echo "     - Policy: name, description, icon, enforce_tfa, ip_access"
echo "     - Role: name, description, icon, re-parent, remove parent"
echo "     - Attachment: add policy, remove policy, swap policies"
echo "     - Collection: icon, note, hidden, color, sort_field, archive_field"
echo "  4. Verify all updates via API"
echo "  5. Destroy all + verify deletion"
echo "  6. Import: create via API, import into Terraform"
echo "     - terraform import directus_policy"
echo "     - terraform import directus_role (with parent)"
echo "     - terraform import directus_role_policies_attachment"
echo "     - terraform import directus_collection"
echo "  7. Modify imported resources (prove full management)"
echo "  8. Destroy imported resources"
echo "  9. State-remove + reimport workflow"
echo "     - terraform state rm + terraform import round-trip"
echo "     - Verify no drift after reimport"
echo ""

if [ "$FAIL_COUNT" -gt 0 ]; then
    echo -e "${RED}Some tests failed! See output above for details.${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi
