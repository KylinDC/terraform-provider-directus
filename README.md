# Terraform Provider for Directus

A Terraform provider for managing [Directus](https://directus.io) resources — policies, roles, role-policy attachments, and collections.

## Features

- ✅ **Policy Management** — Create and manage access policies with granular permissions
- ✅ **Role Management** — Manage roles with hierarchical parent-child inheritance
- ✅ **Role-Policy Attachments** — Attach multiple policies to a role (authoritative, M2M via `directus_access`)
- ✅ **Collection Management** — Create and configure collections with metadata
- 🔒 **Static Token Authentication** — Secure authentication using static API tokens
- 📝 **Full CRUD Support** — Complete Create, Read, Update, Delete operations
- ✨ **Import Support** — Import existing Directus resources into Terraform state
- 🧪 **Test-Driven Development** — Built with comprehensive test coverage

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for development)
- A running Directus instance **>= v11.0** (see [Directus version compatibility](#directus-version-compatibility) below)
- Static API token with appropriate permissions

## Installation

### Using Terraform Registry

Once published, install the provider automatically via `terraform init`:

```hcl
terraform {
  required_providers {
    directus = {
      source  = "kylindc/directus"
      version = "~> 0.1"
    }
  }
}
```

See the [Publishing Guide](./PUBLISHING.md) for step-by-step instructions on publishing to the Terraform Registry.

### Local Development

1. Clone the repository:
```bash
git clone https://github.com/kylindc/terraform-provider-directus.git
cd terraform-provider-directus
```

2. Build the provider:
```bash
go build -o terraform-provider-directus
```

3. Create a local provider configuration:
```bash
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/kylindc/directus/0.1.0/darwin_arm64
cp terraform-provider-directus ~/.terraform.d/plugins/registry.terraform.io/kylindc/directus/0.1.0/darwin_arm64/
```

## Quick Start

### 1. Configure the Provider

```hcl
provider "directus" {
  endpoint = "https://your-directus-instance.com"
  token    = "your-static-token"
}
```

### 2. Create Policies

```hcl
resource "directus_policy" "editor" {
  name         = "Content Editor"
  description  = "Policy for content editors"
  icon         = "edit"
  app_access   = true
  admin_access = false
  enforce_tfa  = true
}

resource "directus_policy" "viewer" {
  name        = "Viewer"
  description = "Read-only access"
  icon        = "visibility"
  app_access  = true
}
```

### 3. Create Roles and Attach Policies

```hcl
resource "directus_role" "content_team" {
  name        = "Content Team"
  description = "Content editors and writers"
  icon        = "people"
}

resource "directus_role_policies_attachment" "content_team_policies" {
  role_id = directus_role.content_team.id
  policy_ids = [
    directus_policy.editor.id,
    directus_policy.viewer.id,
  ]
}
```

### 4. Create Collections

```hcl
resource "directus_collection" "articles" {
  collection = "articles"
  icon       = "article"
  note       = "Blog articles"
  sort_field = "sort_order"
}
```

### 5. Apply Configuration

```bash
terraform init
terraform plan
terraform apply
```

## Available Resources

### `directus_policy`

Manages Directus access policies.

```hcl
resource "directus_policy" "example" {
  name         = "Example Policy"
  description  = "An example policy"
  icon         = "security"

  # Access control
  app_access    = true   # Allow Data Studio access
  admin_access  = false  # No admin privileges
  enforce_tfa   = true   # Require 2FA

  # IP restrictions (optional)
  ip_access = "10.0.0.0/8,192.168.1.0/24"
}
```

**Arguments:**
- `name` (Required) — The name of the policy
- `description` (Optional) — Policy description
- `icon` (Optional) — Google Material Design Icon name
- `app_access` (Optional, Default: false) — Allow access to Data Studio
- `admin_access` (Optional, Default: false) — Grant full admin access
- `enforce_tfa` (Optional, Default: false) — Require two-factor authentication
- `ip_access` (Optional) — CSV of allowed IP addresses, ranges, or CIDR blocks

**Attributes:**
- `id` — The UUID of the policy (auto-generated)

---

### `directus_role`

Manages Directus roles with hierarchical parent-child inheritance.

```hcl
resource "directus_role" "editor" {
  name        = "Editor"
  description = "Content editor role"
  icon        = "edit"
  parent      = directus_role.content_team.id
}
```

**Arguments:**
- `name` (Required) — The name of the role
- `description` (Optional) — Role description
- `icon` (Optional) — Google Material Design Icon name
- `parent` (Optional) — Parent role UUID for hierarchical inheritance

**Attributes:**
- `id` — The UUID of the role (auto-generated)
- `children` (Computed) — List of child role UUIDs
- `users` (Computed) — List of user UUIDs assigned to this role

---

### `directus_role_policies_attachment`

Manages the attachment of one or more policies to a Directus role. This resource is **authoritative** — it manages ALL policy attachments for the specified role. Policies attached outside of Terraform will be detached on the next apply.

```hcl
resource "directus_role_policies_attachment" "team_policies" {
  role_id = directus_role.content_team.id
  policy_ids = [
    directus_policy.editor.id,
    directus_policy.viewer.id,
  ]
}
```

**Arguments:**
- `role_id` (Required) — The UUID of the role (forces replacement if changed)
- `policy_ids` (Required) — Set of policy UUIDs to attach to the role

**Attributes:**
- `id` (Computed) — Resource identifier (equal to `role_id`)

---

### `directus_collection`

Manages Directus collections (database tables with metadata).

```hcl
resource "directus_collection" "articles" {
  collection    = "articles"
  icon          = "article"
  note          = "Blog articles"
  hidden        = false
  singleton     = false
  sort_field    = "sort_order"
  archive_field = "status"
  color         = "#6644FF"
}
```

**Arguments:**
- `collection` (Required) — Collection name / table name (forces replacement if changed)
- `icon` (Optional) — Google Material Design Icon name
- `note` (Optional) — Short description displayed in Data Studio
- `hidden` (Optional, Default: false) — Hide the collection from Data Studio
- `singleton` (Optional, Default: false) — Treat as a singleton collection (single item)
- `sort_field` (Optional) — Field used for manual sorting
- `archive_field` (Optional) — Field used for archive/soft-delete
- `color` (Optional) — Hex color for the collection icon

**Attributes:**
- `collection` — The collection name (also serves as the resource ID)

## Import Existing Resources

Import existing Directus resources into Terraform state:

```bash
# Import a policy by UUID
terraform import directus_policy.editor 12345678-1234-1234-1234-123456789abc

# Import a role by UUID
terraform import directus_role.admin 12345678-1234-1234-1234-123456789abc

# Import role-policy attachments by role UUID
terraform import directus_role_policies_attachment.admin_policies 12345678-1234-1234-1234-123456789abc

# Import a collection by name
terraform import directus_collection.articles articles
```

## Examples

See the [examples/](./examples/) directory for scaffold-aligned usage examples:

- [Provider Configuration](./examples/provider/provider.tf)
- [Policy Resource](./examples/resources/policy/resource.tf)
- [Role Resource](./examples/resources/role/resource.tf)
- [Role-Policy Attachment Resource](./examples/resources/role_policies_attachment/resource.tf)
- [Collection Resource](./examples/resources/collection/resource.tf)

## Authentication

The provider supports static token authentication:

```hcl
provider "directus" {
  endpoint = "https://cms.example.com"
  token    = var.directus_token
}
```

### Getting a Static Token

1. Log in to your Directus instance
2. Navigate to User Directory → Your User Profile
3. Create a new static token
4. Copy the token and use it in your configuration

**Security:** Never commit tokens to version control. Use environment variables or Terraform variables:

```bash
export TF_VAR_directus_token="your-token-here"
```

## Development

### Prerequisites

- Go 1.21+
- Terraform 1.0+
- Docker & Docker Compose (for E2E tests)
- Make (optional)

### Building

```bash
go build -o terraform-provider-directus
```

### Testing

```bash
# Run all unit tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/client/...
go test ./internal/provider/...
```

### End-to-End Tests

E2E tests run against a real Directus instance via Docker Compose:

```bash
# Start Directus
docker-compose up -d

# Set up authentication token
./scripts/setup-directus.sh

# Basic E2E test (policy + role + attachment lifecycle)
./scripts/test-e2e.sh

# Comprehensive E2E test (all resources + relationships + updates)
./scripts/test-e2e-comprehensive.sh
```

### Test-Driven Development

This provider was built following TDD methodology:
1. Write tests first
2. Implement minimal code to pass tests
3. Refactor and improve

## Architecture

```
terraform-provider-directus/
├── internal/
│   ├── client/          # Directus API client
│   │   ├── client.go    # Generic CRUD + HTTP helpers
│   │   └── policy.go    # Policy-specific methods
│   ├── models/          # Reference data models
│   │   ├── collection.go
│   │   ├── policy.go
│   │   └── role.go
│   └── provider/        # Terraform resources
│       ├── provider.go
│       ├── policy_resource.go
│       ├── role_resource.go
│       ├── role_policies_attachment_resource.go
│       └── collection_resource.go
├── examples/            # HCL usage examples
├── scripts/             # E2E test and setup scripts
└── main.go              # Provider entry point
```

### Key Components

**Client** (`internal/client/`)
- HTTP client with static token authentication
- CRUD operations for Directus resources
- Proper error handling and context support
- Handles both system collections (`/roles`, `/policies`) and custom collections (`/items/{collection}`)

**Models** (`internal/models/`)
- Reference Go structs with all Directus fields
- Not used directly in resources (see API response pattern in CLAUDE.md)

**Resources** (`internal/provider/`)
- `directus_policy` — Access policy CRUD with import support
- `directus_role` — Role CRUD with O2M parent-child hierarchy
- `directus_role_policies_attachment` — Authoritative M2M role-policy management
- `directus_collection` — Collection CRUD with metadata configuration

## Directus Version Compatibility

This provider **requires Directus v11.0 or later**. It has been tested against **Directus v11.15**.

Directus v11 introduced a new access-control model where **policies** are first-class objects linked to roles (and users) via the `directus_access` junction table. This provider is built around that model:

| Feature | Minimum Directus Version |
|---|---|
| Policies API (`/policies`) | v11.0 |
| Role-Policy M2M via `directus_access` | v11.0 |
| Roles API (`/roles`) | v11.0 |
| Collections API (`/collections`) | v11.0 |

> **Directus v10.x is NOT supported.** The v10 permission model used a different structure (permissions directly on roles) that is incompatible with this provider's resources.

## Roadmap

- [x] Basic provider scaffold
- [x] Static token authentication
- [x] API client implementation
- [x] Data models (Collection, Policy, Role)
- [x] Policy resource with full CRUD
- [x] Role resource with hierarchy support
- [x] Role-policy attachments (M2M via `directus_access`)
- [x] Collection resource with metadata
- [x] E2E test infrastructure
- [x] Acceptance tests (Terraform SDK test framework)
- [x] CI/CD pipeline (GitHub Actions)
- [x] Release workflow (GoReleaser with GPG signing)
- [x] Terraform Registry documentation (`docs/`)
- [ ] Permission resource (fine-grained permissions)
- [ ] Data sources for read-only queries
- [ ] Terraform Registry publication

## Documentation

Provider documentation for the Terraform Registry is maintained in the [`docs/`](./docs/) directory:

| Document | Description |
|---|---|
| [`docs/index.md`](./docs/index.md) | Provider overview, example usage, and argument reference |
| [`docs/resources/policy.md`](./docs/resources/policy.md) | `directus_policy` resource documentation |
| [`docs/resources/role.md`](./docs/resources/role.md) | `directus_role` resource documentation |
| [`docs/resources/role_policies_attachment.md`](./docs/resources/role_policies_attachment.md) | `directus_role_policies_attachment` resource documentation |
| [`docs/resources/collection.md`](./docs/resources/collection.md) | `directus_collection` resource documentation |
| [`docs/guides/authentication.md`](./docs/guides/authentication.md) | Authentication guide |

You can preview how docs will render using the [Terraform Registry Doc Preview Tool](https://registry.terraform.io/tools/doc-preview).

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch
3. Write tests first (TDD approach)
4. Implement the feature
5. Ensure all tests pass: `go test ./...`
6. Submit a pull request

### Code Style

- Follow Go best practices
- Use `gofmt` for formatting
- Add godoc comments for exported functions
- Write clear commit messages

## Troubleshooting

### Provider fails to authenticate

**Problem:** `Unable to Create Directus Client` error

**Solutions:**
- Verify the endpoint URL is correct and accessible
- Check that your static token is valid
- Ensure the token has appropriate permissions
- Confirm Directus is running and accessible

### Resource creation fails

**Problem:** Resources fail to create with API errors

**Solutions:**
- Check Directus logs for detailed error messages
- Verify your token has create permissions
- Ensure required fields are provided
- Check for naming conflicts with existing resources

### Import fails

**Problem:** `terraform import` command fails

**Solutions:**
- Verify the resource ID/UUID is correct
- Ensure the resource exists in Directus
- Check that your token has read permissions

## Support

- 📖 [Documentation](https://docs.directus.io)
- 🐛 [Issue Tracker](https://github.com/kylindc/terraform-provider-directus/issues)
- 💬 [Discussions](https://github.com/kylindc/terraform-provider-directus/discussions)

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Directus](https://directus.io) — Open-source data platform
- [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) — Provider development framework
- The Terraform provider development community

## Related Projects

- [Directus](https://github.com/directus/directus) — The Directus project
- [Terraform](https://github.com/hashicorp/terraform) — Infrastructure as Code tool
- [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) — Provider framework

---

Made with ❤️ by the community
