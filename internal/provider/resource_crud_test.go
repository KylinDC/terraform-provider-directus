package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test helpers for constructing Plan/State
// ---------------------------------------------------------------------------

// getResourceSchema extracts the schema from a resource.
func getResourceSchema(t *testing.T, r fwresource.Resource) rschema.Schema {
	t.Helper()
	resp := &fwresource.SchemaResponse{}
	r.Schema(context.Background(), fwresource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError())
	return resp.Schema
}

// makePlan creates a tfsdk.Plan populated with the given model.
func makePlan(t *testing.T, schema rschema.Schema, model interface{}) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: schema}
	diags := plan.Set(context.Background(), model)
	require.False(t, diags.HasError(), "makePlan: %v", diags)
	return plan
}

// makeState creates a tfsdk.State populated with the given model.
func makeState(t *testing.T, schema rschema.Schema, model interface{}) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: schema}
	diags := state.Set(context.Background(), model)
	require.False(t, diags.HasError(), "makeState: %v", diags)
	return state
}

// ===========================================================================
// Policy Resource CRUD tests
// ===========================================================================

func TestPolicyResource_Create_Full(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "http://example.com/policies", req.URL.String())

		var body map[string]interface{}
		bodyBytes, _ := io.ReadAll(req.Body)
		json.Unmarshal(bodyBytes, &body)
		assert.Equal(t, "Test Policy", body["name"])

		return mockJSONResponse(200, map[string]interface{}{
			"data": map[string]interface{}{
				"id":           "uuid-1",
				"name":         "Test Policy",
				"enforce_tfa":  false,
				"admin_access": false,
				"app_access":   true,
			},
		}), nil
	})

	r := &PolicyResource{client: mockClient}
	schema := getResourceSchema(t, r)

	plan := makePlan(t, schema, &PolicyResourceModel{
		ID:          types.StringUnknown(),
		Name:        types.StringValue("Test Policy"),
		Icon:        types.StringNull(),
		Description: types.StringNull(),
		IPAccess:    types.StringNull(),
		EnforceTFA:  types.BoolValue(false),
		AdminAccess: types.BoolValue(false),
		AppAccess:   types.BoolValue(true),
	})

	resp := &fwresource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(context.Background(), fwresource.CreateRequest{Plan: plan}, resp)

	require.False(t, resp.Diagnostics.HasError(), "Create diagnostics: %v", resp.Diagnostics)

	var result PolicyResourceModel
	resp.State.Get(context.Background(), &result)
	assert.Equal(t, "uuid-1", result.ID.ValueString())
	assert.Equal(t, "Test Policy", result.Name.ValueString())
	assert.True(t, result.AppAccess.ValueBool())
}

func TestPolicyResource_Create_Error(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockErrorResponse(400, "Validation failed"), nil
	})

	r := &PolicyResource{client: mockClient}
	schema := getResourceSchema(t, r)

	plan := makePlan(t, schema, &PolicyResourceModel{
		ID:          types.StringUnknown(),
		Name:        types.StringValue("Bad"),
		EnforceTFA:  types.BoolValue(false),
		AdminAccess: types.BoolValue(false),
		AppAccess:   types.BoolValue(false),
	})

	resp := &fwresource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(context.Background(), fwresource.CreateRequest{Plan: plan}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestPolicyResource_Read_Full(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "GET", req.Method)
		return mockJSONResponse(200, map[string]interface{}{
			"data": map[string]interface{}{
				"id":          "uuid-1",
				"name":        "Test Policy",
				"icon":        "lock",
				"enforce_tfa": true,
			},
		}), nil
	})

	r := &PolicyResource{client: mockClient}
	schema := getResourceSchema(t, r)

	state := makeState(t, schema, &PolicyResourceModel{
		ID:          types.StringValue("uuid-1"),
		Name:        types.StringValue("Test Policy"),
		EnforceTFA:  types.BoolValue(false),
		AdminAccess: types.BoolValue(false),
		AppAccess:   types.BoolValue(false),
	})

	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: schema}}
	r.Read(context.Background(), fwresource.ReadRequest{State: state}, resp)

	require.False(t, resp.Diagnostics.HasError())

	var result PolicyResourceModel
	resp.State.Get(context.Background(), &result)
	assert.Equal(t, "uuid-1", result.ID.ValueString())
	assert.Equal(t, "lock", result.Icon.ValueString())
	assert.True(t, result.EnforceTFA.ValueBool())
}

func TestPolicyResource_Read_Error(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockErrorResponse(404, "Not found"), nil
	})

	r := &PolicyResource{client: mockClient}
	schema := getResourceSchema(t, r)

	state := makeState(t, schema, &PolicyResourceModel{
		ID:          types.StringValue("nonexistent"),
		Name:        types.StringValue("Missing"),
		EnforceTFA:  types.BoolValue(false),
		AdminAccess: types.BoolValue(false),
		AppAccess:   types.BoolValue(false),
	})

	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: schema}}
	r.Read(context.Background(), fwresource.ReadRequest{State: state}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestPolicyResource_Update_Full(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "PATCH", req.Method)

		var body map[string]interface{}
		bodyBytes, _ := io.ReadAll(req.Body)
		json.Unmarshal(bodyBytes, &body)
		assert.Equal(t, "Updated", body["name"])

		return mockJSONResponse(200, map[string]interface{}{
			"data": map[string]interface{}{
				"id":          "uuid-1",
				"name":        "Updated",
				"enforce_tfa": true,
			},
		}), nil
	})

	r := &PolicyResource{client: mockClient}
	schema := getResourceSchema(t, r)

	plan := makePlan(t, schema, &PolicyResourceModel{
		ID:          types.StringValue("uuid-1"),
		Name:        types.StringValue("Updated"),
		EnforceTFA:  types.BoolValue(true),
		AdminAccess: types.BoolValue(false),
		AppAccess:   types.BoolValue(false),
	})

	resp := &fwresource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(context.Background(), fwresource.UpdateRequest{Plan: plan}, resp)

	require.False(t, resp.Diagnostics.HasError())

	var result PolicyResourceModel
	resp.State.Get(context.Background(), &result)
	assert.Equal(t, "Updated", result.Name.ValueString())
	assert.True(t, result.EnforceTFA.ValueBool())
}

func TestPolicyResource_Update_Error(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockErrorResponse(403, "Forbidden"), nil
	})

	r := &PolicyResource{client: mockClient}
	schema := getResourceSchema(t, r)

	plan := makePlan(t, schema, &PolicyResourceModel{
		ID:          types.StringValue("uuid-1"),
		Name:        types.StringValue("Updated"),
		EnforceTFA:  types.BoolValue(false),
		AdminAccess: types.BoolValue(false),
		AppAccess:   types.BoolValue(false),
	})

	resp := &fwresource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(context.Background(), fwresource.UpdateRequest{Plan: plan}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestPolicyResource_Delete_Full(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "DELETE", req.Method)
		assert.Contains(t, req.URL.String(), "/policies/uuid-1")
		return mockJSONResponse(204, nil), nil
	})

	r := &PolicyResource{client: mockClient}
	schema := getResourceSchema(t, r)

	state := makeState(t, schema, &PolicyResourceModel{
		ID:          types.StringValue("uuid-1"),
		Name:        types.StringValue("To Delete"),
		EnforceTFA:  types.BoolValue(false),
		AdminAccess: types.BoolValue(false),
		AppAccess:   types.BoolValue(false),
	})

	resp := &fwresource.DeleteResponse{State: tfsdk.State{Schema: schema}}
	r.Delete(context.Background(), fwresource.DeleteRequest{State: state}, resp)

	require.False(t, resp.Diagnostics.HasError())
}

func TestPolicyResource_Delete_Error(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockErrorResponse(500, "Server error"), nil
	})

	r := &PolicyResource{client: mockClient}
	schema := getResourceSchema(t, r)

	state := makeState(t, schema, &PolicyResourceModel{
		ID:          types.StringValue("uuid-1"),
		Name:        types.StringValue("To Delete"),
		EnforceTFA:  types.BoolValue(false),
		AdminAccess: types.BoolValue(false),
		AppAccess:   types.BoolValue(false),
	})

	resp := &fwresource.DeleteResponse{State: tfsdk.State{Schema: schema}}
	r.Delete(context.Background(), fwresource.DeleteRequest{State: state}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ===========================================================================
// Role Resource CRUD tests
// ===========================================================================

func TestRoleResource_Create_Full(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "POST", req.Method)

		var body map[string]interface{}
		bodyBytes, _ := io.ReadAll(req.Body)
		json.Unmarshal(bodyBytes, &body)
		assert.Equal(t, "Admin", body["name"])

		return mockJSONResponse(200, map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "role-uuid",
				"name": "Admin",
				"icon": "person",
			},
		}), nil
	})

	r := &RoleResource{client: mockClient}
	schema := getResourceSchema(t, r)

	plan := makePlan(t, schema, &RoleResourceModel{
		ID:          types.StringUnknown(),
		Name:        types.StringValue("Admin"),
		Icon:        types.StringNull(),
		Description: types.StringNull(),
		Parent:      types.StringNull(),
		Children:    types.ListNull(types.StringType),
		Users:       types.ListNull(types.StringType),
	})

	resp := &fwresource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(context.Background(), fwresource.CreateRequest{Plan: plan}, resp)

	require.False(t, resp.Diagnostics.HasError(), "Create diagnostics: %v", resp.Diagnostics)

	var result RoleResourceModel
	resp.State.Get(context.Background(), &result)
	assert.Equal(t, "role-uuid", result.ID.ValueString())
	assert.Equal(t, "Admin", result.Name.ValueString())
}

func TestRoleResource_Create_Error(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockErrorResponse(400, "Bad request"), nil
	})

	r := &RoleResource{client: mockClient}
	schema := getResourceSchema(t, r)

	plan := makePlan(t, schema, &RoleResourceModel{
		ID:       types.StringUnknown(),
		Name:     types.StringValue("Bad"),
		Children: types.ListNull(types.StringType),
		Users:    types.ListNull(types.StringType),
	})

	resp := &fwresource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(context.Background(), fwresource.CreateRequest{Plan: plan}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestRoleResource_Read_Full(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockJSONResponse(200, map[string]interface{}{
			"data": map[string]interface{}{
				"id":          "role-uuid",
				"name":        "Admin",
				"icon":        "person",
				"description": "Admin role",
			},
		}), nil
	})

	r := &RoleResource{client: mockClient}
	schema := getResourceSchema(t, r)

	state := makeState(t, schema, &RoleResourceModel{
		ID:       types.StringValue("role-uuid"),
		Name:     types.StringValue("Admin"),
		Children: types.ListNull(types.StringType),
		Users:    types.ListNull(types.StringType),
	})

	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: schema}}
	r.Read(context.Background(), fwresource.ReadRequest{State: state}, resp)

	require.False(t, resp.Diagnostics.HasError())

	var result RoleResourceModel
	resp.State.Get(context.Background(), &result)
	assert.Equal(t, "Admin", result.Name.ValueString())
	assert.Equal(t, "person", result.Icon.ValueString())
}

func TestRoleResource_Read_Error(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockErrorResponse(404, "Not found"), nil
	})

	r := &RoleResource{client: mockClient}
	schema := getResourceSchema(t, r)

	state := makeState(t, schema, &RoleResourceModel{
		ID:       types.StringValue("bad-id"),
		Name:     types.StringValue("Missing"),
		Children: types.ListNull(types.StringType),
		Users:    types.ListNull(types.StringType),
	})

	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: schema}}
	r.Read(context.Background(), fwresource.ReadRequest{State: state}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestRoleResource_Update_Full(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "PATCH", req.Method)
		return mockJSONResponse(200, map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "role-uuid",
				"name": "Updated Admin",
			},
		}), nil
	})

	r := &RoleResource{client: mockClient}
	schema := getResourceSchema(t, r)

	plan := makePlan(t, schema, &RoleResourceModel{
		ID:       types.StringValue("role-uuid"),
		Name:     types.StringValue("Updated Admin"),
		Children: types.ListNull(types.StringType),
		Users:    types.ListNull(types.StringType),
	})

	resp := &fwresource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(context.Background(), fwresource.UpdateRequest{Plan: plan}, resp)

	require.False(t, resp.Diagnostics.HasError())

	var result RoleResourceModel
	resp.State.Get(context.Background(), &result)
	assert.Equal(t, "Updated Admin", result.Name.ValueString())
}

func TestRoleResource_Delete_Full(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "DELETE", req.Method)
		return mockJSONResponse(204, nil), nil
	})

	r := &RoleResource{client: mockClient}
	schema := getResourceSchema(t, r)

	state := makeState(t, schema, &RoleResourceModel{
		ID:       types.StringValue("role-uuid"),
		Name:     types.StringValue("To Delete"),
		Children: types.ListNull(types.StringType),
		Users:    types.ListNull(types.StringType),
	})

	resp := &fwresource.DeleteResponse{State: tfsdk.State{Schema: schema}}
	r.Delete(context.Background(), fwresource.DeleteRequest{State: state}, resp)

	require.False(t, resp.Diagnostics.HasError())
}

func TestRoleResource_Delete_Error(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockErrorResponse(500, "Server error"), nil
	})

	r := &RoleResource{client: mockClient}
	schema := getResourceSchema(t, r)

	state := makeState(t, schema, &RoleResourceModel{
		ID:       types.StringValue("role-uuid"),
		Name:     types.StringValue("To Delete"),
		Children: types.ListNull(types.StringType),
		Users:    types.ListNull(types.StringType),
	})

	resp := &fwresource.DeleteResponse{State: tfsdk.State{Schema: schema}}
	r.Delete(context.Background(), fwresource.DeleteRequest{State: state}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ===========================================================================
// Collection Resource CRUD tests
// ===========================================================================

func TestCollectionResource_Create_Full(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "POST", req.Method)

		var body map[string]interface{}
		bodyBytes, _ := io.ReadAll(req.Body)
		json.Unmarshal(bodyBytes, &body)
		assert.Equal(t, "articles", body["collection"])
		assert.Contains(t, body, "schema")

		return mockJSONResponse(200, map[string]interface{}{
			"data": map[string]interface{}{
				"collection": "articles",
				"meta": map[string]interface{}{
					"icon":   "article",
					"hidden": false,
				},
			},
		}), nil
	})

	r := &CollectionResource{client: mockClient}
	schema := getResourceSchema(t, r)

	plan := makePlan(t, schema, &CollectionResourceModel{
		Collection: types.StringValue("articles"),
		Icon:       types.StringValue("article"),
		Hidden:     types.BoolValue(false),
		Singleton:  types.BoolValue(false),
	})

	resp := &fwresource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(context.Background(), fwresource.CreateRequest{Plan: plan}, resp)

	require.False(t, resp.Diagnostics.HasError(), "Create diagnostics: %v", resp.Diagnostics)

	var result CollectionResourceModel
	resp.State.Get(context.Background(), &result)
	assert.Equal(t, "articles", result.Collection.ValueString())
}

func TestCollectionResource_Create_Error(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockErrorResponse(400, "Validation failed"), nil
	})

	r := &CollectionResource{client: mockClient}
	schema := getResourceSchema(t, r)

	plan := makePlan(t, schema, &CollectionResourceModel{
		Collection: types.StringValue("bad"),
		Hidden:     types.BoolValue(false),
		Singleton:  types.BoolValue(false),
	})

	resp := &fwresource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(context.Background(), fwresource.CreateRequest{Plan: plan}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCollectionResource_Read_Full(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockJSONResponse(200, map[string]interface{}{
			"data": map[string]interface{}{
				"collection": "articles",
				"meta": map[string]interface{}{
					"icon":   "article",
					"note":   "Blog articles",
					"hidden": false,
				},
			},
		}), nil
	})

	r := &CollectionResource{client: mockClient}
	schema := getResourceSchema(t, r)

	state := makeState(t, schema, &CollectionResourceModel{
		Collection: types.StringValue("articles"),
		Hidden:     types.BoolValue(false),
		Singleton:  types.BoolValue(false),
	})

	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: schema}}
	r.Read(context.Background(), fwresource.ReadRequest{State: state}, resp)

	require.False(t, resp.Diagnostics.HasError())

	var result CollectionResourceModel
	resp.State.Get(context.Background(), &result)
	assert.Equal(t, "articles", result.Collection.ValueString())
	assert.Equal(t, "article", result.Icon.ValueString())
	assert.Equal(t, "Blog articles", result.Note.ValueString())
}

func TestCollectionResource_Read_Error(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockErrorResponse(404, "Not found"), nil
	})

	r := &CollectionResource{client: mockClient}
	schema := getResourceSchema(t, r)

	state := makeState(t, schema, &CollectionResourceModel{
		Collection: types.StringValue("nonexistent"),
		Hidden:     types.BoolValue(false),
		Singleton:  types.BoolValue(false),
	})

	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: schema}}
	r.Read(context.Background(), fwresource.ReadRequest{State: state}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCollectionResource_Update_Full(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "PATCH", req.Method)
		return mockJSONResponse(200, map[string]interface{}{
			"data": map[string]interface{}{
				"collection": "articles",
				"meta": map[string]interface{}{
					"note":   "Updated note",
					"hidden": true,
				},
			},
		}), nil
	})

	r := &CollectionResource{client: mockClient}
	schema := getResourceSchema(t, r)

	plan := makePlan(t, schema, &CollectionResourceModel{
		Collection: types.StringValue("articles"),
		Note:       types.StringValue("Updated note"),
		Hidden:     types.BoolValue(true),
		Singleton:  types.BoolValue(false),
	})

	resp := &fwresource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(context.Background(), fwresource.UpdateRequest{Plan: plan}, resp)

	require.False(t, resp.Diagnostics.HasError())

	var result CollectionResourceModel
	resp.State.Get(context.Background(), &result)
	assert.Equal(t, "Updated note", result.Note.ValueString())
	assert.True(t, result.Hidden.ValueBool())
}

func TestCollectionResource_Update_Error(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockErrorResponse(403, "Forbidden"), nil
	})

	r := &CollectionResource{client: mockClient}
	schema := getResourceSchema(t, r)

	plan := makePlan(t, schema, &CollectionResourceModel{
		Collection: types.StringValue("articles"),
		Hidden:     types.BoolValue(false),
		Singleton:  types.BoolValue(false),
	})

	resp := &fwresource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(context.Background(), fwresource.UpdateRequest{Plan: plan}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCollectionResource_Delete_Full(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "DELETE", req.Method)
		assert.Contains(t, req.URL.String(), "/collections/articles")
		return mockJSONResponse(204, nil), nil
	})

	r := &CollectionResource{client: mockClient}
	schema := getResourceSchema(t, r)

	state := makeState(t, schema, &CollectionResourceModel{
		Collection: types.StringValue("articles"),
		Hidden:     types.BoolValue(false),
		Singleton:  types.BoolValue(false),
	})

	resp := &fwresource.DeleteResponse{State: tfsdk.State{Schema: schema}}
	r.Delete(context.Background(), fwresource.DeleteRequest{State: state}, resp)

	require.False(t, resp.Diagnostics.HasError())
}

func TestCollectionResource_Delete_Error(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockErrorResponse(500, "Server error"), nil
	})

	r := &CollectionResource{client: mockClient}
	schema := getResourceSchema(t, r)

	state := makeState(t, schema, &CollectionResourceModel{
		Collection: types.StringValue("articles"),
		Hidden:     types.BoolValue(false),
		Singleton:  types.BoolValue(false),
	})

	resp := &fwresource.DeleteResponse{State: tfsdk.State{Schema: schema}}
	r.Delete(context.Background(), fwresource.DeleteRequest{State: state}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ===========================================================================
// RolePoliciesAttachment Resource CRUD tests
// ===========================================================================

// makeSetValue creates a types.Set of strings from a slice.
func makeSetValue(t *testing.T, vals []string) types.Set {
	t.Helper()
	elements := make([]attr.Value, len(vals))
	for i, v := range vals {
		elements[i] = types.StringValue(v)
	}
	s, diags := types.SetValue(types.StringType, elements)
	require.False(t, diags.HasError())
	return s
}

func TestRolePoliciesAttachment_Create_Full(t *testing.T) {
	callCount := 0

	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		callCount++
		switch {
		// 1st call: GET to read existing policies
		case callCount == 1 && req.Method == "GET":
			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{
					"id":       "role-1",
					"policies": []interface{}{},
				},
			}), nil

		// 2nd call: PATCH to add policies
		case callCount == 2 && req.Method == "PATCH":
			var body map[string]interface{}
			bodyBytes, _ := io.ReadAll(req.Body)
			json.Unmarshal(bodyBytes, &body)
			policies := body["policies"].(map[string]interface{})
			creates := policies["create"].([]interface{})
			assert.Len(t, creates, 2)

			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{"id": "role-1"},
			}), nil
		}

		t.Fatalf("unexpected call #%d: %s %s", callCount, req.Method, req.URL)
		return nil, nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	schema := getResourceSchema(t, r)

	plan := makePlan(t, schema, &RolePoliciesAttachmentModel{
		ID:        types.StringUnknown(),
		RoleID:    types.StringValue("role-1"),
		PolicyIDs: makeSetValue(t, []string{"policy-a", "policy-b"}),
	})

	resp := &fwresource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(context.Background(), fwresource.CreateRequest{Plan: plan}, resp)

	require.False(t, resp.Diagnostics.HasError(), "Create diagnostics: %v", resp.Diagnostics)

	var result RolePoliciesAttachmentModel
	resp.State.Get(context.Background(), &result)
	assert.Equal(t, "role-1", result.ID.ValueString())
	assert.Equal(t, "role-1", result.RoleID.ValueString())
}

func TestRolePoliciesAttachment_Create_NoChangesNeeded(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		// Only GET call - policies already match
		assert.Equal(t, "GET", req.Method)
		return mockJSONResponse(200, map[string]interface{}{
			"data": map[string]interface{}{
				"id": "role-1",
				"policies": []map[string]interface{}{
					{"id": "access-1", "policy": "policy-a"},
				},
			},
		}), nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	schema := getResourceSchema(t, r)

	plan := makePlan(t, schema, &RolePoliciesAttachmentModel{
		ID:        types.StringUnknown(),
		RoleID:    types.StringValue("role-1"),
		PolicyIDs: makeSetValue(t, []string{"policy-a"}),
	})

	resp := &fwresource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(context.Background(), fwresource.CreateRequest{Plan: plan}, resp)

	require.False(t, resp.Diagnostics.HasError())
}

func TestRolePoliciesAttachment_Create_Error(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockErrorResponse(404, "Role not found"), nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	schema := getResourceSchema(t, r)

	plan := makePlan(t, schema, &RolePoliciesAttachmentModel{
		ID:        types.StringUnknown(),
		RoleID:    types.StringValue("bad-role"),
		PolicyIDs: makeSetValue(t, []string{"policy-a"}),
	})

	resp := &fwresource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(context.Background(), fwresource.CreateRequest{Plan: plan}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestRolePoliciesAttachment_Read_Full(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "GET", req.Method)
		return mockJSONResponse(200, map[string]interface{}{
			"data": map[string]interface{}{
				"id": "role-1",
				"policies": []map[string]interface{}{
					{"id": "access-1", "policy": "policy-a"},
					{"id": "access-2", "policy": "policy-b"},
				},
			},
		}), nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	schema := getResourceSchema(t, r)

	state := makeState(t, schema, &RolePoliciesAttachmentModel{
		ID:        types.StringValue("role-1"),
		RoleID:    types.StringValue("role-1"),
		PolicyIDs: makeSetValue(t, []string{"policy-a"}),
	})

	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: schema}}
	r.Read(context.Background(), fwresource.ReadRequest{State: state}, resp)

	require.False(t, resp.Diagnostics.HasError())

	var result RolePoliciesAttachmentModel
	resp.State.Get(context.Background(), &result)
	assert.Equal(t, "role-1", result.ID.ValueString())

	var policyIDs []string
	result.PolicyIDs.ElementsAs(context.Background(), &policyIDs, false)
	assert.Len(t, policyIDs, 2)
	assert.Contains(t, policyIDs, "policy-a")
	assert.Contains(t, policyIDs, "policy-b")
}

func TestRolePoliciesAttachment_Read_Error(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockErrorResponse(404, "Not found"), nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	schema := getResourceSchema(t, r)

	state := makeState(t, schema, &RolePoliciesAttachmentModel{
		ID:        types.StringValue("bad-role"),
		RoleID:    types.StringValue("bad-role"),
		PolicyIDs: makeSetValue(t, []string{"policy-a"}),
	})

	resp := &fwresource.ReadResponse{State: tfsdk.State{Schema: schema}}
	r.Read(context.Background(), fwresource.ReadRequest{State: state}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestRolePoliciesAttachment_Update_Full(t *testing.T) {
	callCount := 0

	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		callCount++
		switch {
		// 1st call: GET existing policies
		case callCount == 1 && req.Method == "GET":
			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{
					"id": "role-1",
					"policies": []map[string]interface{}{
						{"id": "access-1", "policy": "policy-a"},
						{"id": "access-2", "policy": "policy-b"},
					},
				},
			}), nil

		// 2nd call: PATCH to add new and remove old
		case callCount == 2 && req.Method == "PATCH":
			var body map[string]interface{}
			bodyBytes, _ := io.ReadAll(req.Body)
			json.Unmarshal(bodyBytes, &body)
			policies := body["policies"].(map[string]interface{})

			// Should create policy-c
			creates := policies["create"].([]interface{})
			assert.Len(t, creates, 1)

			// Should delete access-1 (policy-a)
			deletes := policies["delete"].([]interface{})
			assert.Len(t, deletes, 1)

			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{"id": "role-1"},
			}), nil
		}

		t.Fatalf("unexpected call #%d: %s %s", callCount, req.Method, req.URL)
		return nil, nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	schema := getResourceSchema(t, r)

	// Desired: keep policy-b, add policy-c, remove policy-a
	plan := makePlan(t, schema, &RolePoliciesAttachmentModel{
		ID:        types.StringValue("role-1"),
		RoleID:    types.StringValue("role-1"),
		PolicyIDs: makeSetValue(t, []string{"policy-b", "policy-c"}),
	})

	resp := &fwresource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(context.Background(), fwresource.UpdateRequest{Plan: plan}, resp)

	require.False(t, resp.Diagnostics.HasError())
}

func TestRolePoliciesAttachment_Update_Error(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockErrorResponse(403, "Forbidden"), nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	schema := getResourceSchema(t, r)

	plan := makePlan(t, schema, &RolePoliciesAttachmentModel{
		ID:        types.StringValue("role-1"),
		RoleID:    types.StringValue("role-1"),
		PolicyIDs: makeSetValue(t, []string{"policy-a"}),
	})

	resp := &fwresource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(context.Background(), fwresource.UpdateRequest{Plan: plan}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestRolePoliciesAttachment_Delete_Full(t *testing.T) {
	callCount := 0

	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		callCount++
		switch {
		case callCount == 1 && req.Method == "GET":
			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{
					"id": "role-1",
					"policies": []map[string]interface{}{
						{"id": "access-1", "policy": "policy-a"},
						{"id": "access-2", "policy": "policy-b"},
					},
				},
			}), nil

		case callCount == 2 && req.Method == "PATCH":
			var body map[string]interface{}
			bodyBytes, _ := io.ReadAll(req.Body)
			json.Unmarshal(bodyBytes, &body)
			policies := body["policies"].(map[string]interface{})
			deletes := policies["delete"].([]interface{})
			assert.Len(t, deletes, 2)

			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{"id": "role-1"},
			}), nil
		}

		t.Fatalf("unexpected call #%d: %s %s", callCount, req.Method, req.URL)
		return nil, nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	schema := getResourceSchema(t, r)

	state := makeState(t, schema, &RolePoliciesAttachmentModel{
		ID:        types.StringValue("role-1"),
		RoleID:    types.StringValue("role-1"),
		PolicyIDs: makeSetValue(t, []string{"policy-a", "policy-b"}),
	})

	resp := &fwresource.DeleteResponse{State: tfsdk.State{Schema: schema}}
	r.Delete(context.Background(), fwresource.DeleteRequest{State: state}, resp)

	require.False(t, resp.Diagnostics.HasError())
}

func TestRolePoliciesAttachment_Delete_EmptyPolicies(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		// Only GET - returns empty policies, so no PATCH needed
		return mockJSONResponse(200, map[string]interface{}{
			"data": map[string]interface{}{
				"id":       "role-1",
				"policies": []interface{}{},
			},
		}), nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	schema := getResourceSchema(t, r)

	state := makeState(t, schema, &RolePoliciesAttachmentModel{
		ID:        types.StringValue("role-1"),
		RoleID:    types.StringValue("role-1"),
		PolicyIDs: makeSetValue(t, []string{}),
	})

	resp := &fwresource.DeleteResponse{State: tfsdk.State{Schema: schema}}
	r.Delete(context.Background(), fwresource.DeleteRequest{State: state}, resp)

	require.False(t, resp.Diagnostics.HasError())
}

// Test Role Delete with children (triggers warning branch)
func TestRoleResource_Delete_WithChildren(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "DELETE", req.Method)
		return mockJSONResponse(204, nil), nil
	})

	r := &RoleResource{client: mockClient}
	schema := getResourceSchema(t, r)

	childList := stringListOrNull([]string{"child-1", "child-2"})

	state := makeState(t, schema, &RoleResourceModel{
		ID:       types.StringValue("role-uuid"),
		Name:     types.StringValue("Parent Role"),
		Children: childList,
		Users:    types.ListNull(types.StringType),
	})

	resp := &fwresource.DeleteResponse{State: tfsdk.State{Schema: schema}}
	r.Delete(context.Background(), fwresource.DeleteRequest{State: state}, resp)

	// Should succeed but with a warning about children
	require.False(t, resp.Diagnostics.HasError())
	assert.True(t, resp.Diagnostics.WarningsCount() > 0, "should have a warning about children")
}

// makeEmptyState creates a State with proper initialization for ImportState tests.
func makeEmptyState(t *testing.T, schema rschema.Schema) tfsdk.State {
	t.Helper()
	// Initialize with a null model so SetAttribute works
	state := tfsdk.State{Schema: schema}
	state.Set(context.Background(), &RolePoliciesAttachmentModel{
		ID:        types.StringNull(),
		RoleID:    types.StringNull(),
		PolicyIDs: types.SetNull(types.StringType),
	})
	return state
}

// Test RolePoliciesAttachment ImportState with custom logic
func TestRolePoliciesAttachment_ImportState(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockJSONResponse(200, map[string]interface{}{
			"data": map[string]interface{}{
				"id": "role-1",
				"policies": []map[string]interface{}{
					{"id": "access-1", "policy": "policy-a"},
					{"id": "access-2", "policy": "policy-b"},
				},
			},
		}), nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	schema := getResourceSchema(t, r)

	initialState := makeEmptyState(t, schema)
	resp := &fwresource.ImportStateResponse{State: initialState}
	r.ImportState(context.Background(), fwresource.ImportStateRequest{ID: "role-1"}, resp)

	require.False(t, resp.Diagnostics.HasError(), "ImportState diagnostics: %v", resp.Diagnostics)

	var result RolePoliciesAttachmentModel
	resp.State.Get(context.Background(), &result)
	assert.Equal(t, "role-1", result.ID.ValueString())
	assert.Equal(t, "role-1", result.RoleID.ValueString())

	var policyIDs []string
	result.PolicyIDs.ElementsAs(context.Background(), &policyIDs, false)
	assert.Len(t, policyIDs, 2)
}

func TestRolePoliciesAttachment_ImportState_EmptyID(t *testing.T) {
	r := &RolePoliciesAttachmentResource{}
	schema := getResourceSchema(t, r)

	initialState := makeEmptyState(t, schema)
	resp := &fwresource.ImportStateResponse{State: initialState}
	r.ImportState(context.Background(), fwresource.ImportStateRequest{ID: ""}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestRolePoliciesAttachment_ImportState_Error(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockErrorResponse(404, "Role not found"), nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	schema := getResourceSchema(t, r)

	initialState := makeEmptyState(t, schema)
	resp := &fwresource.ImportStateResponse{State: initialState}
	r.ImportState(context.Background(), fwresource.ImportStateRequest{ID: "bad-role"}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestRolePoliciesAttachment_Delete_Error(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockErrorResponse(500, "Server error"), nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	schema := getResourceSchema(t, r)

	state := makeState(t, schema, &RolePoliciesAttachmentModel{
		ID:        types.StringValue("role-1"),
		RoleID:    types.StringValue("role-1"),
		PolicyIDs: makeSetValue(t, []string{"policy-a"}),
	})

	resp := &fwresource.DeleteResponse{State: tfsdk.State{Schema: schema}}
	r.Delete(context.Background(), fwresource.DeleteRequest{State: state}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}
