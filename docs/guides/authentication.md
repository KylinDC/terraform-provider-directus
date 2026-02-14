---
page_title: "Authentication - Directus Provider"
subcategory: ""
description: |-
  Configuring authentication for the Directus Terraform Provider using static tokens.
---

# Authentication

The Directus provider authenticates against the Directus REST API using a **static token**. This page describes how to obtain and configure the token.

## Generating a Static Token

1. Log in to the **Directus Data Studio** as an admin user.
2. Navigate to **User Directory** and select your user profile (or a service account).
3. Scroll down to the **Token** field.
4. Click **Generate** to create a new static token.
5. Copy the token immediately — it will not be shown again.

-> **Note** The user associated with the token must have sufficient permissions for the resources you plan to manage. For full provider functionality, use an admin token.

## Provider Configuration

### Inline Token (not recommended for production)

```hcl
provider "directus" {
  endpoint = "https://your-directus-instance.com"
  token    = "your-static-token-here"
}
```

### Using Terraform Variables (recommended)

```hcl
variable "directus_token" {
  type      = string
  sensitive = true
}

provider "directus" {
  endpoint = "https://your-directus-instance.com"
  token    = var.directus_token
}
```

Set the variable via environment:

```bash
export TF_VAR_directus_token="your-static-token"
terraform plan
```

Or via a `.tfvars` file (ensure it is listed in `.gitignore`):

```hcl
# terraform.tfvars
directus_token = "your-static-token"
```

### Using Environment Variables

The provider will also read configuration from environment variables if the attributes are not set in the provider block:

| Provider Attribute | Environment Variable |
|---|---|
| `endpoint` | `DIRECTUS_ENDPOINT` |
| `token` | `DIRECTUS_TOKEN` |

```bash
export DIRECTUS_ENDPOINT="https://your-directus-instance.com"
export DIRECTUS_TOKEN="your-static-token"
```

```hcl
# No attributes needed when using environment variables
provider "directus" {}
```

## Security Best Practices

!> **Warning** Never commit tokens to version control. Always use environment variables, Terraform variables, or a secrets manager.

- Use a **dedicated service account** with the minimum required permissions.
- Rotate tokens periodically.
- Restrict token scope to only the resources managed by Terraform.
- Use Terraform Cloud/Enterprise or a CI/CD secret store for production workflows.

## Troubleshooting

### "Unable to Create Directus Client" error

- Verify the `endpoint` URL is correct and accessible from your machine.
- Confirm the `token` is valid and has not expired.
- Ensure the Directus instance is running and reachable.

### 403 Forbidden errors

- The token user does not have permission for the requested operation.
- Check the user's role and policies in the Directus Data Studio.
- For admin-level operations, ensure the user has an admin policy attached.
