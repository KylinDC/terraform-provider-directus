# Testing Guide

This document describes how to test the Directus Terraform Provider.

## Test Types

### 1. Unit Tests
Fast, isolated tests that mock external dependencies.

```bash
# Run all unit tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package
go test ./internal/client/...

# Verbose output
go test ./... -v
```

### 2. End-to-End Tests
Tests against a real Directus instance running in Docker.

```bash
# Full setup and test
make test-e2e

# Or manually:
make setup        # Setup Directus and get token
make test-e2e     # Run E2E tests
```

## Quick Start

### Using Make (Recommended)

```bash
# Setup Directus with Docker
make setup

# Run unit tests
make test

# Run end-to-end tests
make test-e2e

# Run all tests
make all
```

### Manual Setup

#### 1. Start Directus

```bash
docker-compose up -d
```

Wait 30-60 seconds for Directus to be ready. Check with:
```bash
curl http://localhost:8055/server/health
```

#### 2. Get Admin Token

```bash
# Login
curl -X POST http://localhost:8055/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "admin123"}'

# Create static token (use access_token from above)
curl -X POST http://localhost:8055/users/me/tokens \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Token", "expire": null}'
```

#### 3. Configure Environment

```bash
# Create .env file
cat > .env <<EOF
TEST_DIRECTUS_ENDPOINT=http://localhost:8055
TEST_DIRECTUS_TOKEN=<YOUR_STATIC_TOKEN>
EOF

# Export variables
export $(cat .env | xargs)
```

#### 4. Run Tests

```bash
# Build provider
go build -o terraform-provider-directus

# Install locally
make install

# Create test configuration
mkdir -p test-e2e
cd test-e2e

cat > main.tf <<EOF
terraform {
  required_providers {
    directus = {
      source = "kylindc/directus"
    }
  }
}

provider "directus" {
  endpoint = "http://localhost:8055"
  token    = "<YOUR_TOKEN>"
}

resource "directus_policy" "test" {
  name = "Test Policy"
  app_access = true
}
EOF

# Run Terraform
terraform init
terraform apply
terraform destroy
```

## Test Automation Scripts

### setup-directus.sh
Automatically sets up Directus and creates a test token.

```bash
./scripts/setup-directus.sh
```

This script:
- Starts Docker containers
- Waits for Directus to be healthy
- Logs in as admin
- Creates a static token
- Saves configuration to `.env`

### test-e2e.sh
Runs comprehensive end-to-end tests.

```bash
./scripts/test-e2e.sh
```

This script:
- Builds the provider
- Installs it locally
- Creates test resources
- Verifies CRUD operations
- Cleans up

## Docker Environment

### Configuration

The `docker-compose.yml` provides:
- **Directus**: v11.15 (latest stable)
- **PostgreSQL**: v16 Alpine
- **Port**: 8055 (Directus), 5432 (PostgreSQL)

### Default Credentials

```
Directus Admin:
  Email: admin@example.com
  Password: admin123

Database:
  User: directus
  Password: directus
  Database: directus
```

### Docker Commands

```bash
# Start
docker-compose up -d

# Stop
docker-compose down

# Stop and remove data
docker-compose down -v

# View logs
docker-compose logs -f directus

# Restart
docker-compose restart directus

# Check health
docker-compose ps
```

## Testing Workflow

### Test-Driven Development

1. **Write test first**
   ```go
   func TestPolicyResource_Create(t *testing.T) {
       // Test implementation
   }
   ```

2. **Run test (should fail)**
   ```bash
   go test ./internal/provider/... -v -run TestPolicyResource_Create
   ```

3. **Implement minimum code to pass**

4. **Run test again (should pass)**

5. **Refactor**

### Coverage Goals

- **Client package**: 80%+ coverage
- **Provider resources**: 70%+ coverage
- **Models**: 60%+ coverage (mostly data structures)

### Testing Best Practices

1. **Isolation**: Use mocks for external dependencies in unit tests
2. **Cleanup**: Always clean up resources in tests
3. **Idempotency**: Tests should be repeatable
4. **Fast**: Unit tests should be fast (<1s)
5. **Descriptive**: Use clear test names

## Troubleshooting

### Directus won't start

```bash
# Check logs
docker-compose logs directus

# Common issues:
# - Port 8055 already in use
# - Database not ready
# - Volume permissions

# Reset everything
docker-compose down -v
docker-compose up -d
```

### Tests fail with 401 Unauthorized

- Token expired or invalid
- Run `make setup` to regenerate token
- Check `.env` file exists and is loaded

### Provider not found

```bash
# Reinstall provider
make install

# Check installation
ls ~/.terraform.d/plugins/registry.terraform.io/kylindc/directus/
```

### Database connection issues

```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Restart database
docker-compose restart postgres
```

## Continuous Integration

### GitHub Actions (Future)

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Unit Tests
        run: make test
      - name: E2E Tests
        run: make test-e2e
```

## Performance Testing

### Load Testing

```bash
# Create many resources
for i in {1..100}; do
  terraform apply -auto-approve -var="policy_name=Policy_$i"
done

# Monitor Directus performance
docker stats directus-cms
```

### Benchmarking

```go
func BenchmarkPolicyCreate(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Create policy
    }
}
```

Run with:
```bash
go test -bench=. ./...
```

## Test Data

### Example Policies

```hcl
# Admin policy
resource "directus_policy" "admin" {
  name = "Admin"
  admin_access = true
  app_access = true
}

# Editor policy
resource "directus_policy" "editor" {
  name = "Editor"
  app_access = true
  ip_access = "10.0.0.0/8"
}

# Viewer policy
resource "directus_policy" "viewer" {
  name = "Viewer"
  app_access = true
  enforce_tfa = false
}
```

## Additional Resources

- [Directus API Documentation](https://docs.directus.io/reference/introduction)
- [Terraform Plugin Testing](https://developer.hashicorp.com/terraform/plugin/framework/acctests)
- [Go Testing](https://golang.org/pkg/testing/)
