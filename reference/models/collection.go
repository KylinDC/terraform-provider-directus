package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Collection represents a Directus collection (database table with metadata).
// Collections are the core data structure in Directus, similar to tables in a database.
type Collection struct {
	// Collection is the unique name of the collection. This matches the table name in the database.
	// Required: true
	Collection types.String `tfsdk:"collection" json:"collection"`

	// Meta contains Directus-specific metadata and configuration for the collection.
	// Optional: true
	Meta *CollectionMeta `tfsdk:"meta" json:"meta,omitempty"`

	// Schema contains the database schema information for the collection.
	// Optional: true (can be null for collection folders that don't have an underlying table)
	Schema *CollectionSchema `tfsdk:"schema" json:"schema,omitempty"`

	// Fields contains the fields (columns) in this collection.
	// Optional: true
	Fields []Field `tfsdk:"fields" json:"fields,omitempty"`
}

// CollectionMeta contains Directus-specific metadata for a collection.
type CollectionMeta struct {
	// Collection is the unique name of the collection.
	// Required: false (inherited from parent)
	Collection types.String `tfsdk:"collection" json:"collection,omitempty"`

	// Icon is the name of a Google Material Design Icon assigned to this collection.
	// Optional: true
	Icon types.String `tfsdk:"icon" json:"icon,omitempty"`

	// Note is a short description displayed in the Data Studio.
	// Optional: true
	Note types.String `tfsdk:"note" json:"note,omitempty"`

	// DisplayTemplate defines how items in this collection should be displayed when viewed relationally.
	// Optional: true
	DisplayTemplate types.String `tfsdk:"display_template" json:"display_template,omitempty"`

	// Hidden determines whether this collection is hidden from the Data Studio.
	// Optional: true
	// Default: false
	Hidden types.Bool `tfsdk:"hidden" json:"hidden,omitempty"`

	// Singleton indicates whether this collection should be treated as a singleton (single item).
	// Optional: true
	// Default: false
	Singleton types.Bool `tfsdk:"singleton" json:"singleton,omitempty"`

	// Translations contains translation strings for this collection.
	// Optional: true
	Translations types.List `tfsdk:"translations" json:"translations,omitempty"`

	// ArchiveField is the field used to archive items (soft delete).
	// Optional: true
	ArchiveField types.String `tfsdk:"archive_field" json:"archive_field,omitempty"`

	// ArchiveAppFilter determines whether archived items are filtered in the Data Studio.
	// Optional: true
	// Default: true
	ArchiveAppFilter types.Bool `tfsdk:"archive_app_filter" json:"archive_app_filter,omitempty"`

	// ArchiveValue is the value to set in the archive field when archiving items.
	// Optional: true
	ArchiveValue types.String `tfsdk:"archive_value" json:"archive_value,omitempty"`

	// UnarchiveValue is the value to set in the archive field when unarchiving items.
	// Optional: true
	UnarchiveValue types.String `tfsdk:"unarchive_value" json:"unarchive_value,omitempty"`

	// SortField is the field used for manual sorting of items.
	// Optional: true
	SortField types.String `tfsdk:"sort_field" json:"sort_field,omitempty"`

	// Accountability determines how accountability (tracking) is handled for this collection.
	// Possible values: "all", "activity", null
	// Optional: true
	// Default: "all"
	Accountability types.String `tfsdk:"accountability" json:"accountability,omitempty"`

	// Color is a color hex code associated with this collection.
	// Optional: true
	Color types.String `tfsdk:"color" json:"color,omitempty"`

	// ItemDuplicationFields specifies which fields should be duplicated when duplicating an item.
	// Optional: true
	ItemDuplicationFields types.List `tfsdk:"item_duplication_fields" json:"item_duplication_fields,omitempty"`

	// Sort is the default sort order for items in this collection.
	// Optional: true
	Sort types.Int64 `tfsdk:"sort" json:"sort,omitempty"`

	// Group is the parent collection for creating nested collection groups.
	// Optional: true
	Group types.String `tfsdk:"group" json:"group,omitempty"`

	// Collapse determines whether this collection group is collapsed in the Data Studio.
	// Possible values: "open", "closed", "locked"
	// Optional: true
	// Default: "open"
	Collapse types.String `tfsdk:"collapse" json:"collapse,omitempty"`

	// PreviewURL is a URL template for previewing items from this collection.
	// Optional: true
	PreviewURL types.String `tfsdk:"preview_url" json:"preview_url,omitempty"`

	// Versioning determines whether content versioning is enabled for this collection.
	// Optional: true
	// Default: false
	Versioning types.Bool `tfsdk:"versioning" json:"versioning,omitempty"`
}

// CollectionSchema represents the database schema information for a collection.
type CollectionSchema struct {
	// Name is the table name in the database.
	// Required: false (inherited from parent)
	Name types.String `tfsdk:"name" json:"name,omitempty"`

	// Comment is a database comment for the table.
	// Optional: true
	Comment types.String `tfsdk:"comment" json:"comment,omitempty"`
}

// Field represents a field (column) in a Directus collection.
type Field struct {
	// ID is the unique identifier for the field in the directus_fields collection.
	// Optional: true (computed)
	ID types.Int64 `tfsdk:"id" json:"id,omitempty"`

	// Collection is the name of the collection this field belongs to.
	// Required: true
	Collection types.String `tfsdk:"collection" json:"collection"`

	// Field is the unique name of the field within the collection.
	// Required: true
	Field types.String `tfsdk:"field" json:"field"`

	// Type is the Directus-specific data type used to cast values in the API.
	// Possible values: "string", "text", "uuid", "hash", "integer", "bigInteger",
	// "float", "decimal", "boolean", "timestamp", "datetime", "date", "time",
	// "binary", "json", "csv", "alias", and geospatial types
	// Required: true
	Type types.String `tfsdk:"type" json:"type"`

	// Meta contains Directus-specific metadata for the field.
	// Optional: true
	Meta *FieldMeta `tfsdk:"meta" json:"meta,omitempty"`

	// Schema contains database schema information for the field.
	// Optional: true
	Schema *FieldSchema `tfsdk:"schema" json:"schema,omitempty"`
}

// FieldMeta contains Directus-specific metadata for a field.
type FieldMeta struct {
	// ID is the unique identifier for the field in the directus_fields collection.
	// Optional: true (computed)
	ID types.Int64 `tfsdk:"id" json:"id,omitempty"`

	// Collection is the name of the collection this field belongs to.
	// Required: false (inherited from parent)
	Collection types.String `tfsdk:"collection" json:"collection,omitempty"`

	// Field is the unique name of the field.
	// Required: false (inherited from parent)
	Field types.String `tfsdk:"field" json:"field,omitempty"`

	// Special contains special transform flags that apply to this field.
	// Examples: "cast-boolean", "conceal", "file", "m2o", "o2m", "m2m", "m2a", "translations"
	// Optional: true
	Special types.List `tfsdk:"special" json:"special,omitempty"`

	// Interface is the interface used for this field in the Data Studio.
	// Examples: "input", "select-dropdown", "datetime", "file", "wysiwyg", etc.
	// Optional: true
	Interface types.String `tfsdk:"interface" json:"interface,omitempty"`

	// Options contains interface-specific configuration options.
	// Optional: true
	Options types.Map `tfsdk:"options" json:"options,omitempty"`

	// Display is the display template used for this field.
	// Optional: true
	Display types.String `tfsdk:"display" json:"display,omitempty"`

	// DisplayOptions contains display-specific configuration options.
	// Optional: true
	DisplayOptions types.Map `tfsdk:"display_options" json:"display_options,omitempty"`

	// ReadOnly determines whether this field is read-only in the Data Studio.
	// Optional: true
	// Default: false
	ReadOnly types.Bool `tfsdk:"readonly" json:"readonly,omitempty"`

	// Hidden determines whether this field is hidden from the Data Studio.
	// Optional: true
	// Default: false
	Hidden types.Bool `tfsdk:"hidden" json:"hidden,omitempty"`

	// Sort is the sort order for this field in the Data Studio.
	// Optional: true
	Sort types.Int64 `tfsdk:"sort" json:"sort,omitempty"`

	// Width determines the width of this field in the Data Studio.
	// Possible values: "half", "half-left", "half-right", "full", "fill"
	// Optional: true
	// Default: "full"
	Width types.String `tfsdk:"width" json:"width,omitempty"`

	// Translations contains translation strings for this field.
	// Optional: true
	Translations types.List `tfsdk:"translations" json:"translations,omitempty"`

	// Note is a helpful note that explains the field's purpose.
	// Optional: true
	Note types.String `tfsdk:"note" json:"note,omitempty"`

	// Conditions contains conditional logic for showing/hiding this field.
	// Optional: true
	Conditions types.List `tfsdk:"conditions" json:"conditions,omitempty"`

	// Required determines whether this field is required when creating items.
	// Optional: true
	// Default: false
	Required types.Bool `tfsdk:"required" json:"required,omitempty"`

	// Group is the field group this field belongs to.
	// Optional: true
	Group types.String `tfsdk:"group" json:"group,omitempty"`

	// Validation contains validation rules for this field.
	// Optional: true
	Validation types.Map `tfsdk:"validation" json:"validation,omitempty"`

	// ValidationMessage is the custom validation message to display.
	// Optional: true
	ValidationMessage types.String `tfsdk:"validation_message" json:"validation_message,omitempty"`
}

// FieldSchema contains database schema information for a field.
type FieldSchema struct {
	// Name is the column name in the database.
	// Required: false (inherited from parent)
	Name types.String `tfsdk:"name" json:"name,omitempty"`

	// Table is the table name in the database.
	// Required: false (inherited from parent)
	Table types.String `tfsdk:"table" json:"table,omitempty"`

	// DataType is the database-specific data type.
	// Optional: true
	DataType types.String `tfsdk:"data_type" json:"data_type,omitempty"`

	// DefaultValue is the default value for the field.
	// Optional: true
	DefaultValue types.String `tfsdk:"default_value" json:"default_value,omitempty"`

	// MaxLength is the maximum length for the field (for string types).
	// Optional: true
	MaxLength types.Int64 `tfsdk:"max_length" json:"max_length,omitempty"`

	// NumericPrecision is the numeric precision (for decimal types).
	// Optional: true
	NumericPrecision types.Int64 `tfsdk:"numeric_precision" json:"numeric_precision,omitempty"`

	// NumericScale is the numeric scale (for decimal types).
	// Optional: true
	NumericScale types.Int64 `tfsdk:"numeric_scale" json:"numeric_scale,omitempty"`

	// IsNullable determines whether the field can be null.
	// Optional: true
	// Default: true
	IsNullable types.Bool `tfsdk:"is_nullable" json:"is_nullable,omitempty"`

	// IsPrimaryKey determines whether this field is the primary key.
	// Optional: true
	// Default: false
	IsPrimaryKey types.Bool `tfsdk:"is_primary_key" json:"is_primary_key,omitempty"`

	// HasAutoIncrement determines whether this field auto-increments.
	// Optional: true
	// Default: false
	HasAutoIncrement types.Bool `tfsdk:"has_auto_increment" json:"has_auto_increment,omitempty"`

	// ForeignKeyColumn is the foreign key column name (for relationships).
	// Optional: true
	ForeignKeyColumn types.String `tfsdk:"foreign_key_column" json:"foreign_key_column,omitempty"`

	// ForeignKeyTable is the foreign key table name (for relationships).
	// Optional: true
	ForeignKeyTable types.String `tfsdk:"foreign_key_table" json:"foreign_key_table,omitempty"`

	// Comment is a database comment for the column.
	// Optional: true
	Comment types.String `tfsdk:"comment" json:"comment,omitempty"`

	// Schema is the database schema (PostgreSQL only).
	// Optional: true
	Schema types.String `tfsdk:"schema" json:"schema,omitempty"`

	// ForeignKeySchema is the foreign key schema (PostgreSQL only).
	// Optional: true
	ForeignKeySchema types.String `tfsdk:"foreign_key_schema" json:"foreign_key_schema,omitempty"`
}
