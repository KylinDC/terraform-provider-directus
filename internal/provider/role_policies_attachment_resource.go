package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/kylindc/terraform-provider-directus/internal/client"
)

var (
	_ resource.Resource                = &RolePoliciesAttachmentResource{}
	_ resource.ResourceWithConfigure   = &RolePoliciesAttachmentResource{}
	_ resource.ResourceWithImportState = &RolePoliciesAttachmentResource{}
)

// NewRolePoliciesAttachmentResource creates a new role-policies attachment resource.
func NewRolePoliciesAttachmentResource() resource.Resource {
	return &RolePoliciesAttachmentResource{}
}

// RolePoliciesAttachmentResource manages the link between a Directus role and
// one or more policies via the directus_access junction table.
// This resource is authoritative: it manages ALL policy attachments for the role.
type RolePoliciesAttachmentResource struct {
	client *client.Client
}

// RolePoliciesAttachmentModel describes the resource data model.
type RolePoliciesAttachmentModel struct {
	ID        types.String `tfsdk:"id"`         // Equal to role_id
	RoleID    types.String `tfsdk:"role_id"`    // The role UUID
	PolicyIDs types.Set    `tfsdk:"policy_ids"` // Set of policy UUIDs
}

// accessRecordResponse represents a single entry in the directus_access junction table
// as returned when expanding the role's policies field.
type accessRecordResponse struct {
	ID     string `json:"id"`
	Policy string `json:"policy"`
}

// roleWithPoliciesResponse is used to unmarshal GET /roles/{id}?fields=policies.id,policies.policy
type roleWithPoliciesResponse struct {
	ID       string                 `json:"id"`
	Policies []accessRecordResponse `json:"policies"`
}

func (r *RolePoliciesAttachmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_policies_attachment"
}

func (r *RolePoliciesAttachmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the attachment of one or more policies to a Directus role. " +
			"In Directus, roles and policies are linked via the `directus_access` junction table. " +
			"This resource is **authoritative**: it manages ALL policy attachments for the specified role. " +
			"Policies attached outside of Terraform will be detached on the next apply.\n\n" +
			"Import using the role UUID: `terraform import directus_role_policies_attachment.example <role_id>`.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of this resource (equal to role_id).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role_id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the role.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_ids": schema.SetAttribute{
				MarkdownDescription: "The set of policy UUIDs to attach to the role.",
				Required:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *RolePoliciesAttachmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

// Create attaches the specified policies to a role via the role API's nested relational operations.
func (r *RolePoliciesAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RolePoliciesAttachmentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleID := plan.RoleID.ValueString()

	var desiredPolicyIDs []string
	resp.Diagnostics.Append(plan.PolicyIDs.ElementsAs(ctx, &desiredPolicyIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read current policies attached to this role.
	existing, err := r.readRolePolicies(ctx, roleID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Role Policies",
			fmt.Sprintf("Could not read policies for role %s: %s", roleID, err.Error()),
		)
		return
	}

	// Build map of existing: policyID -> accessID
	existingMap := make(map[string]string)
	for _, rec := range existing {
		existingMap[rec.Policy] = rec.ID
	}

	// Compute policies to add (desired but not yet attached).
	var toCreate []map[string]interface{}
	for _, pid := range desiredPolicyIDs {
		if _, exists := existingMap[pid]; !exists {
			toCreate = append(toCreate, map[string]interface{}{"policy": pid})
		}
	}

	// Compute policies to remove (attached but not desired) — authoritative.
	desiredSet := make(map[string]bool)
	for _, pid := range desiredPolicyIDs {
		desiredSet[pid] = true
	}
	var toDelete []string
	for _, rec := range existing {
		if !desiredSet[rec.Policy] {
			toDelete = append(toDelete, rec.ID)
		}
	}

	// Apply changes via a single PATCH if needed.
	if len(toCreate) > 0 || len(toDelete) > 0 {
		policiesOp := map[string]interface{}{}
		if len(toCreate) > 0 {
			policiesOp["create"] = toCreate
		}
		if len(toDelete) > 0 {
			policiesOp["delete"] = toDelete
		}

		patchBody := map[string]interface{}{
			"policies": policiesOp,
		}

		if err := r.client.Update(ctx, "roles", roleID, patchBody, nil); err != nil {
			resp.Diagnostics.AddError(
				"Error Attaching Policies to Role",
				fmt.Sprintf("Could not update policy attachments for role %s: %s", roleID, err.Error()),
			)
			return
		}
	}

	plan.ID = types.StringValue(roleID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the policy attachment state from the Directus API.
func (r *RolePoliciesAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RolePoliciesAttachmentModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleID := state.RoleID.ValueString()

	records, err := r.readRolePolicies(ctx, roleID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Role Policies",
			fmt.Sprintf("Could not read policies for role %s: %s", roleID, err.Error()),
		)
		return
	}

	// Build the set of attached policy IDs.
	policyElements := make([]attr.Value, 0, len(records))
	for _, rec := range records {
		policyElements = append(policyElements, types.StringValue(rec.Policy))
	}

	policySet, diags := types.SetValue(types.StringType, policyElements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(roleID)
	state.PolicyIDs = policySet

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update computes the diff between current and desired policies and applies it.
func (r *RolePoliciesAttachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RolePoliciesAttachmentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleID := plan.RoleID.ValueString()

	var desiredPolicyIDs []string
	resp.Diagnostics.Append(plan.PolicyIDs.ElementsAs(ctx, &desiredPolicyIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read current policies.
	existing, err := r.readRolePolicies(ctx, roleID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Role Policies",
			fmt.Sprintf("Could not read policies for role %s: %s", roleID, err.Error()),
		)
		return
	}

	// Build map of existing: policyID -> accessID
	existingMap := make(map[string]string)
	for _, rec := range existing {
		existingMap[rec.Policy] = rec.ID
	}

	// Compute policies to add.
	var toCreate []map[string]interface{}
	for _, pid := range desiredPolicyIDs {
		if _, exists := existingMap[pid]; !exists {
			toCreate = append(toCreate, map[string]interface{}{"policy": pid})
		}
	}

	// Compute policies to remove.
	desiredSet := make(map[string]bool)
	for _, pid := range desiredPolicyIDs {
		desiredSet[pid] = true
	}
	var toDelete []string
	for _, rec := range existing {
		if !desiredSet[rec.Policy] {
			toDelete = append(toDelete, rec.ID)
		}
	}

	// Apply changes via a single PATCH if needed.
	if len(toCreate) > 0 || len(toDelete) > 0 {
		policiesOp := map[string]interface{}{}
		if len(toCreate) > 0 {
			policiesOp["create"] = toCreate
		}
		if len(toDelete) > 0 {
			policiesOp["delete"] = toDelete
		}

		patchBody := map[string]interface{}{
			"policies": policiesOp,
		}

		if err := r.client.Update(ctx, "roles", roleID, patchBody, nil); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Policy Attachments",
				fmt.Sprintf("Could not update policy attachments for role %s: %s", roleID, err.Error()),
			)
			return
		}
	}

	plan.ID = types.StringValue(roleID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes all policy attachments from the role.
func (r *RolePoliciesAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RolePoliciesAttachmentModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleID := state.RoleID.ValueString()

	// Read current policies to get access record IDs.
	existing, err := r.readRolePolicies(ctx, roleID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Role Policies",
			fmt.Sprintf("Could not read policies for role %s: %s", roleID, err.Error()),
		)
		return
	}

	if len(existing) == 0 {
		return
	}

	// Collect all access record IDs to delete.
	accessIDs := make([]string, 0, len(existing))
	for _, rec := range existing {
		accessIDs = append(accessIDs, rec.ID)
	}

	patchBody := map[string]interface{}{
		"policies": map[string]interface{}{
			"delete": accessIDs,
		},
	}

	if err := r.client.Update(ctx, "roles", roleID, patchBody, nil); err != nil {
		resp.Diagnostics.AddError(
			"Error Detaching Policies from Role",
			fmt.Sprintf("Could not detach policies from role %s: %s", roleID, err.Error()),
		)
		return
	}
}

// ImportState imports an existing role's policy attachments.
// The import ID is the role UUID.
func (r *RolePoliciesAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	roleID := req.ID
	if roleID == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID to be a non-empty role UUID.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), roleID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role_id"), roleID)...)

	// Read the role's current policies to populate policy_ids.
	records, err := r.readRolePolicies(ctx, roleID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Role Policies During Import",
			fmt.Sprintf("Could not read policies for role %s: %s", roleID, err.Error()),
		)
		return
	}

	policyElements := make([]attr.Value, 0, len(records))
	for _, rec := range records {
		policyElements = append(policyElements, types.StringValue(rec.Policy))
	}

	policySet, diags := types.SetValue(types.StringType, policyElements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_ids"), policySet)...)
}

// readRolePolicies fetches the role with expanded policies and returns the access records.
func (r *RolePoliciesAttachmentResource) readRolePolicies(ctx context.Context, roleID string) ([]accessRecordResponse, error) {
	var result struct {
		Data roleWithPoliciesResponse `json:"data"`
	}

	params := map[string]string{
		"fields": "id,policies.id,policies.policy",
	}

	if err := r.client.GetWithParams(ctx, "roles", roleID, params, &result); err != nil {
		return nil, err
	}

	return result.Data.Policies, nil
}
