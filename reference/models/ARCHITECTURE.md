# Data Models Architecture

## Overview

This document describes the architecture and design decisions for the Directus Terraform Provider data models.

## Design Principles

1. **API Fidelity**: Models match the Directus API structure exactly to ensure compatibility
2. **Terraform Native**: Use Terraform Plugin Framework types throughout for proper three-state logic
3. **Comprehensive Documentation**: Every field is documented with purpose, requirements, and constraints
4. **Type Safety**: Leverage Go's type system to prevent errors at compile time
5. **Extensibility**: Models can be easily extended as new Directus features are added

## Model Hierarchy

```
models/
├── collection.go       # Collection and Field models
├── policy.go          # Policy and Permission models
├── role.go            # Role models with hierarchy
├── README.md          # Usage documentation
└── ARCHITECTURE.md    # This file
```

## Data Flow

```
Terraform Config → Resource Schema → Model → JSON → Directus API
                ←                  ←       ←        ←
```

1. **Terraform Config**: User defines resources in HCL
2. **Resource Schema**: Terraform Plugin Framework validates and parses
3. **Model**: Data is mapped to Go structs with proper types
4. **JSON**: Models are serialized to JSON for API calls
5. **Directus API**: Processes requests and returns responses
6. Flow reverses for reads/responses

## Key Design Decisions

### 1. Terraform Plugin Framework Types

**Decision**: Use `types.String`, `types.Bool`, `types.Int64`, `types.List`, `types.Map` instead of native Go types.

**Rationale**:
- Terraform requires three-state logic: null, unknown, known
- Native Go types only support two states (zero value vs set)
- Framework types integrate seamlessly with Terraform's plan/apply workflow
- Proper null handling prevents accidental overwrites

**Example**:
```go
// Good: Can represent null, unknown, or a value
Icon types.String `tfsdk:"icon" json:"icon,omitempty"`

// Bad: Empty string is indistinguishable from unset
Icon string `tfsdk:"icon" json:"icon,omitempty"`
```

### 2. Separate Input Models

**Decision**: Create separate `*CreateInput` and `*UpdateInput` models for policies and roles.

**Rationale**:
- API requirements differ for create vs update operations
- Some fields are computed and should not be sent in requests
- Prevents accidental modification of read-only fields
- Clearer intent in code

**Example**:
```go
// Main model with computed fields
type Policy struct {
    ID    types.String  // Computed on creation
    Name  types.String
    Users types.List    // Computed relationship
}

// Create input without computed fields
type PolicyCreateInput struct {
    Name  types.String  // Required
    Users types.List    // Optional, to assign users
}
```

### 3. Comprehensive Field Documentation

**Decision**: Document every field with description, requirement status, and defaults.

**Rationale**:
- Models serve as API documentation
- Developers can understand fields without checking API docs
- Clear contracts for required vs optional fields
- Examples of valid values provided inline

**Format**:
```go
// FieldName is a description of what this field does.
// Possible values: "value1", "value2", "value3"
// Required: true|false
// Default: defaultValue
FieldName types.String `tfsdk:"field_name" json:"field_name,omitempty"`
```

### 4. Nested Structures

**Decision**: Use nested structs for complex objects (Meta, Schema, etc.).

**Rationale**:
- Matches API structure exactly
- Groups related fields logically
- Reusable across different models
- Easier to validate and test

**Example**:
```go
type Collection struct {
    Collection types.String
    Meta       *CollectionMeta  // Pointer allows nil
    Schema     *CollectionSchema
}

type CollectionMeta struct {
    Icon    types.String
    Note    types.String
    // ... more fields
}
```

### 5. JSON Tags with omitempty

**Decision**: Always include `json:"field_name,omitempty"` tags.

**Rationale**:
- Omits null/empty fields from API requests
- Reduces payload size
- Prevents sending unset values
- Matches Directus API expectations

### 6. Relationship Fields as Lists

**Decision**: Model relationships (users, roles, policies) as `types.List`.

**Rationale**:
- Represents arrays of IDs correctly
- Framework handles list validation
- Easy to iterate and manipulate
- Supports empty lists vs null

**Example**:
```go
// Many-to-many relationship
Policies types.List `tfsdk:"policies" json:"policies,omitempty"` // List of UUIDs
```

## Model Responsibilities

### Collection Model
- Represents database tables with Directus metadata
- Includes nested fields and their configurations
- Handles schema information (types, constraints)
- Supports collection folders (schema=null)
- Enables features like versioning, archiving, sorting

### Policy Model
- Defines access control permissions
- Supports IP-based restrictions
- Manages 2FA enforcement
- Grants admin/app access
- Contains multiple permissions per policy
- Additive permission model (never subtractive)

### Role Model
- Organizational grouping of users
- Hierarchical with parent/child relationships
- Assigned multiple policies
- Computed effective permissions
- Special types: AdminRole, PublicRole

## Validation Strategy

Validation happens at multiple levels:

1. **Schema Level** (Terraform Plugin Framework):
   - Required fields enforced
   - Type checking
   - Format validation

2. **Resource Level** (Provider code):
   - Business logic validation
   - Cross-field validation
   - API-specific constraints

3. **API Level** (Directus):
   - Final validation
   - Database constraints
   - Permission checks

## Error Handling

Models themselves don't handle errors, but provide structure for error handling:

1. **Type Safety**: Compile-time type checking prevents common errors
2. **Nil Checks**: Pointer fields require nil checks before access
3. **IsNull/IsUnknown**: Framework types provide state checking methods
4. **Diagnostics**: Resource code translates model issues to Terraform diagnostics

## Testing Strategy

### Unit Tests
- Test JSON serialization/deserialization
- Validate tfsdk tag correctness
- Test null/unknown/known value handling
- Verify omitempty behavior

### Integration Tests
- Test full create/read/update/delete cycles
- Verify API compatibility
- Test relationship handling
- Validate computed fields

### Example Test:
```go
func TestCollectionSerialization(t *testing.T) {
    collection := models.Collection{
        Collection: types.StringValue("test"),
        Meta: &models.CollectionMeta{
            Icon: types.StringValue("star"),
            Hidden: types.BoolValue(false),
        },
    }

    json, err := json.Marshal(collection)
    // Assert correct JSON output
}
```

## Future Enhancements

### Planned Additions
1. **Relation Models**: Full relationship type support (M2O, O2M, M2M, M2A)
2. **Field Validation**: Built-in validation rules
3. **Default Values**: Smart defaults based on Directus behavior
4. **Migration Helpers**: Support for schema migrations
5. **Diff Detection**: Efficient change detection

### API Version Compatibility
- Models are designed for Directus v11+
- Policy-based permissions (v11 feature)
- Deprecated fields noted for backward compatibility
- Version detection in provider configuration

## Dependencies

- `github.com/hashicorp/terraform-plugin-framework`: Core framework types
- Standard library: `encoding/json` for serialization

No external API client dependencies in models - they are pure data structures.

## Best Practices for Model Updates

When updating models:

1. **Never remove fields**: Mark as deprecated instead
2. **Add optional fields**: Always use pointers or omitempty
3. **Document changes**: Update comments and CHANGELOG
4. **Update tests**: Add test cases for new fields
5. **Check API docs**: Verify against latest Directus API spec
6. **Version compatibility**: Note minimum Directus version if applicable

## Maintenance

### Regular Reviews
- Quarterly: Check against latest Directus API documentation
- Per release: Verify compatibility with new Directus versions
- On issues: Update based on user feedback

### Version Control
- Models match provider version
- Breaking changes trigger major version bump
- New fields trigger minor version bump
- Bug fixes trigger patch version bump

## References

- [Directus API Reference](https://docs.directus.io/reference/introduction)
- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [Go JSON Encoding](https://pkg.go.dev/encoding/json)
