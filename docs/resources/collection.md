---
page_title: "directus_collection Resource - Directus"
description: |-
  Manages a Directus collection. Collections represent database tables with additional metadata and configuration.
---

# directus_collection (Resource)

Manages a Directus collection. Collections are the foundation of Directus, representing database tables with additional metadata and configuration.

This resource creates a real database table backed by a schema. Metadata such as icons, notes, sort fields, and archive fields are configured via the Directus `meta` object.

See the [Directus Collections API documentation](https://docs.directus.io/reference/system/collections.html) for more details.

## Example Usage

Registry-ready example files:

- `examples/resources/collection/resource.tf`
- `examples/resources/collection/import.sh`

### Basic Example

```hcl
resource "directus_collection" "articles" {
  collection = "articles"
  icon       = "article"
  note       = "Blog articles and posts"
}
```

## Argument Reference

The following arguments are supported:

* `collection` - (Required, Forces Replacement) The unique name of the collection. This is used as the table name in the database. Changing this value will destroy the existing collection and create a new one.
* `icon` - (Optional) The name of a [Google Material Design Icon](https://fonts.google.com/icons) assigned to this collection.
* `note` - (Optional) A short description displayed in the Data Studio.
* `hidden` - (Optional) Whether this collection is hidden from the Data Studio. Defaults to `false`.
* `singleton` - (Optional) Whether this collection should be treated as a singleton (single item). Defaults to `false`.
* `sort_field` - (Optional) The field used for manual sorting of items.
* `archive_field` - (Optional) The field used to archive items (soft delete).
* `color` - (Optional) A hex color code associated with this collection icon (e.g., `#6644FF`).

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `collection` - The collection name (also serves as the resource identifier).

## Import

Collections can be imported using the collection name:

```shell
terraform import directus_collection.example articles
```
