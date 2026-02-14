# Directus Terraform Provider Examples

This directory contains example Terraform configurations for using the Directus provider.

## Prerequisites

1. A running Directus instance
2. A static API token with appropriate permissions
3. Terraform v1.0+

## Getting Started

### 1. Configure the Provider

```hcl
provider "directus" {
  endpoint = "https://your-directus-instance.com"
  token    = "your-static-token"
}
```

You can also use environment variables:
```bash
export DIRECTUS_ENDPOINT="https://your-directus-instance.com"
export DIRECTUS_TOKEN="your-static-token"
```

### 2. Run the Examples

```bash
# Initialize Terraform
terraform init

# Preview changes
terraform plan

# Apply configuration
terraform apply
```

## Available Resources

### `directus_policy` ✅ Implemented

Manages Directus access policies.

**Example:**
```hcl
resource "directus_policy" "editor" {
  name         = "Content Editor"
  description  = "Policy for content editors"
  icon         = "edit"
  app_access   = true
  admin_access = false
  enforce_tfa  = true
  ip_access    = "10.0.0.0/8"
}
```

**Arguments:**
- `name` (Required) - Policy name
- `description` (Optional) - Policy description
- `icon` (Optional) - Google Material Design Icon name
- `app_access` (Optional) - Allow Data Studio access (default: false)
- `admin_access` (Optional) - Grant full admin access (default: false)
- `enforce_tfa` (Optional) - Require two-factor authentication (default: false)
- `ip_access` (Optional) - CSV of allowed IP addresses/ranges/CIDR blocks

**Attributes:**
- `id` - Policy UUID (auto-generated)

### `directus_role` ✅ Implemented

Manages Directus roles with hierarchical support.

**Example:**
```hcl
resource "directus_role" "editor" {
  name        = "Editor"
  description = "Content editor role"
  icon        = "edit"
  parent      = directus_role.parent_role.id
}
```

**Arguments:**
- `name` (Required) - Role name
- `description` (Optional) - Role description
- `icon` (Optional) - Google Material Design Icon name
- `parent` (Optional) - Parent role UUID for inheritance

**Attributes:**
- `id` - Role UUID (auto-generated)
- `children` - List of child role UUIDs (computed)

### `directus_role_policies_attachment` ✅ Implemented

Manages the attachment of one or more policies to a Directus role. This resource is **authoritative**: it manages ALL policy attachments for the specified role.

**Example:**
```hcl
resource "directus_role_policies_attachment" "editor_policies" {
  role_id = directus_role.editor.id
  policy_ids = [
    directus_policy.content_editor.id,
    directus_policy.viewer.id,
  ]
}
```

**Arguments:**
- `role_id` (Required) - The UUID of the role (forces replacement if changed)
- `policy_ids` (Required) - Set of policy UUIDs to attach to the role

**Attributes:**
- `id` - Resource identifier (equal to role_id)

**Import:**
```bash
terraform import directus_role_policies_attachment.example <role_id>
```

### `directus_collection` ✅ Implemented

Manages Directus collections (database tables).

**Example:**
```hcl
resource "directus_collection" "articles" {
  collection    = "articles"
  icon          = "article"
  note          = "Blog articles"
  sort_field    = "sort_order"
  archive_field = "status"
  color         = "#6644FF"
}
```

**Arguments:**
- `collection` (Required) - Collection name (forces replacement if changed)
- `icon` (Optional) - Google Material Design Icon name
- `note` (Optional) - Description/note for the collection
- `hidden` (Optional) - Hide the collection in Data Studio (default: false)
- `singleton` (Optional) - Treat as singleton collection (default: false)
- `sort_field` (Optional) - Field used for manual sorting
- `archive_field` (Optional) - Field used for archive/soft-delete
- `color` (Optional) - Hex color for the collection icon

**Attributes:**
- `id` - Collection name (same as collection argument)

## Examples

### Terraform Registry / Scaffolding-style Structure

These examples follow the HashiCorp scaffolding-framework convention and are suitable for Terraform Registry docs linkage:

- Provider config: [provider/provider.tf](./provider/provider.tf)
- Policy resource: [resources/policy/resource.tf](./resources/policy/resource.tf) | [resources/policy/import.sh](./resources/policy/import.sh)
- Role resource: [resources/role/resource.tf](./resources/role/resource.tf) | [resources/role/import.sh](./resources/role/import.sh)
- Role-policy attachment resource: [resources/role_policies_attachment/resource.tf](./resources/role_policies_attachment/resource.tf) | [resources/role_policies_attachment/import.sh](./resources/role_policies_attachment/import.sh)
- Collection resource: [resources/collection/resource.tf](./resources/collection/resource.tf) | [resources/collection/import.sh](./resources/collection/import.sh)

These are the canonical examples used to keep the repository aligned with the Terraform provider scaffolding conventions.

## Authentication

The provider supports static token authentication only:

```hcl
provider "directus" {
  endpoint = "https://cms.example.com"
  token    = var.directus_token  # Store securely in variables
}
```

### Getting a Static Token

1. Log in to your Directus instance
2. Go to User Directory → Your User
3. Create a new static token
4. Copy the token and use it in your provider configuration

**Security Note:** Never commit tokens to version control. Use Terraform variables or environment variables.

## Import Existing Resources

You can import existing Directus resources:

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

## Tips and Best Practices

1. **Use Variables for Sensitive Data**
   ```hcl
   variable "directus_token" {
     type      = string
     sensitive = true
   }
   ```

2. **Organize by Environment**
   ```
   examples/
   ├── dev/
   ├── staging/
   └── production/
   ```

3. **Use Modules for Reusability**
   ```hcl
   module "content_team_policy" {
     source      = "./modules/policy"
     name        = "Content Team"
     app_access  = true
     enforce_tfa = true
   }
   ```

4. **Tag Resources with Metadata**
   Use descriptions to document your infrastructure

5. **Test in Development First**
   Always test configuration changes in a development environment

## Troubleshooting

### Authentication Errors
- Verify your token is valid and has necessary permissions
- Check that the endpoint URL is correct and accessible

### Resource Not Found
- Ensure the resource exists in Directus
- Check that you're using the correct ID/name

### Permission Denied
- Verify your token has appropriate permissions for the operation
- Check Directus policy configurations

## Additional Resources

- [Directus Documentation](https://docs.directus.io)
- [Directus API Reference](https://docs.directus.io/reference/introduction)
- [Terraform Provider Development](https://developer.hashicorp.com/terraform/plugin/framework)

## Contributing

Found an issue or want to contribute? Please visit the [GitHub repository](https://github.com/kylindc/terraform-provider-directus).
