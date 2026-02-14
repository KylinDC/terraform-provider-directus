---
page_title: "Directus Provider"
description: |-
  The Directus provider enables Terraform to manage Directus CMS resources such as policies, roles, role-policy attachments, and collections.
---

# Directus Provider

The Directus provider enables infrastructure-as-code management of [Directus](https://directus.io) resources. Use it to declaratively configure access policies, roles, role-policy attachments, and collections in your Directus instance.

This provider is built on the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) and communicates with the Directus REST API using static token authentication.

## Directus Version Compatibility

This provider **requires Directus v11.0 or later**. It has been tested against Directus v11.15.

Directus v11 introduced a new access-control model where **policies** are first-class objects linked to roles (and users) via the `directus_access` junction table. This provider is built around that model.

~> **Note** Directus v10.x is NOT supported. The v10 permission model used a different structure (permissions directly on roles) that is incompatible with this provider's resources.

## Example Usage

```hcl
terraform {
  required_providers {
    directus = {
      source  = "kylindc/directus"
      version = "~> 0.1"
    }
  }
}

provider "directus" {
  endpoint = "https://your-directus-instance.com"
  token    = var.directus_token
}

# Create a policy
resource "directus_policy" "editor" {
  name        = "Content Editor"
  description = "Policy for content editors"
  app_access  = true
}

# Create a role
resource "directus_role" "content_team" {
  name        = "Content Team"
  description = "Content editors and writers"
}

# Attach policy to role
resource "directus_role_policies_attachment" "content_team_policies" {
  role_id    = directus_role.content_team.id
  policy_ids = [directus_policy.editor.id]
}

# Create a collection
resource "directus_collection" "articles" {
  collection = "articles"
  icon       = "article"
  note       = "Blog articles"
}
```

For Terraform Registry-style standalone examples, see:

- `examples/provider/provider.tf`
- `examples/resources/policy/resource.tf`
- `examples/resources/role/resource.tf`
- `examples/resources/role_policies_attachment/resource.tf`
- `examples/resources/collection/resource.tf`

## Authentication

The provider authenticates via a **static token**. You can generate a static token in your Directus instance:

1. Log in to the Directus Data Studio.
2. Navigate to **User Directory** and select your user profile.
3. Scroll to the **Token** field and generate a new static token.
4. Copy the token and use it in the provider configuration.

-> **Note** Never commit tokens to version control. Use environment variables or Terraform input variables instead.

```bash
export TF_VAR_directus_token="your-static-token"
```

## Argument Reference

* `endpoint` - (Required) The base URL of your Directus instance (e.g., `https://cms.example.com`). Can also be set via the `DIRECTUS_ENDPOINT` environment variable.
* `token` - (Required, Sensitive) A static API token for authentication. Can also be set via the `DIRECTUS_TOKEN` environment variable.
