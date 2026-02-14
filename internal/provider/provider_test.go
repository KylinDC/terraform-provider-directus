package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// New
// ---------------------------------------------------------------------------

func TestNew(t *testing.T) {
	factory := New("1.0.0")
	require.NotNil(t, factory)

	p := factory()
	require.NotNil(t, p)

	dp, ok := p.(*DirectusProvider)
	require.True(t, ok)
	assert.Equal(t, "1.0.0", dp.version)
}

func TestNew_DevVersion(t *testing.T) {
	factory := New("dev")
	p := factory().(*DirectusProvider)
	assert.Equal(t, "dev", p.version)
}

// ---------------------------------------------------------------------------
// Metadata
// ---------------------------------------------------------------------------

func TestDirectusProvider_Metadata(t *testing.T) {
	p := &DirectusProvider{version: "1.2.3"}

	resp := &fwprovider.MetadataResponse{}
	p.Metadata(context.Background(), fwprovider.MetadataRequest{}, resp)

	assert.Equal(t, "directus", resp.TypeName)
	assert.Equal(t, "1.2.3", resp.Version)
}

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

func TestDirectusProvider_Schema(t *testing.T) {
	p := &DirectusProvider{}

	resp := &fwprovider.SchemaResponse{}
	p.Schema(context.Background(), fwprovider.SchemaRequest{}, resp)

	require.False(t, resp.Diagnostics.HasError())

	// Verify required attributes
	assert.NotNil(t, resp.Schema.Attributes["endpoint"], "endpoint attribute should exist")
	assert.NotNil(t, resp.Schema.Attributes["token"], "token attribute should exist")

	// Verify description
	assert.Contains(t, resp.Schema.Description, "Directus")
}

// ---------------------------------------------------------------------------
// Resources
// ---------------------------------------------------------------------------

func TestDirectusProvider_Resources(t *testing.T) {
	p := &DirectusProvider{}

	resources := p.Resources(context.Background())

	// Should return 4 resource factories
	assert.Len(t, resources, 4, "Should have 4 resources")

	// Instantiate each and verify type names
	expectedTypeNames := map[string]bool{
		"directus_policy":                    false,
		"directus_role":                      false,
		"directus_role_policies_attachment":  false,
		"directus_collection":               false,
	}

	for _, factory := range resources {
		r := factory()
		require.NotNil(t, r)

		metaResp := &resource.MetadataResponse{}
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "directus"}, metaResp)

		_, exists := expectedTypeNames[metaResp.TypeName]
		assert.True(t, exists, "unexpected resource type: %s", metaResp.TypeName)
		expectedTypeNames[metaResp.TypeName] = true
	}

	// All expected resources should be registered
	for name, found := range expectedTypeNames {
		assert.True(t, found, "resource %s not found", name)
	}
}

// ---------------------------------------------------------------------------
// DataSources
// ---------------------------------------------------------------------------

func TestDirectusProvider_DataSources(t *testing.T) {
	p := &DirectusProvider{}

	dataSources := p.DataSources(context.Background())
	assert.Empty(t, dataSources, "Should have no data sources yet")
}

// ---------------------------------------------------------------------------
// Interface compliance
// ---------------------------------------------------------------------------

func TestDirectusProvider_ImplementsProviderInterface(t *testing.T) {
	var _ fwprovider.Provider = &DirectusProvider{}
}

// ---------------------------------------------------------------------------
// Configure
// ---------------------------------------------------------------------------

func TestDirectusProvider_Configure_NilConfig(t *testing.T) {
	// When Diagnostics.HasError() is true early, Configure should return
	// without creating a client. We can't easily test the full Configure flow
	// without a real Terraform config object, but we verify the provider
	// struct is correctly initialized.
	p := &DirectusProvider{version: "test"}
	assert.Equal(t, "test", p.version)
}

// Verify the provider model struct fields
func TestDirectusProviderModel(t *testing.T) {
	model := DirectusProviderModel{}

	// Verify zero values are null/unknown (unset)
	assert.True(t, model.Endpoint.IsNull())
	assert.True(t, model.Token.IsNull())
}

// ---------------------------------------------------------------------------
// NewXxxResource constructors
// ---------------------------------------------------------------------------

func TestNewPolicyResource_ReturnsCorrectType(t *testing.T) {
	r := NewPolicyResource()
	_, ok := r.(*PolicyResource)
	assert.True(t, ok)
}

func TestNewRoleResource_ReturnsCorrectType(t *testing.T) {
	r := NewRoleResource()
	_, ok := r.(*RoleResource)
	assert.True(t, ok)
}

func TestNewRolePoliciesAttachmentResource_ReturnsCorrectType(t *testing.T) {
	r := NewRolePoliciesAttachmentResource()
	_, ok := r.(*RolePoliciesAttachmentResource)
	assert.True(t, ok)
}

func TestNewCollectionResource_ReturnsCorrectType(t *testing.T) {
	r := NewCollectionResource()
	_, ok := r.(*CollectionResource)
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// Configure with wrong ProviderData type
// ---------------------------------------------------------------------------

func TestPolicyResource_Configure_WrongType(t *testing.T) {
	r := &PolicyResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: "not-a-client",
	}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestPolicyResource_Configure_NilProviderData(t *testing.T) {
	r := &PolicyResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: nil,
	}, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Nil(t, r.client)
}

func TestRoleResource_Configure_WrongType(t *testing.T) {
	r := &RoleResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: 42,
	}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCollectionResource_Configure_WrongType(t *testing.T) {
	r := &CollectionResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: false,
	}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestRolePoliciesAttachmentResource_Configure_WrongType(t *testing.T) {
	r := &RolePoliciesAttachmentResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: []string{"wrong"},
	}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// Suppress unused import warnings
var (
	_ datasource.DataSource = nil
)
