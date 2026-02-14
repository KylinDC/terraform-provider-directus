package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Schema & Metadata
// ---------------------------------------------------------------------------

func TestRoleResourceSchema(t *testing.T) {
	r := &RoleResource{}
	schemaResp := fwresource.SchemaResponse{}
	r.Schema(context.Background(), fwresource.SchemaRequest{}, &schemaResp)

	require.False(t, schemaResp.Diagnostics.HasError())

	expectedAttrs := []string{"id", "name", "icon", "description", "parent", "children", "users"}
	for _, attr := range expectedAttrs {
		assert.NotNil(t, schemaResp.Schema.Attributes[attr], "%s attribute should exist", attr)
	}

	// admin_access and app_access are policy-level fields, not role-level
	assert.Nil(t, schemaResp.Schema.Attributes["admin_access"])
	assert.Nil(t, schemaResp.Schema.Attributes["app_access"])
}

func TestRoleResourceMetadata(t *testing.T) {
	r := &RoleResource{}
	metadataResp := &fwresource.MetadataResponse{}
	r.Metadata(context.Background(), fwresource.MetadataRequest{ProviderTypeName: "directus"}, metadataResp)

	assert.Equal(t, "directus_role", metadataResp.TypeName)
}

// ---------------------------------------------------------------------------
// buildRoleInput
// ---------------------------------------------------------------------------

func TestBuildRoleCreateInput(t *testing.T) {
	t.Run("minimal fields", func(t *testing.T) {
		input := buildRoleInput(RoleResourceModel{
			Name: types.StringValue("Test Role"),
		}, true)

		assert.Equal(t, "Test Role", input["name"])
		assert.NotContains(t, input, "icon")
		assert.NotContains(t, input, "description")
		assert.NotContains(t, input, "parent")
	})

	t.Run("all fields", func(t *testing.T) {
		input := buildRoleInput(RoleResourceModel{
			Name:        types.StringValue("Test Role"),
			Icon:        types.StringValue("person"),
			Description: types.StringValue("Test description"),
			Parent:      types.StringValue("parent-uuid"),
		}, true)

		assert.Equal(t, "Test Role", input["name"])
		assert.Equal(t, "person", input["icon"])
		assert.Equal(t, "Test description", input["description"])
		assert.Equal(t, "parent-uuid", input["parent"])
	})
}

func TestBuildRoleUpdateInput(t *testing.T) {
	t.Run("partial update", func(t *testing.T) {
		input := buildRoleInput(RoleResourceModel{
			Name: types.StringValue("Updated Role"),
			Icon: types.StringNull(),
		}, false)

		assert.Equal(t, "Updated Role", input["name"])
		assert.NotContains(t, input, "icon")
	})

	t.Run("update parent", func(t *testing.T) {
		input := buildRoleInput(RoleResourceModel{
			Name:   types.StringValue("Updated Role"),
			Parent: types.StringValue("new-parent-uuid"),
		}, false)

		assert.Equal(t, "Updated Role", input["name"])
		assert.Equal(t, "new-parent-uuid", input["parent"])
	})
}

// ---------------------------------------------------------------------------
// roleAPIResponse.toModel
// ---------------------------------------------------------------------------

func TestRoleAPIResponseToModel(t *testing.T) {
	t.Run("minimal response", func(t *testing.T) {
		model := (&roleAPIResponse{ID: "test-uuid", Name: "Test Role"}).toModel()

		assert.Equal(t, "test-uuid", model.ID.ValueString())
		assert.Equal(t, "Test Role", model.Name.ValueString())
		assert.True(t, model.Icon.IsNull())
		assert.True(t, model.Description.IsNull())
		assert.True(t, model.Parent.IsNull())
		assert.True(t, model.Children.IsNull())
		assert.True(t, model.Users.IsNull())
	})

	t.Run("full response", func(t *testing.T) {
		model := (&roleAPIResponse{
			ID: "test-uuid", Name: "Test Role", Icon: "person",
			Description: "Test description", Parent: "parent-uuid",
			Children: []string{"child1-uuid", "child2-uuid"},
			Users:    []string{"user1-uuid"},
		}).toModel()

		assert.Equal(t, "test-uuid", model.ID.ValueString())
		assert.Equal(t, "person", model.Icon.ValueString())
		assert.Equal(t, "parent-uuid", model.Parent.ValueString())

		var children []string
		model.Children.ElementsAs(context.Background(), &children, false)
		assert.Equal(t, []string{"child1-uuid", "child2-uuid"}, children)

		var users []string
		model.Users.ElementsAs(context.Background(), &users, false)
		assert.Equal(t, []string{"user1-uuid"}, users)
	})
}

// ---------------------------------------------------------------------------
// CRUD mock tests
// ---------------------------------------------------------------------------

func TestRoleResource_Create(t *testing.T) {
	t.Run("success with minimal fields", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "POST", req.Method)
			assert.Equal(t, "http://example.com/roles", req.URL.String())
			assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))

			var reqBody map[string]interface{}
			bodyBytes, _ := io.ReadAll(req.Body)
			json.Unmarshal(bodyBytes, &reqBody)
			assert.Equal(t, "Test Role", reqBody["name"])

			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{"id": "role-uuid", "name": "Test Role"},
			}), nil
		})

		r := &RoleResource{client: mockClient}
		input := buildRoleInput(RoleResourceModel{Name: types.StringValue("Test Role")}, true)

		var result struct {
			Data roleAPIResponse `json:"data"`
		}
		err := r.client.Create(context.Background(), "roles", input, &result)

		require.NoError(t, err)
		assert.Equal(t, "role-uuid", result.Data.ID)
		assert.Equal(t, "Test Role", result.Data.Name)
	})

	t.Run("success with parent", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			var reqBody map[string]interface{}
			bodyBytes, _ := io.ReadAll(req.Body)
			json.Unmarshal(bodyBytes, &reqBody)
			assert.Equal(t, "Child Role", reqBody["name"])
			assert.Equal(t, "parent-uuid", reqBody["parent"])

			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{"id": "child-uuid", "name": "Child Role", "parent": "parent-uuid"},
			}), nil
		})

		r := &RoleResource{client: mockClient}
		input := buildRoleInput(RoleResourceModel{
			Name:   types.StringValue("Child Role"),
			Parent: types.StringValue("parent-uuid"),
		}, true)

		var result struct {
			Data roleAPIResponse `json:"data"`
		}
		err := r.client.Create(context.Background(), "roles", input, &result)

		require.NoError(t, err)
		assert.Equal(t, "child-uuid", result.Data.ID)
		assert.Equal(t, "parent-uuid", result.Data.Parent)
	})

	t.Run("API error", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			return mockErrorResponse(400, "Bad request"), nil
		})

		r := &RoleResource{client: mockClient}
		input := buildRoleInput(RoleResourceModel{Name: types.StringValue("Bad")}, true)

		var result struct {
			Data roleAPIResponse `json:"data"`
		}
		err := r.client.Create(context.Background(), "roles", input, &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "400")
	})
}

func TestRoleResource_Read(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "GET", req.Method)
			assert.Equal(t, "http://example.com/roles/role-uuid", req.URL.String())

			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{
					"id": "role-uuid", "name": "Test Role",
					"icon": "person", "description": "Test description",
					"children": []string{"child1-uuid"},
				},
			}), nil
		})

		r := &RoleResource{client: mockClient}
		var result struct {
			Data roleAPIResponse `json:"data"`
		}
		err := r.client.Get(context.Background(), "roles", "role-uuid", &result)

		require.NoError(t, err)
		assert.Equal(t, "role-uuid", result.Data.ID)
		assert.Equal(t, "Test Role", result.Data.Name)
		assert.Equal(t, "person", result.Data.Icon)
		assert.Equal(t, []string{"child1-uuid"}, result.Data.Children)
	})

	t.Run("not found", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			return mockErrorResponse(404, "Item not found"), nil
		})

		r := &RoleResource{client: mockClient}
		var result struct {
			Data roleAPIResponse `json:"data"`
		}
		err := r.client.Get(context.Background(), "roles", "nonexistent-id", &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})
}

func TestRoleResource_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "PATCH", req.Method)

			var reqBody map[string]interface{}
			bodyBytes, _ := io.ReadAll(req.Body)
			json.Unmarshal(bodyBytes, &reqBody)
			assert.Equal(t, "Updated Role", reqBody["name"])

			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{"id": "role-uuid", "name": "Updated Role"},
			}), nil
		})

		r := &RoleResource{client: mockClient}
		input := buildRoleInput(RoleResourceModel{
			ID:   types.StringValue("role-uuid"),
			Name: types.StringValue("Updated Role"),
		}, false)

		var result struct {
			Data roleAPIResponse `json:"data"`
		}
		err := r.client.Update(context.Background(), "roles", "role-uuid", input, &result)

		require.NoError(t, err)
		assert.Equal(t, "Updated Role", result.Data.Name)
	})
}

func TestRoleResource_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "DELETE", req.Method)
			assert.Equal(t, "http://example.com/roles/role-uuid", req.URL.String())

			return mockJSONResponse(204, nil), nil
		})

		r := &RoleResource{client: mockClient}
		err := r.client.Delete(context.Background(), "roles", "role-uuid")
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			return mockErrorResponse(404, "Item not found"), nil
		})

		r := &RoleResource{client: mockClient}
		err := r.client.Delete(context.Background(), "roles", "nonexistent-id")
		require.Error(t, err)
	})
}
