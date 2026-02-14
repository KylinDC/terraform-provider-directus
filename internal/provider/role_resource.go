package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/kylindc/terraform-provider-directus/internal/client"
)

var _ resource.ResourceWithConfigure = &RoleResource{}

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RoleResource{}
var _ resource.ResourceWithImportState = &RoleResource{}

// NewRoleResource creates a new role resource.
func NewRoleResource() resource.Resource {
	return &RoleResource{}
}

// RoleResource defines the resource implementation.
type RoleResource struct {
	client *client.Client
}

// RoleResourceModel describes the resource data model.
// Policy associations are managed separately via directus_role_policies_attachment.
type RoleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Icon        types.String `tfsdk:"icon"`
	Description types.String `tfsdk:"description"`
	Parent      types.String `tfsdk:"parent"`
	Children    types.List   `tfsdk:"children"` // List of child role UUIDs (computed)
	Users       types.List   `tfsdk:"users"`    // List of user UUIDs (computed)
}

func (r *RoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *RoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Directus Role resource. Roles define user positions and permissions within a project. " +
			"Roles can inherit from parent roles and have policies assigned via the directus_access junction table.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the role (UUID).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the role.",
				Required:            true,
			},
			"icon": schema.StringAttribute{
				MarkdownDescription: "The name of a Google Material Design Icon assigned to this role.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the role.",
				Optional:            true,
			},
			"parent": schema.StringAttribute{
				MarkdownDescription: "The ID of the parent role. Child roles inherit permissions from their parent. " +
					"Note: Circular references are not allowed.",
				Optional: true,
			},
			"children": schema.ListAttribute{
				MarkdownDescription: "List of child role UUIDs that inherit from this role. This is a computed field.",
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"users": schema.ListAttribute{
				MarkdownDescription: "List of user UUIDs assigned to this role. This is a computed field.",
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *RoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RoleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build create input
	createInput := buildRoleInput(data, true)

	// Create role via API
	var result struct {
		Data roleAPIResponse `json:"data"`
	}

	if err := r.client.Create(ctx, "roles", createInput, &result); err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Role",
			"Could not create role, unexpected error: "+err.Error(),
		)
		return
	}

	// Convert API response to model
	createdRole := result.Data.toModel()
	data.ID = createdRole.ID
	data.Name = createdRole.Name
	data.Icon = createdRole.Icon
	data.Description = createdRole.Description
	data.Parent = createdRole.Parent
	data.Children = createdRole.Children
	data.Users = createdRole.Users

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RoleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get role from API
	var result struct {
		Data roleAPIResponse `json:"data"`
	}

	if err := r.client.Get(ctx, "roles", data.ID.ValueString(), &result); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Role",
			"Could not read role ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Convert API response to model
	readRole := result.Data.toModel()
	data.Name = readRole.Name
	data.Icon = readRole.Icon
	data.Description = readRole.Description
	data.Parent = readRole.Parent
	data.Children = readRole.Children
	data.Users = readRole.Users

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RoleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build update input
	updateInput := buildRoleInput(data, false)

	// Update role via API
	var result struct {
		Data roleAPIResponse `json:"data"`
	}

	if err := r.client.Update(ctx, "roles", data.ID.ValueString(), updateInput, &result); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Role",
			"Could not update role ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Convert API response to model
	updatedRole := result.Data.toModel()
	data.Name = updatedRole.Name
	data.Icon = updatedRole.Icon
	data.Description = updatedRole.Description
	data.Parent = updatedRole.Parent
	data.Children = updatedRole.Children
	data.Users = updatedRole.Users

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RoleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if role has children - warn user
	if !data.Children.IsNull() && !data.Children.IsUnknown() {
		var children []string
		data.Children.ElementsAs(ctx, &children, false)
		if len(children) > 0 {
			resp.Diagnostics.AddWarning(
				"Deleting Role with Children",
				fmt.Sprintf("This role has %d child role(s). The children will become orphaned.", len(children)),
			)
		}
	}

	// Delete role via API
	if err := r.client.Delete(ctx, "roles", data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Role",
			"Could not delete role ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *RoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// roleAPIResponse represents the API response for role operations with plain Go types.
// Per the Directus Roles API, response fields are: id, name, icon, description, parent,
// children, policies, users. Note: admin_access/app_access do NOT exist on roles — they
// are policy-level fields only.
type roleAPIResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon,omitempty"`
	Description string   `json:"description,omitempty"`
	Parent      string   `json:"parent,omitempty"`
	Children    []string `json:"children,omitempty"`
	Users       []string `json:"users,omitempty"`
}

// toModel converts roleAPIResponse to RoleResourceModel
func (r *roleAPIResponse) toModel() *RoleResourceModel {
	return &RoleResourceModel{
		ID:          types.StringValue(r.ID),
		Name:        types.StringValue(r.Name),
		Icon:        stringOrNull(r.Icon),
		Description: stringOrNull(r.Description),
		Parent:      stringOrNull(r.Parent),
		Children:    stringListOrNull(r.Children),
		Users:       stringListOrNull(r.Users),
	}
}

// buildRoleInput constructs the input from the resource model (used for both create and update)
func buildRoleInput(data RoleResourceModel, isCreate bool) map[string]interface{} {
	input := make(map[string]interface{})

	// Name is required for create, optional for update
	if isCreate {
		input["name"] = data.Name.ValueString()
	} else {
		setStringField(input, "name", data.Name)
	}

	setStringField(input, "icon", data.Icon)
	setStringField(input, "description", data.Description)

	// Parent requires special handling: when removed from config (null),
	// we must explicitly send null to the API to clear the relationship.
	if !isCreate {
		setNullableStringField(input, "parent", data.Parent)
	} else {
		setStringField(input, "parent", data.Parent)
	}

	return input
}
