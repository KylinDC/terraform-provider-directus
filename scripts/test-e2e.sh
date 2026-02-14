#!/bin/bash
set -e

echo "🧪 Running End-to-End Tests..."

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
else
    echo -e "${RED}✗ .env file not found. Run ./scripts/setup-directus.sh first${NC}"
    exit 1
fi

# Check if Directus is running
if ! curl -s http://localhost:8055/server/health > /dev/null; then
    echo -e "${RED}✗ Directus is not running. Start it with: docker-compose up -d${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Directus is running${NC}"

# Build the provider
echo -e "${YELLOW}Building provider...${NC}"
go build -o terraform-provider-directus
echo -e "${GREEN}✓ Provider built${NC}"

# Create test directory
TEST_DIR="test-e2e"
rm -rf $TEST_DIR
mkdir -p $TEST_DIR

# Create test configuration using Terraform variables for sensitive values
cat > $TEST_DIR/main.tf <<EOF
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

provider "directus" {
  endpoint = "$TEST_DIRECTUS_ENDPOINT"
  token    = var.directus_token
}

# Test Policy Resource
resource "directus_policy" "test_policy" {
  name         = "E2E Test Policy"
  description  = "Policy created by end-to-end test"
  icon         = "check_circle"
  app_access   = true
  admin_access = false
  enforce_tfa  = false
}

# Test Role Resource
resource "directus_role" "test_role" {
  name        = "E2E Test Role"
  description = "Role created by end-to-end test"
  icon        = "person"
}

# Test Role-Policies Attachment
resource "directus_role_policies_attachment" "test_attachment" {
  role_id    = directus_role.test_role.id
  policy_ids = [directus_policy.test_policy.id]
}

# Output for verification
output "policy_id" {
  value = directus_policy.test_policy.id
}

output "policy_name" {
  value = directus_policy.test_policy.name
}

output "role_id" {
  value = directus_role.test_role.id
}

output "role_name" {
  value = directus_role.test_role.name
}
EOF

# Setup local provider
echo -e "${YELLOW}Setting up local provider...${NC}"
PROVIDER_DIR="$HOME/.terraform.d/plugins/registry.terraform.io/kylindc/directus/0.1.0"

# Detect OS and architecture
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

PROVIDER_PATH="$PROVIDER_DIR/${OS_NAME}_${ARCH_NAME}"
mkdir -p "$PROVIDER_PATH"
cp terraform-provider-directus "$PROVIDER_PATH/"
chmod +x "$PROVIDER_PATH/terraform-provider-directus"

echo -e "${GREEN}✓ Provider installed to $PROVIDER_PATH${NC}"

# Run Terraform
cd $TEST_DIR

echo -e "${YELLOW}Initializing Terraform...${NC}"
terraform init

echo -e "${YELLOW}Planning...${NC}"
terraform plan -var="directus_token=$TEST_DIRECTUS_TOKEN"

echo -e "${YELLOW}Applying...${NC}"
terraform apply -auto-approve -var="directus_token=$TEST_DIRECTUS_TOKEN"

# Verify the policy was created
echo -e "${YELLOW}Verifying policy in Directus...${NC}"
POLICY_ID=$(terraform output -raw policy_id)

POLICY_CHECK=$(curl -s -X GET "http://localhost:8055/policies/$POLICY_ID" \
    -H "Authorization: Bearer $TEST_DIRECTUS_TOKEN")

if echo "$POLICY_CHECK" | grep -q "E2E Test Policy"; then
    echo -e "${GREEN}✓ Policy verified in Directus!${NC}"
else
    echo -e "${RED}✗ Policy not found in Directus${NC}"
    echo "  Response: $POLICY_CHECK"
    exit 1
fi

# Verify the role was created
echo -e "${YELLOW}Verifying role in Directus...${NC}"
ROLE_ID=$(terraform output -raw role_id)

ROLE_CHECK=$(curl -s -X GET "http://localhost:8055/roles/$ROLE_ID" \
    -H "Authorization: Bearer $TEST_DIRECTUS_TOKEN")

if echo "$ROLE_CHECK" | grep -q "E2E Test Role"; then
    echo -e "${GREEN}✓ Role verified in Directus!${NC}"
else
    echo -e "${RED}✗ Role not found in Directus${NC}"
    echo "  Response: $ROLE_CHECK"
    exit 1
fi

# Verify role-policy attachment
echo -e "${YELLOW}Verifying role-policy attachment...${NC}"
ATTACHMENT_CHECK=$(curl -s -X GET "http://localhost:8055/roles/$ROLE_ID?fields=policies.id,policies.policy" \
    -H "Authorization: Bearer $TEST_DIRECTUS_TOKEN")

if echo "$ATTACHMENT_CHECK" | grep -q "$POLICY_ID"; then
    echo -e "${GREEN}✓ Role-policy attachment verified!${NC}"
else
    echo -e "${RED}✗ Role-policy attachment not found${NC}"
    echo "  Response: $ATTACHMENT_CHECK"
    exit 1
fi

# Test update
echo -e "${YELLOW}Testing update...${NC}"
sed -i.bak 's/E2E Test Policy/E2E Test Policy Updated/g' main.tf
terraform apply -auto-approve -var="directus_token=$TEST_DIRECTUS_TOKEN"

# Verify update
POLICY_CHECK=$(curl -s -X GET "http://localhost:8055/policies/$POLICY_ID" \
    -H "Authorization: Bearer $TEST_DIRECTUS_TOKEN")

if echo "$POLICY_CHECK" | grep -q "E2E Test Policy Updated"; then
    echo -e "${GREEN}✓ Policy update verified!${NC}"
else
    echo -e "${RED}✗ Policy update failed${NC}"
    echo "  Response: $POLICY_CHECK"
    exit 1
fi

# Test destroy
echo -e "${YELLOW}Testing destroy...${NC}"
terraform destroy -auto-approve -var="directus_token=$TEST_DIRECTUS_TOKEN"

# Verify deletion
POLICY_CHECK=$(curl -s -o /dev/null -w "%{http_code}" \
    -X GET "http://localhost:8055/policies/$POLICY_ID" \
    -H "Authorization: Bearer $TEST_DIRECTUS_TOKEN")

if [ "$POLICY_CHECK" = "404" ] || [ "$POLICY_CHECK" = "403" ]; then
    echo -e "${GREEN}✓ Policy deletion verified!${NC}"
else
    echo -e "${RED}✗ Policy still exists after destroy (HTTP $POLICY_CHECK)${NC}"
    exit 1
fi

ROLE_CHECK=$(curl -s -o /dev/null -w "%{http_code}" \
    -X GET "http://localhost:8055/roles/$ROLE_ID" \
    -H "Authorization: Bearer $TEST_DIRECTUS_TOKEN")

if [ "$ROLE_CHECK" = "404" ] || [ "$ROLE_CHECK" = "403" ]; then
    echo -e "${GREEN}✓ Role deletion verified!${NC}"
else
    echo -e "${RED}✗ Role still exists after destroy (HTTP $ROLE_CHECK)${NC}"
    exit 1
fi

# Cleanup
cd ..
rm -rf $TEST_DIR

echo ""
echo "=========================================="
echo -e "${GREEN}All E2E Tests Passed! ✓${NC}"
echo "=========================================="
echo ""
echo "Tests completed:"
echo "  ✓ Provider initialization"
echo "  ✓ Policy creation and verification"
echo "  ✓ Role creation and verification"
echo "  ✓ Role-policy attachment and verification"
echo "  ✓ Policy update"
echo "  ✓ Resource deletion"
echo ""
