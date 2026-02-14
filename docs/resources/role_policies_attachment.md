---
page_title: "directus_role_policies_attachment Resource - Directus"
description: |-
  Manages the attachment of one or more policies to a Directus role via the directus_access junction table. This resource is authoritative.
---

# directus_role_policies_attachment (Resource)

Manages the attachment of one or more policies to a Directus role.

In Directus v11+, roles and policies are linked via the `directus_access` junction table. This resource manages that link.

!> **Warning** This resource is **authoritative** — it manages ALL policy attachments for the specified role. Policies attached to the role outside of Terraform will be detached on the next `terraform apply`. If you need non-authoritative behavior, manage `directus_access` records individually.

## Example Usage

Registry-ready example files:

- `examples/resources/role_policies_attachment/resource.tf`
- `examples/resources/role_policies_attachment/import.sh`

### Basic Example

```hcl
resource "directus_role_policies_attachment" "editor_policies" {
  role_id    = directus_role.editor.id
  policy_ids = [directus_policy.content_editor.id]
}
```

## Argument Reference

The following arguments are supported:

* `role_id` - (Required, Forces Replacement) The UUID of the role to attach policies to. Changing this value will destroy the existing attachment and create a new one.
* `policy_ids` - (Required) A set of policy UUIDs to attach to the role.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource identifier, equal to `role_id`.

## Import

Role-policy attachments can be imported using the role UUID:

```shell
terraform import directus_role_policies_attachment.example 12345678-1234-1234-1234-123456789abc
```

After import, the state will be populated with all policies currently attached to the role.
