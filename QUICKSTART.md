# Quick Start Guide

Get the Directus Terraform Provider up and running in 5 minutes!

## Prerequisites

- Docker & Docker Compose installed
- Go 1.21+ installed
- Terraform 1.0+ installed
- Make (optional, but recommended)

## Step 1: Clone and Build

```bash
cd directus-terraform-provider
go mod download
make build
```

## Step 2: Start Directus

```bash
# Start Directus with Docker
make setup
```

This will:
- Start PostgreSQL and Directus containers
- Wait for services to be healthy
- Create an admin user
- Generate a static token
- Save configuration to `.env`

**Wait 30-60 seconds** for the setup to complete.

## Step 3: Verify Setup

```bash
# Check Directus is running
curl http://localhost:8055/server/health

# Load environment variables
source .env

# View your token
echo $TF_VAR_directus_token
```

You can also access Directus UI:
- URL: http://localhost:8055
- Email: admin@example.com
- Password: admin123

## Step 4: Install Provider

```bash
make install
```

This installs the provider to your local Terraform plugins directory.

## Step 5: Create Your First Resource

Create a test directory:
```bash
mkdir terraform-test
cd terraform-test
```

Create `main.tf`:
```hcl
terraform {
  required_providers {
    directus = {
      source = "kylindc/directus"
    }
  }
}

provider "directus" {
  endpoint = "http://localhost:8055"
  token    = var.directus_token
}

variable "directus_token" {
  type      = string
  sensitive = true
}

resource "directus_policy" "my_first_policy" {
  name         = "My First Policy"
  description  = "Created with Terraform!"
  icon         = "favorite"
  app_access   = true
  admin_access = false
}

output "policy_id" {
  value = directus_policy.my_first_policy.id
}
```

## Step 6: Apply Configuration

```bash
# Initialize Terraform
terraform init

# Preview changes
terraform plan

# Apply
terraform apply -var="directus_token=$TF_VAR_directus_token"
```

## Step 7: Verify in Directus

1. Open http://localhost:8055 in your browser
2. Login with admin@example.com / admin123
3. Go to Settings → Access Control → Policies
4. You should see "My First Policy"!

## Step 8: Make Changes

Update your policy in `main.tf`:
```hcl
resource "directus_policy" "my_first_policy" {
  name         = "My Updated Policy"  # Changed!
  description  = "Updated with Terraform!"
  icon         = "star"  # Changed!
  app_access   = true
  admin_access = false
  enforce_tfa  = true  # Added!
}
```

Apply the changes:
```bash
terraform apply -var="directus_token=$TF_VAR_directus_token"
```

## Step 9: Clean Up

```bash
# Destroy resources
terraform destroy -var="directus_token=$TF_VAR_directus_token"

# Stop Directus
cd ..
make docker-down

# Or remove all data
make docker-clean
```

## Running Tests

```bash
# Run unit tests
make test

# Run end-to-end tests
make test-e2e

# Run all tests
make all
```

## Common Commands

```bash
# Start Directus
make docker-up

# Stop Directus
make docker-down

# View logs
make docker-logs

# Setup/reset token
make setup

# Build provider
make build

# Install provider
make install

# Format code
make fmt

# Run linter
make lint
```

## Troubleshooting

### "Directus is not ready"
Wait a bit longer, it takes 30-60 seconds to start. Check logs:
```bash
make docker-logs
```

### "Provider not found"
Make sure you installed it:
```bash
make install
```

### "Authentication failed"
Regenerate the token:
```bash
make setup
source .env
```

### Port already in use
Change ports in `docker-compose.yml`:
```yaml
ports:
  - "8056:8055"  # Change 8055 to 8056
```

## Next Steps

- Check out [examples/](./examples/) for more resource configurations
- Read [TESTING.md](./TESTING.md) for detailed testing guide
- See [README.md](./README.md) for full documentation
- Explore the [Directus API docs](https://docs.directus.io/reference/introduction)

## Support

Having issues? Check:
- Docker is running: `docker ps`
- Directus is healthy: `curl http://localhost:8055/server/health`
- Token is valid: Check `.env` file
- Provider is installed: `ls ~/.terraform.d/plugins/`

Still stuck? Open an issue on GitHub!
