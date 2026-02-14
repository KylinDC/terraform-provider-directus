package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Policy represents a Directus access policy.
// Policies are composable units that define a specific set of access permissions
// and can be assigned to both roles and users. Multiple policies are additive,
// meaning each policy adds permissions but never takes them away.
type Policy struct {
	// ID is the unique identifier for the policy (UUID).
	// Optional: true (computed)
	ID types.String `tfsdk:"id" json:"id,omitempty"`

	// Name is the name of the policy.
	// Required: true
	Name types.String `tfsdk:"name" json:"name"`

	// Icon is the name of a Google Material Design Icon assigned to this policy.
	// Optional: true
	Icon types.String `tfsdk:"icon" json:"icon,omitempty"`

	// Description is a description for the policy, displayed in the Data Studio.
	// Optional: true
	Description types.String `tfsdk:"description" json:"description,omitempty"`

	// IPAccess is a CSV of IP addresses that this policy applies to.
	// Allows you to configure an allowlist of IP addresses, IP ranges, and CIDR blocks.
	// If empty, no IP restrictions are applied.
	// Optional: true
	// Example: "192.168.1.1,10.0.0.0/8,172.16.0.0-172.16.255.255"
	IPAccess types.String `tfsdk:"ip_access" json:"ip_access,omitempty"`

	// EnforceTFA determines whether Two-Factor Authentication is required for users with this policy.
	// Optional: true
	// Default: false
	EnforceTFA types.Bool `tfsdk:"enforce_tfa" json:"enforce_tfa,omitempty"`

	// AdminAccess grants users with this policy full admin access to everything.
	// This means complete, unrestricted control over the project, including data model and all data.
	// Optional: true
	// Default: false
	AdminAccess types.Bool `tfsdk:"admin_access" json:"admin_access,omitempty"`

	// AppAccess determines whether users with this policy have access to the Data Studio.
	// If false, users can only access the project via API.
	// Optional: true
	// Default: false
	AppAccess types.Bool `tfsdk:"app_access" json:"app_access,omitempty"`

	// Users contains the user IDs that this policy is assigned to directly.
	// This does not include users who receive this policy through a role.
	// Many-to-many relationship to users via the directus_access collection.
	// Optional: true
	Users types.List `tfsdk:"users" json:"users,omitempty"` // List of user UUIDs (string)

	// Roles contains the role IDs that this policy is assigned to.
	// Many-to-many relationship to roles via the directus_access collection.
	// Optional: true
	Roles types.List `tfsdk:"roles" json:"roles,omitempty"` // List of role UUIDs (string)

	// Permissions contains the permission IDs assigned to this policy.
	// One-to-many relationship to permissions.
	// Optional: true
	Permissions types.List `tfsdk:"permissions" json:"permissions,omitempty"` // List of permission IDs (integer)
}

// Permission represents a single permission within a policy.
// A permission is scoped to a collection and an action (create, read, update, delete, share).
type Permission struct {
	// ID is the unique identifier for the permission.
	// Optional: true (computed)
	ID types.Int64 `tfsdk:"id" json:"id,omitempty"`

	// Policy is the ID of the policy this permission belongs to.
	// Required: true (for create)
	Policy types.String `tfsdk:"policy" json:"policy,omitempty"`

	// Collection is the name of the collection this permission applies to.
	// Required: true
	Collection types.String `tfsdk:"collection" json:"collection"`

	// Action is the action this permission applies to.
	// Possible values: "create", "read", "update", "delete", "share"
	// Required: true
	Action types.String `tfsdk:"action" json:"action"`

	// Permissions defines the level of access.
	// Can be:
	// - null or empty: No access
	// - "full": Full access to all items/fields
	// - JSON object: Custom permissions with filters and field restrictions
	// Optional: true
	Permissions types.String `tfsdk:"permissions" json:"permissions,omitempty"` // JSON string

	// ValidationRule is a custom validation rule that must pass for the action to be allowed.
	// Uses filter syntax.
	// Optional: true
	ValidationRule types.String `tfsdk:"validation" json:"validation,omitempty"` // JSON string

	// Presets contains default field values that are automatically applied.
	// Optional: true
	Presets types.String `tfsdk:"presets" json:"presets,omitempty"` // JSON string

	// Fields contains the list of fields that this permission applies to.
	// If empty or null, no fields are accessible.
	// If contains "*", all fields are accessible.
	// Optional: true
	Fields types.List `tfsdk:"fields" json:"fields,omitempty"` // List of field names (string)
}

// PolicyCreateInput represents the input for creating a new policy.
type PolicyCreateInput struct {
	// Name is the name of the policy.
	// Required: true
	Name types.String `tfsdk:"name" json:"name"`

	// Icon is the name of a Google Material Design Icon assigned to this policy.
	// Optional: true
	Icon types.String `tfsdk:"icon" json:"icon,omitempty"`

	// Description is a description for the policy.
	// Optional: true
	Description types.String `tfsdk:"description" json:"description,omitempty"`

	// IPAccess is a CSV of IP addresses that this policy applies to.
	// Optional: true
	IPAccess types.String `tfsdk:"ip_access" json:"ip_access,omitempty"`

	// EnforceTFA determines whether 2FA is required.
	// Optional: true
	EnforceTFA types.Bool `tfsdk:"enforce_tfa" json:"enforce_tfa,omitempty"`

	// AdminAccess grants admin access.
	// Optional: true
	AdminAccess types.Bool `tfsdk:"admin_access" json:"admin_access,omitempty"`

	// AppAccess grants access to the Data Studio.
	// Optional: true
	AppAccess types.Bool `tfsdk:"app_access" json:"app_access,omitempty"`

	// Users contains user IDs to assign this policy to.
	// Optional: true
	Users types.List `tfsdk:"users" json:"users,omitempty"`

	// Roles contains role IDs to assign this policy to.
	// Optional: true
	Roles types.List `tfsdk:"roles" json:"roles,omitempty"`

	// Permissions contains permissions to create with this policy.
	// Optional: true
	Permissions types.List `tfsdk:"permissions" json:"permissions,omitempty"`
}

// PolicyUpdateInput represents the input for updating a policy.
type PolicyUpdateInput struct {
	// Name is the name of the policy.
	// Optional: true
	Name types.String `tfsdk:"name" json:"name,omitempty"`

	// Icon is the icon for the policy.
	// Optional: true
	Icon types.String `tfsdk:"icon" json:"icon,omitempty"`

	// Description is the description for the policy.
	// Optional: true
	Description types.String `tfsdk:"description" json:"description,omitempty"`

	// IPAccess is the IP access configuration.
	// Optional: true
	IPAccess types.String `tfsdk:"ip_access" json:"ip_access,omitempty"`

	// EnforceTFA determines whether 2FA is required.
	// Optional: true
	EnforceTFA types.Bool `tfsdk:"enforce_tfa" json:"enforce_tfa,omitempty"`

	// AdminAccess grants admin access.
	// Optional: true
	AdminAccess types.Bool `tfsdk:"admin_access" json:"admin_access,omitempty"`

	// AppAccess grants access to the Data Studio.
	// Optional: true
	AppAccess types.Bool `tfsdk:"app_access" json:"app_access,omitempty"`

	// Users contains user IDs to assign this policy to.
	// Optional: true
	Users types.List `tfsdk:"users" json:"users,omitempty"`

	// Roles contains role IDs to assign this policy to.
	// Optional: true
	Roles types.List `tfsdk:"roles" json:"roles,omitempty"`

	// Permissions contains permissions for this policy.
	// Optional: true
	Permissions types.List `tfsdk:"permissions" json:"permissions,omitempty"`
}
