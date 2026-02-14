package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Role represents a Directus role.
// Roles are organizational tools that define a user's position within a project.
// A role can have any number of policies and can be applied to any number of users.
// Roles can also have child roles that inherit permissions from the parent.
type Role struct {
	// ID is the unique identifier for the role (UUID).
	// Optional: true (computed)
	ID types.String `tfsdk:"id" json:"id,omitempty"`

	// Name is the name of the role.
	// Required: true
	Name types.String `tfsdk:"name" json:"name"`

	// Icon is the name of a Google Material Design Icon assigned to this role.
	// Optional: true
	Icon types.String `tfsdk:"icon" json:"icon,omitempty"`

	// Description is a description of the role.
	// Optional: true
	Description types.String `tfsdk:"description" json:"description,omitempty"`

	// Parent is the ID of the optional parent role that this role inherits permissions from.
	// Many-to-one relationship to roles.
	// Optional: true
	Parent types.String `tfsdk:"parent" json:"parent,omitempty"` // UUID of parent role

	// Children contains the nested child roles that inherit this role's permissions.
	// One-to-many relationship to roles.
	// Optional: true (computed)
	Children types.List `tfsdk:"children" json:"children,omitempty"` // List of role UUIDs (string)

	// Policies contains the policy IDs assigned to this role.
	// Many-to-many relationship to policies via the directus_access collection.
	// Optional: true
	Policies types.List `tfsdk:"policies" json:"policies,omitempty"` // List of policy UUIDs (string)

	// Users contains the users assigned to this role.
	// One-to-many relationship to users.
	// Optional: true (computed)
	Users types.List `tfsdk:"users" json:"users,omitempty"` // List of user UUIDs (string)

	// IPAccess is an array of IP addresses that are allowed to connect to the API as a user of this role.
	// Supports individual IPs, IP ranges, and CIDR blocks.
	// Optional: true
	// Note: This field is deprecated in favor of policy-level IP access.
	IPAccess types.List `tfsdk:"ip_access" json:"ip_access,omitempty"` // List of IP address strings

	// EnforceTFA determines whether this role enforces the use of Two-Factor Authentication.
	// Optional: true
	// Default: false
	// Note: This field is deprecated in favor of policy-level 2FA enforcement.
	EnforceTFA types.Bool `tfsdk:"enforce_tfa" json:"enforce_tfa,omitempty"`

	// ExternalID is the ID used with external services in SCIM.
	// Used for integration with external identity providers.
	// Optional: true
	ExternalID types.String `tfsdk:"external_id" json:"external_id,omitempty"`

	// ModuleListing is a custom override for the admin app module bar navigation.
	// Optional: true
	ModuleListing types.String `tfsdk:"module_listing" json:"module_listing,omitempty"` // JSON string
}

// RoleCreateInput represents the input for creating a new role.
type RoleCreateInput struct {
	// Name is the name of the role.
	// Required: true
	Name types.String `tfsdk:"name" json:"name"`

	// Icon is the icon for the role.
	// Optional: true
	Icon types.String `tfsdk:"icon" json:"icon,omitempty"`

	// Description is the description of the role.
	// Optional: true
	Description types.String `tfsdk:"description" json:"description,omitempty"`

	// Parent is the ID of the parent role.
	// Optional: true
	Parent types.String `tfsdk:"parent" json:"parent,omitempty"`

	// Policies contains policy IDs to assign to this role.
	// Optional: true
	Policies types.List `tfsdk:"policies" json:"policies,omitempty"`

	// IPAccess contains IP addresses allowed for this role.
	// Optional: true
	// Note: Deprecated in favor of policy-level IP access.
	IPAccess types.List `tfsdk:"ip_access" json:"ip_access,omitempty"`

	// EnforceTFA determines whether 2FA is enforced.
	// Optional: true
	// Note: Deprecated in favor of policy-level 2FA enforcement.
	EnforceTFA types.Bool `tfsdk:"enforce_tfa" json:"enforce_tfa,omitempty"`

	// ExternalID is the ID for external services.
	// Optional: true
	ExternalID types.String `tfsdk:"external_id" json:"external_id,omitempty"`

	// ModuleListing is the custom module bar navigation.
	// Optional: true
	ModuleListing types.String `tfsdk:"module_listing" json:"module_listing,omitempty"`
}

// RoleUpdateInput represents the input for updating a role.
type RoleUpdateInput struct {
	// Name is the name of the role.
	// Optional: true
	Name types.String `tfsdk:"name" json:"name,omitempty"`

	// Icon is the icon for the role.
	// Optional: true
	Icon types.String `tfsdk:"icon" json:"icon,omitempty"`

	// Description is the description of the role.
	// Optional: true
	Description types.String `tfsdk:"description" json:"description,omitempty"`

	// Parent is the ID of the parent role.
	// Optional: true
	Parent types.String `tfsdk:"parent" json:"parent,omitempty"`

	// Policies contains policy IDs to assign to this role.
	// Optional: true
	Policies types.List `tfsdk:"policies" json:"policies,omitempty"`

	// IPAccess contains IP addresses allowed for this role.
	// Optional: true
	// Note: Deprecated in favor of policy-level IP access.
	IPAccess types.List `tfsdk:"ip_access" json:"ip_access,omitempty"`

	// EnforceTFA determines whether 2FA is enforced.
	// Optional: true
	// Note: Deprecated in favor of policy-level 2FA enforcement.
	EnforceTFA types.Bool `tfsdk:"enforce_tfa" json:"enforce_tfa,omitempty"`

	// ExternalID is the ID for external services.
	// Optional: true
	ExternalID types.String `tfsdk:"external_id" json:"external_id,omitempty"`

	// ModuleListing is the custom module bar navigation.
	// Optional: true
	ModuleListing types.String `tfsdk:"module_listing" json:"module_listing,omitempty"`
}

// RoleWithPermissions represents a role with all its computed permissions.
// This is useful for understanding the effective permissions a role has.
type RoleWithPermissions struct {
	Role

	// EffectivePolicies contains all policies (including inherited from parent roles).
	// Computed: true
	EffectivePolicies []Policy `tfsdk:"effective_policies" json:"effective_policies,omitempty"`

	// EffectivePermissions contains all permissions from all policies.
	// Computed: true
	EffectivePermissions []Permission `tfsdk:"effective_permissions" json:"effective_permissions,omitempty"`
}
