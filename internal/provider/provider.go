package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/kylindc/terraform-provider-directus/internal/client"
)

var _ provider.Provider = &DirectusProvider{}

type DirectusProvider struct {
	version string
}

type DirectusProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DirectusProvider{
			version: version,
		}
	}
}

func (p *DirectusProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "directus"
	resp.Version = p.version
}

func (p *DirectusProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing Directus resources.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "The Directus instance endpoint URL.",
				Required:    true,
			},
			"token": schema.StringAttribute{
				Description: "Static token for authentication.",
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *DirectusProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config DirectusProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Initialize Directus client
	directusClient, err := client.NewClient(ctx, client.Config{
		BaseURL: config.Endpoint.ValueString(),
		Token:   config.Token.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Directus Client",
			"An unexpected error occurred when creating the Directus API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Directus Client Error: "+err.Error(),
		)
		return
	}

	// Make the client available to resources
	resp.DataSourceData = directusClient
	resp.ResourceData = directusClient
}

func (p *DirectusProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPolicyResource,
		NewRoleResource,
		NewRolePoliciesAttachmentResource,
		NewCollectionResource,
	}
}

func (p *DirectusProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// TODO: Add data sources if needed
	}
}
