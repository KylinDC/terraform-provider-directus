# Reference Models

This directory contains reference data models that document the complete structure of Directus API entities.

## Important: These Models Are NOT Used in the Provider

**These models are for reference/documentation purposes only.** They are NOT used in the actual Terraform provider implementation due to type incompatibilities between Terraform Plugin Framework types and JSON unmarshaling.

## Why These Models Exist

They provide:
1. **Complete API Documentation**: Full field definitions from Directus API
2. **Type Reference**: Show all available fields and their purposes
3. **Relationship Documentation**: Document M2M and O2M relationships

## Actual Implementation Pattern

The provider resources use a different pattern to avoid type mismatches:

### Each Resource Has Two Models:

1. **ResourceModel** (e.g., `PolicyResourceModel` in `policy_resource.go`)
   - Uses Terraform types: `types.String`, `types.Bool`, `types.List`
   - Has `tfsdk` tags matching the Terraform schema
   - This is what Terraform state uses

2. **APIResponse** (e.g., `policyAPIResponse` in `policy_resource.go`)
   - Uses plain Go types: `string`, `bool`, `[]string`
   - Has `json` tags for API unmarshaling
   - Has a `toModel()` method to convert to ResourceModel

### Why Two Models?

JSON unmarshaling **cannot directly convert** to Terraform types:
```go
// ❌ WRONG - Causes "cannot unmarshal string into types.StringValue" errors
var result struct {
    Data models.Policy `json:"data"`  // Uses types.String, types.Bool
}
json.Unmarshal(body, &result)

// ✅ CORRECT - Unmarshal to plain types first, then convert
var result struct {
    Data policyAPIResponse `json:"data"`  // Uses string, bool
}
json.Unmarshal(body, &result)
model := result.Data.toModel()  // Convert to Terraform types
```

## Reference Models in This Directory

- **policy.go**: Policy and Permission types with all fields
- **role.go**: Role types including deprecated fields
- **collection.go**: Collection, Field, and metadata types

## See Also

- `CLAUDE.md`: Project architecture and patterns
- `internal/provider/`: Actual resource implementations
- Directus API Documentation: https://docs.directus.io/reference/introduction.html
