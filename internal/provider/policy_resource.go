package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/kylindc/terraform-provider-directus/internal/client"
)

var (
	_ resource.Resource                = &PolicyResource{}
	_ resource.ResourceWithConfigure   = &PolicyResource{}
	_ resource.ResourceWithImportState = &PolicyResource{}
)

// NewPolicyResource creates a new Policy resource.
func NewPolicyResource() resource.Resource {
	return &PolicyResource{}
}

// PolicyResource defines the resource implementation.
type PolicyResource struct {
	client *client.Client
}

// PolicyResourceModel defines the model for the resource (simplified from models.Policy)
type PolicyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Icon        types.String `tfsdk:"icon"`
	Description types.String `tfsdk:"description"`
	IPAccess    types.String `tfsdk:"ip_access"`
	EnforceTFA  types.Bool   `tfsdk:"enforce_tfa"`
	AdminAccess types.Bool   `tfsdk:"admin_access"`
	AppAccess   types.Bool   `tfsdk:"app_access"`
}

// Metadata returns the resource type name.
func (r *PolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

// Schema defines the schema for the resource.
func (r *PolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Directus access policy. Policies are composable units that define a specific set of access permissions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the policy (UUID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the policy.",
				Required:    true,
			},
			"icon": schema.StringAttribute{
				Description: "The name of a Google Material Design Icon assigned to this policy.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description for the policy, displayed in the Data Studio.",
				Optional:    true,
			},
			"ip_access": schema.StringAttribute{
				Description: "A CSV of IP addresses that this policy applies to. Allows you to configure an allowlist of IP addresses, IP ranges, and CIDR blocks.",
				Optional:    true,
			},
			"enforce_tfa": schema.BoolAttribute{
				Description: "Whether Two-Factor Authentication is required for users with this policy.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"admin_access": schema.BoolAttribute{
				Description: "Grants users with this policy full admin access to everything.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"app_access": schema.BoolAttribute{
				Description: "Determines whether users with this policy have access to the Data Studio.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *PolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates a new policy.
func (r *PolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build request body
	reqBody := map[string]interface{}{
		"name": plan.Name.ValueString(),
	}
	setStringField(reqBody, "icon", plan.Icon)
	setStringField(reqBody, "description", plan.Description)
	setIPAccessField(reqBody, plan.IPAccess)
	setBoolField(reqBody, "enforce_tfa", plan.EnforceTFA)
	setBoolField(reqBody, "admin_access", plan.AdminAccess)
	setBoolField(reqBody, "app_access", plan.AppAccess)

	// Call the API
	var response struct {
		Data policyAPIResponse `json:"data"`
	}

	if err := r.client.Create(ctx, "policies", reqBody, &response); err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Policy",
			fmt.Sprintf("Could not create policy: %s", err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, response.Data.toModel())...)
}

// Read reads the policy.
func (r *PolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var response struct {
		Data policyAPIResponse `json:"data"`
	}

	if err := r.client.Get(ctx, "policies", state.ID.ValueString(), &response); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Policy",
			fmt.Sprintf("Could not read policy %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, response.Data.toModel())...)
}

// Update updates the policy.
func (r *PolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build request body
	reqBody := make(map[string]interface{})
	setStringField(reqBody, "name", plan.Name)
	setStringField(reqBody, "icon", plan.Icon)
	setStringField(reqBody, "description", plan.Description)
	setIPAccessField(reqBody, plan.IPAccess)
	setBoolField(reqBody, "enforce_tfa", plan.EnforceTFA)
	setBoolField(reqBody, "admin_access", plan.AdminAccess)
	setBoolField(reqBody, "app_access", plan.AppAccess)

	// Call the API
	var response struct {
		Data policyAPIResponse `json:"data"`
	}

	if err := r.client.Update(ctx, "policies", plan.ID.ValueString(), reqBody, &response); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Policy",
			fmt.Sprintf("Could not update policy %s: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, response.Data.toModel())...)
}

// Delete deletes the policy.
func (r *PolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete(ctx, "policies", state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Policy",
			fmt.Sprintf("Could not delete policy %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports the policy by ID.
func (r *PolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// policyAPIResponse represents the API response for policy operations
type policyAPIResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon,omitempty"`
	Description string   `json:"description,omitempty"`
	IPAccess    []string `json:"ip_access,omitempty"`
	EnforceTFA  bool     `json:"enforce_tfa,omitempty"`
	AdminAccess bool     `json:"admin_access,omitempty"`
	AppAccess   bool     `json:"app_access,omitempty"`
}

// toModel converts policyAPIResponse to PolicyResourceModel
func (p *policyAPIResponse) toModel() *PolicyResourceModel {
	// Convert ip_access array from API to comma-separated string for Terraform
	var ipAccess types.String
	if len(p.IPAccess) > 0 {
		ipAccess = types.StringValue(strings.Join(p.IPAccess, ","))
	} else {
		ipAccess = types.StringNull()
	}

	return &PolicyResourceModel{
		ID:          types.StringValue(p.ID),
		Name:        types.StringValue(p.Name),
		Icon:        stringOrNull(p.Icon),
		Description: stringOrNull(p.Description),
		IPAccess:    ipAccess,
		EnforceTFA:  types.BoolValue(p.EnforceTFA),
		AdminAccess: types.BoolValue(p.AdminAccess),
		AppAccess:   types.BoolValue(p.AppAccess),
	}
}

// setIPAccessField converts the CSV ip_access string to a JSON array for the Directus API.
func setIPAccessField(input map[string]interface{}, value types.String) {
	if value.IsNull() || value.IsUnknown() {
		return
	}
	csv := value.ValueString()
	if csv == "" {
		return
	}
	parts := strings.Split(csv, ",")
	trimmed := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			trimmed = append(trimmed, s)
		}
	}
	if len(trimmed) > 0 {
		input["ip_access"] = trimmed
	}
}

