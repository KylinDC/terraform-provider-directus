#!/bin/bash
set -e

echo "🚀 Setting up Directus for testing..."

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Start Docker Compose
echo -e "${YELLOW}Starting Docker containers...${NC}"
docker-compose up -d

# Wait for Directus to be healthy (using HTTP health endpoint)
echo -e "${YELLOW}Waiting for Directus to be ready...${NC}"
max_attempts=60
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if curl -sf http://localhost:8055/server/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Directus is ready!${NC}"
        break
    fi
    attempt=$((attempt + 1))
    echo "Waiting... (attempt $attempt/$max_attempts)"
    sleep 2
done

if [ $attempt -eq $max_attempts ]; then
    echo -e "${RED}✗ Directus failed to start${NC}"
    docker-compose logs directus
    exit 1
fi

# Get admin token
echo -e "${YELLOW}Logging in to get admin token...${NC}"
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8055/auth/login \
    -H "Content-Type: application/json" \
    -d '{
        "email": "admin@example.com",
        "password": "admin123"
    }')

ACCESS_TOKEN=$(echo $TOKEN_RESPONSE | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

if [ -z "$ACCESS_TOKEN" ]; then
    echo -e "${RED}✗ Failed to get access token${NC}"
    echo "Response: $TOKEN_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Got access token!${NC}"

# Create a static token for testing
# In Directus v11, static tokens are set by updating the user's token field via PATCH /users/me
echo -e "${YELLOW}Creating static token...${NC}"
STATIC_TOKEN=$(openssl rand -hex 32)

PATCH_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X PATCH http://localhost:8055/users/me \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"token\": \"$STATIC_TOKEN\"}")

if [ "$PATCH_RESPONSE" != "200" ]; then
    echo -e "${RED}✗ Failed to set static token (HTTP $PATCH_RESPONSE)${NC}"
    exit 1
fi

# Verify the static token works
VERIFY_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X GET http://localhost:8055/server/ping \
    -H "Authorization: Bearer $STATIC_TOKEN")

if [ "$VERIFY_RESPONSE" != "200" ]; then
    echo -e "${RED}✗ Static token verification failed (HTTP $VERIFY_RESPONSE)${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Created and verified static token!${NC}"

# Save to .env file
cat > .env <<EOF
# Directus Configuration
DIRECTUS_ENDPOINT=http://localhost:8055
DIRECTUS_EMAIL=admin@example.com
DIRECTUS_PASSWORD=admin123

# Terraform Provider Configuration
TF_VAR_directus_endpoint=http://localhost:8055
TF_VAR_directus_token=$STATIC_TOKEN

# Test Configuration
TEST_DIRECTUS_ENDPOINT=http://localhost:8055
TEST_DIRECTUS_TOKEN=$STATIC_TOKEN
EOF

echo -e "${GREEN}✓ Configuration saved to .env${NC}"

# Print summary
echo ""
echo "=========================================="
echo -e "${GREEN}Directus Setup Complete!${NC}"
echo "=========================================="
echo ""
echo "Directus URL: http://localhost:8055"
echo "Admin Email:  admin@example.com"
echo "Admin Pass:   admin123"
echo ""
echo "Static Token: $STATIC_TOKEN"
echo ""
echo "Configuration saved to .env file"
echo ""
echo "To use with Terraform:"
echo "  export TF_VAR_directus_token=\"$STATIC_TOKEN\""
echo ""
echo "To stop Directus:"
echo "  docker-compose down"
echo ""
echo "To view logs:"
echo "  docker-compose logs -f directus"
echo ""
