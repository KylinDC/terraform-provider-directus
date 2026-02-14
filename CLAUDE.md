# Directus Terraform Provider - Project Guide

This document provides context and guidelines for working on the Directus Terraform Provider project.

## Project Overview

A Terraform provider for managing Directus CMS resources, built using the Terraform Plugin Framework. The provider enables infrastructure-as-code management of Directus policies, roles, and collections.

**Current Version**: 0.1.0 (Development)
**Terraform Plugin Framework**: v1.17.0
**Go Version**: 1.21+
**Directus Version Tested**: v11.15

## Architecture

### Directory Structure

```
directus-terraform-provider/
├── internal/
│   ├── client/          # Directus API client
│   │   ├── client.go    # Generic CRUD operations
│   │   └── policy.go    # Policy-specific methods
│   ├── models/          # Data models (reference only)
│   │   ├── policy.go
│   │   ├── role.go
│   │   └── collection.go
│   └── provider/        # Terraform resources
│       ├── provider.go
│       ├── policy_resource.go
│       ├── role_resource.go
│       ├── role_policies_attachment_resource.go
│       └── collection_resource.go
├── docs/                # Terraform Registry documentation
│   ├── index.md         # Provider overview
│   ├── resources/       # Resource documentation
│   │   ├── policy.md
│   │   ├── role.md
│   │   ├── role_policies_attachment.md
│   │   └── collection.md
│   └── guides/          # Usage guides
│       └── authentication.md
├── examples/            # HCL usage examples
├── scripts/             # Testing and setup scripts
├── .github/workflows/   # CI/CD pipelines
│   ├── ci.yml           # CI: lint, test, build, acceptance
│   └── release.yml      # Release: test gate + GoReleaser
├── .goreleaser.yml      # GoReleaser configuration
├── terraform-registry-manifest.json  # Registry metadata
├── docker-compose.yml   # Local Directus for E2E tests
└── main.go              # Provider entry point
```

### Key Components

#### 1. API Client (`internal/client/`)
- **Purpose**: HTTP client for Directus API communication
- **Authentication**: Static token (Bearer token in Authorization header)
- **Coverage**: 81.1% test coverage
- **Key Functions**:
  - `NewClient()`: Initialize client with config
  - `Get/List/Create/Update/Delete()`: Generic CRUD operations
  - `GetPolicy/CreatePolicy/UpdatePolicy/DeletePolicy()`: Policy-specific methods
  - `buildCollectionPath()`: Handles system vs custom collections

#### 2. Data Models (`internal/models/`)
- **Purpose**: Reference models with all Directus fields
- **Important**: These are NOT used directly in resources due to type mismatches
- **Contains**: Complete field definitions including relationships

#### 3. Provider Resources (`internal/provider/`)

**Policy Resource** (`policy_resource.go`):
- Manages Directus access policies
- Fields: name, icon, description, ip_access, enforce_tfa, admin_access, app_access
- Uses `PolicyResourceModel` (NOT `models.Policy`) to avoid type mismatches
- Has `policyAPIResponse` for API unmarshaling

**Role Resource** (`role_resource.go`):
- Manages Directus roles with hierarchy support
- Fields: name, icon, description, parent, policies, children, users, admin_access, app_access
- Supports O2M parent-child relationships
- Uses `RoleResourceModel` (NOT `models.Role`) to avoid type mismatches
- Has `roleAPIResponse` for API unmarshaling
- M2M policy associations via `/access` endpoint (placeholder implementation)

## Critical Design Patterns

### Pattern 1: Separate API Response Models

**IMPORTANT**: Always use separate response structs for API unmarshaling.

```go
// ❌ WRONG - Will cause "cannot unmarshal string into types.StringValue" errors
var result struct {
    Data models.Policy `json:"data"`
}

// ✅ CORRECT - Use plain Go types for API responses
type policyAPIResponse struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    // ... other fields
}

var result struct {
    Data policyAPIResponse `json:"data"`
}

// Then convert to Terraform model
model := result.Data.toModel()
```

**Why**: JSON unmarshaling cannot convert strings to `types.StringValue` directly. You must unmarshal to plain Go types first, then convert to Terraform types.

### Pattern 2: Resource Model Structure

Each resource should have:

1. **ResourceModel**: Terraform schema model (e.g., `PolicyResourceModel`)
   - Uses `types.String`, `types.Bool`, `types.List`
   - Has `tfsdk` tags
   - Matches schema exactly

2. **APIResponse**: API unmarshaling model (e.g., `policyAPIResponse`)
   - Uses plain Go types: `string`, `bool`, `[]string`
   - Has `json` tags
   - Has `toModel()` method

3. **Helper Functions**:
   - `buildXxxCreateInput()`: Convert model to API request
   - `buildXxxUpdateInput()`: Convert model to API request
   - `toModel()`: Convert API response to Terraform model

### Pattern 3: Relationship Handling

**O2M (One-to-Many)**: Role hierarchy (parent → children)
- Parent field: `types.String` with parent role UUID
- Children field: `types.List` of child role UUIDs (computed)
- Handled directly via role record

**M2M (Many-to-Many)**: Role-Policy associations
- Policies field: `types.List` of policy UUIDs
- Requires `/access` endpoint with filtering
- Currently placeholder implementation

## Testing Strategy

### Unit Tests (49.0% coverage)

**Location**: `internal/client/*_test.go`, `internal/provider/*_test.go`

**Approach**:
- Mock HTTP transport using `httptest.Server`
- Test each CRUD operation independently
- Cover error paths and edge cases
- Test helper functions (buildInput, toModel)

**Run**: `make test` or `go test ./...`

### E2E Tests

**Location**: `scripts/test-e2e-comprehensive.sh`

**Approach**:
- Real Directus instance via Docker Compose
- Full lifecycle testing (create → read → update → delete)
- Relationship verification via direct API calls
- Tests O2M role hierarchies

**Run**: `./scripts/test-e2e-comprehensive.sh`

**Setup**:
1. `docker-compose up -d` - Start Directus
2. `./scripts/setup-directus.sh` - Create token, save to .env
3. Run E2E tests

## Common Tasks

### Adding a New Resource

1. **Create resource file**: `internal/provider/xxx_resource.go`
2. **Define models**:
   ```go
   type XxxResourceModel struct {
       ID   types.String `tfsdk:"id"`
       Name types.String `tfsdk:"name"`
       // ... fields matching schema
   }

   type xxxAPIResponse struct {
       ID   string `json:"id"`
       Name string `json:"name"`
       // ... fields matching API
   }

   func (r *xxxAPIResponse) toModel() *XxxResourceModel {
       // Convert plain types to Terraform types
   }
   ```
3. **Implement interfaces**:
   - `resource.Resource`
   - `resource.ResourceWithConfigure`
   - `resource.ResourceWithImportState`
4. **Implement CRUD**:
   - `Create()`, `Read()`, `Update()`, `Delete()`
   - Use helper functions for clean code
5. **Add to provider**: Update `provider.go` `Resources()` method
6. **Write tests**: Unit tests + E2E tests
7. **Add examples**: Create HCL examples in `examples/`

### Adding Client Methods

If adding resource-specific client methods (like `policy.go`):

1. Create `internal/client/xxx.go`
2. Add methods that call `c.doRequest()`
3. Use plain Go types for parameters/results
4. Add comprehensive unit tests
5. Test error paths (empty ID, nil data, etc.)

### Fixing Type Mismatch Errors

If you see: `"cannot unmarshal string into Go struct field Xxx.data.id of type basetypes.StringValue"`

**Fix**:
1. Create `xxxAPIResponse` struct with plain Go types
2. Add `toModel()` method to convert to ResourceModel
3. Update CRUD methods to use `xxxAPIResponse`
4. Update tests to match

## Development Workflow

### Setup

```bash
# Clone repo
git clone <repo-url>
cd directus-terraform-provider

# Install dependencies
go mod download

# Start Directus for testing
docker-compose up -d
./scripts/setup-directus.sh

# Build provider
make build

# Install locally
make install
```

### Testing

```bash
# Unit tests
make test

# E2E tests
make test-e2e

# Comprehensive tests (with relationships)
./scripts/test-e2e-comprehensive.sh

# Coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Making Changes

1. **Write tests first** (TDD approach)
2. **Implement feature**
3. **Run unit tests**: `make test`
4. **Run E2E tests**: `make test-e2e`
5. **Check coverage**: Aim for >70% on new code
6. **Build**: `make build`
7. **Commit**: Follow conventional commits

## API Conventions

### Directus API Endpoints

- **System collections**: `/roles`, `/policies`, `/users`, etc.
- **Custom collections**: `/items/{collection}`
- **Authentication**: `Authorization: Bearer {token}`
- **Response format**: `{"data": {...}}` for single items
- **Response format**: `{"data": [...]}` for lists

### System Collections

The client automatically routes to correct endpoint:
- `roles`, `policies`, `users`, `permissions`, `presets`, `webhooks`, etc. → `/{collection}`
- Custom collections → `/items/{collection}`

See `buildCollectionPath()` in `internal/client/client.go`

## Known Limitations

1. **Permission Resource**: Fine-grained permissions are not yet implemented
2. **Data Sources**: Read-only data sources are not yet available
3. **Field Relationships**: Collection fields support relationships but needs full implementation
4. **Environment Variable Fallback**: Provider does not yet fall back to `DIRECTUS_ENDPOINT` / `DIRECTUS_TOKEN` env vars automatically

## Troubleshooting

### "Route doesn't exist" errors
- Check Directus version (v11+ has different token endpoints)
- Use access tokens for testing (they expire in 15 minutes)
- Verify endpoint in Directus API documentation

### Type mismatch errors during tests
- Ensure using separate API response structs
- Check that `toModel()` handles all field types
- Verify test mocks return correct JSON structure

### E2E tests fail
- Check Directus is running: `docker-compose ps`
- Verify .env file exists and has valid token
- Check token hasn't expired (access tokens: 15 min)
- Review Directus logs: `docker-compose logs directus`

### Build failures
- Run `go mod tidy`
- Check Go version: `go version` (need 1.21+)
- Clear build cache: `go clean -cache`

## Resources

- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [Directus API Documentation](https://docs.directus.io/reference/introduction.html)
- [Directus Policies API](https://directus.io/docs/api/policies)
- [Go Testing](https://go.dev/doc/tutorial/add-a-test)

## Best Practices

1. **Always separate API models from Terraform models**
2. **Use table-driven tests for multiple scenarios**
3. **Test error paths, not just happy paths**
4. **Keep resource files focused (one resource per file)**
5. **Document complex logic with comments**
6. **Use meaningful variable names**
7. **Validate inputs before API calls**
8. **Handle optional fields carefully (check IsNull/IsUnknown)**
9. **Write clear, actionable error messages**
10. **Keep E2E tests comprehensive but focused**

## Code Style

- **Error handling**: Always check errors, provide context
- **Logging**: Use descriptive messages (currently minimal)
- **Naming**: CamelCase for exported, camelCase for private
- **Comments**: Godoc style for exported functions
- **Line length**: Aim for <120 characters
- **Imports**: Grouped (stdlib, external, internal)

## Git Workflow

```bash
# Create feature branch
git checkout -b feature/add-collection-resource

# Make changes, commit frequently
git add .
git commit -m "feat: add collection resource"

# Run tests before pushing
make test
./scripts/test-e2e-comprehensive.sh

# Push and create PR
git push origin feature/add-collection-resource
```

## Contributing

When adding features:
1. Check this document for patterns
2. Follow existing code structure
3. Add unit tests (aim for >70% coverage)
4. Add E2E test scenarios
5. Update examples if needed
6. Update documentation

## Version History

- **0.1.0** (Current): Initial development
  - Policy resource (CRUD + import)
  - Role resource with O2M parent-child hierarchies
  - Role-policy attachment resource (authoritative M2M via `directus_access`)
  - Collection resource with metadata
  - Terraform Registry documentation (`docs/`)
  - CI/CD pipeline (GitHub Actions: CI + Release workflows)
  - GoReleaser with optional GPG signing
  - E2E test infrastructure
  - Docker-based development environment

## Future Plans

- [ ] Permission resource (fine-grained permissions)
- [ ] Data sources (read-only resources)
- [ ] Field relationship management
- [ ] Terraform Registry publication
