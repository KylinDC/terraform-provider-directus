package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/kylindc/terraform-provider-directus/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CollectionResource{}
var _ resource.ResourceWithImportState = &CollectionResource{}

// NewCollectionResource creates a new collection resource.
func NewCollectionResource() resource.Resource {
	return &CollectionResource{}
}

// CollectionResource defines the resource implementation.
type CollectionResource struct {
	client *client.Client
}

// CollectionResourceModel describes the resource data model.
type CollectionResourceModel struct {
	Collection types.String `tfsdk:"collection"`
	Icon       types.String `tfsdk:"icon"`
	Note       types.String `tfsdk:"note"`
	Hidden     types.Bool   `tfsdk:"hidden"`
	Singleton  types.Bool   `tfsdk:"singleton"`
	SortField  types.String `tfsdk:"sort_field"`
	Archive    types.String `tfsdk:"archive_field"`
	Color      types.String `tfsdk:"color"`
}

func (r *CollectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_collection"
}

func (r *CollectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Directus Collection resource. Collections are the foundation of Directus, " +
			"representing database tables with additional metadata and configuration.",

		Attributes: map[string]schema.Attribute{
			"collection": schema.StringAttribute{
				MarkdownDescription: "The unique name of the collection. This is used as the table name in the database.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"icon": schema.StringAttribute{
				MarkdownDescription: "The name of a Google Material Design Icon assigned to this collection.",
				Optional:            true,
			},
			"note": schema.StringAttribute{
				MarkdownDescription: "A short description displayed in the Data Studio.",
				Optional:            true,
			},
			"hidden": schema.BoolAttribute{
				MarkdownDescription: "Whether this collection is hidden from the Data Studio.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"singleton": schema.BoolAttribute{
				MarkdownDescription: "Whether this collection should be treated as a singleton (single item).",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"sort_field": schema.StringAttribute{
				MarkdownDescription: "The field used for manual sorting of items.",
				Optional:            true,
			},
			"archive_field": schema.StringAttribute{
				MarkdownDescription: "The field used to archive items (soft delete).",
				Optional:            true,
			},
			"color": schema.StringAttribute{
				MarkdownDescription: "A color hex code associated with this collection (e.g., #6644FF).",
				Optional:            true,
			},
		},
	}
}

func (r *CollectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CollectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CollectionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build create input
	createInput := buildCollectionInput(data, true)

	// Create collection via API
	var result struct {
		Data collectionAPIResponse `json:"data"`
	}

	if err := r.client.Create(ctx, "collections", createInput, &result); err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Collection",
			"Could not create collection, unexpected error: "+err.Error(),
		)
		return
	}

	// Convert API response to model
	createdCollection := result.Data.toModel()
	data.Collection = createdCollection.Collection
	data.Icon = createdCollection.Icon
	data.Note = createdCollection.Note
	data.Hidden = createdCollection.Hidden
	data.Singleton = createdCollection.Singleton
	data.SortField = createdCollection.SortField
	data.Archive = createdCollection.Archive
	data.Color = createdCollection.Color

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CollectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CollectionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get collection from API
	var result struct {
		Data collectionAPIResponse `json:"data"`
	}

	if err := r.client.Get(ctx, "collections", data.Collection.ValueString(), &result); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Collection",
			"Could not read collection "+data.Collection.ValueString()+": "+err.Error(),
		)
		return
	}

	// Convert API response to model
	readCollection := result.Data.toModel()
	data.Icon = readCollection.Icon
	data.Note = readCollection.Note
	data.Hidden = readCollection.Hidden
	data.Singleton = readCollection.Singleton
	data.SortField = readCollection.SortField
	data.Archive = readCollection.Archive
	data.Color = readCollection.Color

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CollectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CollectionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build update input
	updateInput := buildCollectionInput(data, false)

	// Update collection via API
	var result struct {
		Data collectionAPIResponse `json:"data"`
	}

	if err := r.client.Update(ctx, "collections", data.Collection.ValueString(), updateInput, &result); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Collection",
			"Could not update collection "+data.Collection.ValueString()+": "+err.Error(),
		)
		return
	}

	// Convert API response to model
	updatedCollection := result.Data.toModel()
	data.Icon = updatedCollection.Icon
	data.Note = updatedCollection.Note
	data.Hidden = updatedCollection.Hidden
	data.Singleton = updatedCollection.Singleton
	data.SortField = updatedCollection.SortField
	data.Archive = updatedCollection.Archive
	data.Color = updatedCollection.Color

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CollectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CollectionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete collection via API
	if err := r.client.Delete(ctx, "collections", data.Collection.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Collection",
			"Could not delete collection "+data.Collection.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *CollectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("collection"), req, resp)
}

// collectionAPIResponse represents the API response for collection operations with plain Go types
type collectionAPIResponse struct {
	Collection string                    `json:"collection"`
	Meta       *collectionMetaResponse   `json:"meta,omitempty"`
	Schema     *collectionSchemaResponse `json:"schema,omitempty"`
}

type collectionMetaResponse struct {
	Collection   string `json:"collection,omitempty"`
	Icon         string `json:"icon,omitempty"`
	Note         string `json:"note,omitempty"`
	Hidden       bool   `json:"hidden,omitempty"`
	Singleton    bool   `json:"singleton,omitempty"`
	SortField    string `json:"sort_field,omitempty"`
	ArchiveField string `json:"archive_field,omitempty"`
	Color        string `json:"color,omitempty"`
}

type collectionSchemaResponse struct {
	Name    string `json:"name,omitempty"`
	Comment string `json:"comment,omitempty"`
}

// toModel converts collectionAPIResponse to CollectionResourceModel
func (c *collectionAPIResponse) toModel() *CollectionResourceModel {
	collection := &CollectionResourceModel{
		Collection: types.StringValue(c.Collection),
		Hidden:     types.BoolValue(false),
		Singleton:  types.BoolValue(false),
		Icon:       types.StringNull(),
		Note:       types.StringNull(),
		SortField:  types.StringNull(),
		Archive:    types.StringNull(),
		Color:      types.StringNull(),
	}

	// Extract values from meta if present
	if c.Meta != nil {
		collection.Icon = stringOrNull(c.Meta.Icon)
		collection.Note = stringOrNull(c.Meta.Note)
		collection.Hidden = types.BoolValue(c.Meta.Hidden)
		collection.Singleton = types.BoolValue(c.Meta.Singleton)
		collection.SortField = stringOrNull(c.Meta.SortField)
		collection.Archive = stringOrNull(c.Meta.ArchiveField)
		collection.Color = stringOrNull(c.Meta.Color)
	}

	return collection
}

// buildCollectionInput constructs the input from the resource model (used for both create and update)
func buildCollectionInput(data CollectionResourceModel, isCreate bool) map[string]interface{} {
	input := make(map[string]interface{})

	// Collection name and schema are required for create
	if isCreate {
		input["collection"] = data.Collection.ValueString()
		// schema: {} is required to create a real database table.
		// Without it, Directus creates a virtual collection (folder) with no backing table.
		input["schema"] = map[string]interface{}{}
	}

	// Build meta object
	meta := make(map[string]interface{})
	setStringField(meta, "icon", data.Icon)
	setStringField(meta, "note", data.Note)
	setBoolField(meta, "hidden", data.Hidden)
	setBoolField(meta, "singleton", data.Singleton)
	setStringField(meta, "sort_field", data.SortField)
	setStringField(meta, "archive_field", data.Archive)
	setStringField(meta, "color", data.Color)

	if len(meta) > 0 {
		input["meta"] = meta
	}

	return input
}
