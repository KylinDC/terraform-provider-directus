package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kylindc/terraform-provider-directus/internal/client"
)

// ---------------------------------------------------------------------------
// Mock infrastructure (shared across all provider test files)
// ---------------------------------------------------------------------------

// mockTransport implements http.RoundTripper for mocking HTTP calls.
type mockTransport struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

// newMockClient creates a *client.Client backed by a mock HTTP transport.
func newMockClient(doFunc func(req *http.Request) (*http.Response, error)) *client.Client {
	return &client.Client{
		BaseURL: "http://example.com",
		HTTPClient: &http.Client{
			Transport: &mockTransport{doFunc: doFunc},
		},
		Token: "test-token",
	}
}

// mockJSONResponse is a helper that builds a successful JSON HTTP response.
func mockJSONResponse(statusCode int, body interface{}) *http.Response {
	data, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader(data)),
		Header:     make(http.Header),
	}
}

// mockErrorResponse is a helper that builds a Directus-style error HTTP response.
func mockErrorResponse(statusCode int, message string) *http.Response {
	body := map[string]interface{}{
		"errors": []map[string]interface{}{
			{"message": message, "extensions": map[string]interface{}{"code": "FORBIDDEN"}},
		},
	}
	return mockJSONResponse(statusCode, body)
}

// ---------------------------------------------------------------------------
// Schema & Metadata
// ---------------------------------------------------------------------------

func TestPolicyResource_Schema(t *testing.T) {
	res := &PolicyResource{}
	schemaResp := &fwresource.SchemaResponse{}
	res.Schema(context.Background(), fwresource.SchemaRequest{}, schemaResp)

	require.False(t, schemaResp.Diagnostics.HasError())

	expectedAttrs := []string{"id", "name", "icon", "description", "ip_access", "enforce_tfa", "admin_access", "app_access"}
	for _, attr := range expectedAttrs {
		assert.NotNil(t, schemaResp.Schema.Attributes[attr], "%s attribute should exist", attr)
	}
}

func TestPolicyResource_Metadata(t *testing.T) {
	res := &PolicyResource{}
	metadataResp := &fwresource.MetadataResponse{}
	res.Metadata(context.Background(), fwresource.MetadataRequest{ProviderTypeName: "directus"}, metadataResp)

	assert.Equal(t, "directus_policy", metadataResp.TypeName)
}

// ---------------------------------------------------------------------------
// policyAPIResponse.toModel
// ---------------------------------------------------------------------------

func TestPolicyAPIResponseToModel(t *testing.T) {
	t.Run("minimal response", func(t *testing.T) {
		resp := &policyAPIResponse{
			ID:   "uuid-1",
			Name: "My Policy",
		}
		model := resp.toModel()

		assert.Equal(t, "uuid-1", model.ID.ValueString())
		assert.Equal(t, "My Policy", model.Name.ValueString())
		assert.True(t, model.Icon.IsNull())
		assert.True(t, model.Description.IsNull())
		assert.True(t, model.IPAccess.IsNull())
		assert.False(t, model.EnforceTFA.ValueBool())
		assert.False(t, model.AdminAccess.ValueBool())
		assert.False(t, model.AppAccess.ValueBool())
	})

	t.Run("full response", func(t *testing.T) {
		resp := &policyAPIResponse{
			ID:          "uuid-2",
			Name:        "Admin Policy",
			Icon:        "shield",
			Description: "Full admin access",
			IPAccess:    []string{"192.168.1.0/24"},
			EnforceTFA:  true,
			AdminAccess: true,
			AppAccess:   true,
		}
		model := resp.toModel()

		assert.Equal(t, "uuid-2", model.ID.ValueString())
		assert.Equal(t, "Admin Policy", model.Name.ValueString())
		assert.Equal(t, "shield", model.Icon.ValueString())
		assert.Equal(t, "Full admin access", model.Description.ValueString())
		assert.Equal(t, "192.168.1.0/24", model.IPAccess.ValueString())
		assert.True(t, model.EnforceTFA.ValueBool())
		assert.True(t, model.AdminAccess.ValueBool())
		assert.True(t, model.AppAccess.ValueBool())
	})

	t.Run("partial optional fields", func(t *testing.T) {
		resp := &policyAPIResponse{
			ID:        "uuid-3",
			Name:      "App Only",
			AppAccess: true,
		}
		model := resp.toModel()

		assert.Equal(t, "uuid-3", model.ID.ValueString())
		assert.True(t, model.Icon.IsNull())
		assert.True(t, model.Description.IsNull())
		assert.True(t, model.IPAccess.IsNull())
		assert.False(t, model.EnforceTFA.ValueBool())
		assert.False(t, model.AdminAccess.ValueBool())
		assert.True(t, model.AppAccess.ValueBool())
	})
}

// ---------------------------------------------------------------------------
// CRUD mock tests
// ---------------------------------------------------------------------------

func TestPolicyResource_Create(t *testing.T) {
	t.Run("success with minimal fields", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "POST", req.Method)
			assert.Equal(t, "http://example.com/policies", req.URL.String())
			assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))

			var reqBody map[string]interface{}
			bodyBytes, _ := io.ReadAll(req.Body)
			json.Unmarshal(bodyBytes, &reqBody)
			assert.Equal(t, "New Policy", reqBody["name"])

			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "new-uuid",
					"name": "New Policy",
				},
			}), nil
		})

		r := &PolicyResource{client: mockClient}

		reqBody := map[string]interface{}{
			"name": "New Policy",
		}

		var result struct {
			Data policyAPIResponse `json:"data"`
		}

		err := r.client.Create(context.Background(), "policies", reqBody, &result)

		require.NoError(t, err)
		assert.Equal(t, "new-uuid", result.Data.ID)
		assert.Equal(t, "New Policy", result.Data.Name)
	})

	t.Run("success with all fields", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			var reqBody map[string]interface{}
			bodyBytes, _ := io.ReadAll(req.Body)
			json.Unmarshal(bodyBytes, &reqBody)

			assert.Equal(t, "Admin Policy", reqBody["name"])
			assert.Equal(t, "shield", reqBody["icon"])
			assert.Equal(t, "Full access", reqBody["description"])
			assert.Equal(t, true, reqBody["admin_access"])
			assert.Equal(t, true, reqBody["app_access"])

			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{
					"id":           "admin-uuid",
					"name":         "Admin Policy",
					"icon":         "shield",
					"description":  "Full access",
					"admin_access": true,
					"app_access":   true,
				},
			}), nil
		})

		r := &PolicyResource{client: mockClient}

		reqBody := map[string]interface{}{
			"name":         "Admin Policy",
			"icon":         "shield",
			"description":  "Full access",
			"admin_access": true,
			"app_access":   true,
		}

		var result struct {
			Data policyAPIResponse `json:"data"`
		}

		err := r.client.Create(context.Background(), "policies", reqBody, &result)

		require.NoError(t, err)
		assert.Equal(t, "admin-uuid", result.Data.ID)
		assert.True(t, result.Data.AdminAccess)
		assert.True(t, result.Data.AppAccess)
	})

	t.Run("API error", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			return mockErrorResponse(403, "Forbidden"), nil
		})

		r := &PolicyResource{client: mockClient}

		reqBody := map[string]interface{}{"name": "Test"}
		var result struct {
			Data policyAPIResponse `json:"data"`
		}

		err := r.client.Create(context.Background(), "policies", reqBody, &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "403")
	})
}

func TestPolicyResource_Read(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "GET", req.Method)
			assert.Equal(t, "http://example.com/policies/uuid-1", req.URL.String())

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

		var result struct {
			Data policyAPIResponse `json:"data"`
		}

		err := r.client.Get(context.Background(), "policies", "uuid-1", &result)

		require.NoError(t, err)
		assert.Equal(t, "uuid-1", result.Data.ID)
		assert.Equal(t, "Test Policy", result.Data.Name)
		assert.Equal(t, "lock", result.Data.Icon)
		assert.True(t, result.Data.EnforceTFA)
	})

	t.Run("not found", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			return mockErrorResponse(404, "Item not found"), nil
		})

		r := &PolicyResource{client: mockClient}

		var result struct {
			Data policyAPIResponse `json:"data"`
		}

		err := r.client.Get(context.Background(), "policies", "nonexistent", &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})
}

func TestPolicyResource_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "PATCH", req.Method)
			assert.Equal(t, "http://example.com/policies/uuid-1", req.URL.String())

			var reqBody map[string]interface{}
			bodyBytes, _ := io.ReadAll(req.Body)
			json.Unmarshal(bodyBytes, &reqBody)
			assert.Equal(t, "Updated Policy", reqBody["name"])

			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "uuid-1",
					"name": "Updated Policy",
				},
			}), nil
		})

		r := &PolicyResource{client: mockClient}

		reqBody := map[string]interface{}{
			"name": "Updated Policy",
		}

		var result struct {
			Data policyAPIResponse `json:"data"`
		}

		err := r.client.Update(context.Background(), "policies", "uuid-1", reqBody, &result)

		require.NoError(t, err)
		assert.Equal(t, "Updated Policy", result.Data.Name)
	})
}

func TestPolicyResource_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "DELETE", req.Method)
			assert.Equal(t, "http://example.com/policies/uuid-1", req.URL.String())

			return &http.Response{
				StatusCode: 204,
				Body:       io.NopCloser(bytes.NewReader([]byte{})),
				Header:     make(http.Header),
			}, nil
		})

		r := &PolicyResource{client: mockClient}

		err := r.client.Delete(context.Background(), "policies", "uuid-1")
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			return mockErrorResponse(404, "Item not found"), nil
		})

		r := &PolicyResource{client: mockClient}

		err := r.client.Delete(context.Background(), "policies", "nonexistent")
		require.Error(t, err)
	})
}

// ---------------------------------------------------------------------------
// Build request body from model (mirrors how Create/Update build their payloads)
// ---------------------------------------------------------------------------

func TestPolicyBuildRequestBody(t *testing.T) {
	t.Run("minimal create body", func(t *testing.T) {
		plan := PolicyResourceModel{
			Name: types.StringValue("Basic Policy"),
		}

		reqBody := map[string]interface{}{
			"name": plan.Name.ValueString(),
		}
		setStringField(reqBody, "icon", plan.Icon)
		setStringField(reqBody, "description", plan.Description)
		setIPAccessField(reqBody, plan.IPAccess)
		setBoolField(reqBody, "enforce_tfa", plan.EnforceTFA)
		setBoolField(reqBody, "admin_access", plan.AdminAccess)
		setBoolField(reqBody, "app_access", plan.AppAccess)

		assert.Equal(t, "Basic Policy", reqBody["name"])
		assert.NotContains(t, reqBody, "icon")
		assert.NotContains(t, reqBody, "description")
		assert.NotContains(t, reqBody, "ip_access")
		assert.NotContains(t, reqBody, "enforce_tfa")
		assert.NotContains(t, reqBody, "admin_access")
		assert.NotContains(t, reqBody, "app_access")
	})

	t.Run("full create body", func(t *testing.T) {
		plan := PolicyResourceModel{
			Name:        types.StringValue("Admin Policy"),
			Icon:        types.StringValue("shield"),
			Description: types.StringValue("Admin access"),
			IPAccess:    types.StringValue("10.0.0.0/8"),
			EnforceTFA:  types.BoolValue(true),
			AdminAccess: types.BoolValue(true),
			AppAccess:   types.BoolValue(true),
		}

		reqBody := map[string]interface{}{
			"name": plan.Name.ValueString(),
		}
		setStringField(reqBody, "icon", plan.Icon)
		setStringField(reqBody, "description", plan.Description)
		setIPAccessField(reqBody, plan.IPAccess)
		setBoolField(reqBody, "enforce_tfa", plan.EnforceTFA)
		setBoolField(reqBody, "admin_access", plan.AdminAccess)
		setBoolField(reqBody, "app_access", plan.AppAccess)

		assert.Equal(t, "Admin Policy", reqBody["name"])
		assert.Equal(t, "shield", reqBody["icon"])
		assert.Equal(t, "Admin access", reqBody["description"])
		assert.Equal(t, []string{"10.0.0.0/8"}, reqBody["ip_access"])
		assert.Equal(t, true, reqBody["enforce_tfa"])
		assert.Equal(t, true, reqBody["admin_access"])
		assert.Equal(t, true, reqBody["app_access"])
	})

	t.Run("update body with partial fields", func(t *testing.T) {
		plan := PolicyResourceModel{
			ID:   types.StringValue("uuid-1"),
			Name: types.StringValue("Updated"),
			Icon: types.StringNull(),
		}

		reqBody := make(map[string]interface{})
		setStringField(reqBody, "name", plan.Name)
		setStringField(reqBody, "icon", plan.Icon)
		setBoolField(reqBody, "enforce_tfa", plan.EnforceTFA)

		assert.Equal(t, "Updated", reqBody["name"])
		assert.NotContains(t, reqBody, "icon")
		assert.NotContains(t, reqBody, "enforce_tfa")
	})
}
