package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// setStringField adds a string field to the input map if it's not null/unknown
func setStringField(input map[string]interface{}, key string, value types.String) {
	if !value.IsNull() && !value.IsUnknown() {
		input[key] = value.ValueString()
	}
}

// setNullableStringField adds a string field to the input map.
// Unlike setStringField, this explicitly sends null when the value is null,
// which is needed for API fields that must be cleared (e.g., role parent).
func setNullableStringField(input map[string]interface{}, key string, value types.String) {
	if value.IsUnknown() {
		return
	}
	if value.IsNull() {
		input[key] = nil
	} else {
		input[key] = value.ValueString()
	}
}

// setBoolField adds a bool field to the input map if it's not null/unknown
func setBoolField(input map[string]interface{}, key string, value types.Bool) {
	if !value.IsNull() && !value.IsUnknown() {
		input[key] = value.ValueBool()
	}
}

// stringOrNull converts a plain string to types.String, returning null for empty strings
func stringOrNull(s string) types.String {
	if s != "" {
		return types.StringValue(s)
	}
	return types.StringNull()
}

// stringListOrNull converts a string slice to types.List, returning null for empty slices
func stringListOrNull(items []string) types.List {
	if len(items) > 0 {
		elements := make([]attr.Value, len(items))
		for i, item := range items {
			elements[i] = types.StringValue(item)
		}
		listValue, _ := types.ListValue(types.StringType, elements)
		return listValue
	}
	return types.ListNull(types.StringType)
}

// buildInputMap creates a map for API requests from optional fields
// This is a common pattern for both create and update operations
func buildInputMap(fields map[string]interface{}) map[string]interface{} {
	input := make(map[string]interface{})
	for key, value := range fields {
		// Only add non-nil values
		if value != nil {
			input[key] = value
		}
	}
	return input
}
