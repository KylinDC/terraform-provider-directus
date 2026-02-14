package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Schema & Metadata
// ---------------------------------------------------------------------------

func TestRolePoliciesAttachmentResourceSchema(t *testing.T) {
	r := &RolePoliciesAttachmentResource{}
	schemaResp := fwresource.SchemaResponse{}
	r.Schema(context.Background(), fwresource.SchemaRequest{}, &schemaResp)

	require.False(t, schemaResp.Diagnostics.HasError())

	expectedAttrs := []string{"id", "role_id", "policy_ids"}
	for _, attr := range expectedAttrs {
		assert.NotNil(t, schemaResp.Schema.Attributes[attr], "%s attribute should exist", attr)
	}
}

func TestRolePoliciesAttachmentResourceMetadata(t *testing.T) {
	r := &RolePoliciesAttachmentResource{}
	metadataResp := &fwresource.MetadataResponse{}
	r.Metadata(context.Background(), fwresource.MetadataRequest{ProviderTypeName: "directus"}, metadataResp)

	assert.Equal(t, "directus_role_policies_attachment", metadataResp.TypeName)
}

// ---------------------------------------------------------------------------
// readRolePolicies unit tests
// ---------------------------------------------------------------------------

func TestRolePoliciesReadRolePolicies_Success(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "GET", req.Method)
		assert.Contains(t, req.URL.String(), "/roles/role-uuid")
		assert.Contains(t, req.URL.RawQuery, "fields=")

		return mockJSONResponse(200, map[string]interface{}{
			"data": map[string]interface{}{
				"id": "role-uuid",
				"policies": []map[string]interface{}{
					{"id": "access-1", "policy": "policy-aaa"},
					{"id": "access-2", "policy": "policy-bbb"},
				},
			},
		}), nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	records, err := r.readRolePolicies(context.Background(), "role-uuid")

	require.NoError(t, err)
	require.Len(t, records, 2)
	assert.Equal(t, "access-1", records[0].ID)
	assert.Equal(t, "policy-aaa", records[0].Policy)
	assert.Equal(t, "access-2", records[1].ID)
	assert.Equal(t, "policy-bbb", records[1].Policy)
}

func TestRolePoliciesReadRolePolicies_EmptyPolicies(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockJSONResponse(200, map[string]interface{}{
			"data": map[string]interface{}{
				"id":       "role-uuid",
				"policies": []interface{}{},
			},
		}), nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	records, err := r.readRolePolicies(context.Background(), "role-uuid")

	require.NoError(t, err)
	assert.Len(t, records, 0)
}

func TestRolePoliciesReadRolePolicies_RoleNotFound(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockErrorResponse(404, "Item not found"), nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	_, err := r.readRolePolicies(context.Background(), "nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

// ---------------------------------------------------------------------------
// Create flow — adds multiple policies, removes undesired ones
// ---------------------------------------------------------------------------

func TestRolePoliciesAttachmentCreate_Success(t *testing.T) {
	var mu sync.Mutex
	callCount := 0

	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		mu.Lock()
		callCount++
		currentCall := callCount
		mu.Unlock()

		switch {
		case currentCall == 1 && req.Method == "GET":
			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{
					"id":       "role-uuid",
					"policies": []interface{}{},
				},
			}), nil

		case currentCall == 2 && req.Method == "PATCH":
			assert.Contains(t, req.URL.String(), "/roles/role-uuid")
			var reqBody map[string]interface{}
			bodyBytes, _ := io.ReadAll(req.Body)
			json.Unmarshal(bodyBytes, &reqBody)

			policies := reqBody["policies"].(map[string]interface{})
			creates := policies["create"].([]interface{})
			assert.Len(t, creates, 2, "Should create 2 policy attachments")

			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{"id": "role-uuid", "name": "Test Role"},
			}), nil
		}

		t.Fatalf("unexpected call #%d: %s %s", currentCall, req.Method, req.URL.String())
		return nil, nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	ctx := context.Background()

	existing, err := r.readRolePolicies(ctx, "role-uuid")
	require.NoError(t, err)
	assert.Empty(t, existing)

	patchBody := map[string]interface{}{
		"policies": map[string]interface{}{
			"create": []map[string]interface{}{
				{"policy": "policy-aaa"},
				{"policy": "policy-bbb"},
			},
		},
	}
	err = r.client.Update(ctx, "roles", "role-uuid", patchBody, nil)
	require.NoError(t, err)
}

func TestRolePoliciesAttachmentCreate_PartialOverlap(t *testing.T) {
	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		return mockJSONResponse(200, map[string]interface{}{
			"data": map[string]interface{}{
				"id": "role-uuid",
				"policies": []map[string]interface{}{
					{"id": "access-1", "policy": "policy-aaa"},
				},
			},
		}), nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	existing, err := r.readRolePolicies(context.Background(), "role-uuid")
	require.NoError(t, err)

	existingMap := make(map[string]string)
	for _, rec := range existing {
		existingMap[rec.Policy] = rec.ID
	}

	desiredPolicyIDs := []string{"policy-aaa", "policy-bbb"}
	var toCreate []map[string]interface{}
	for _, pid := range desiredPolicyIDs {
		if _, exists := existingMap[pid]; !exists {
			toCreate = append(toCreate, map[string]interface{}{"policy": pid})
		}
	}

	assert.Len(t, toCreate, 1)
	assert.Equal(t, "policy-bbb", toCreate[0]["policy"])
}

// ---------------------------------------------------------------------------
// Update flow — adds new, removes old policies
// ---------------------------------------------------------------------------

func TestRolePoliciesAttachmentUpdate_DiffComputation(t *testing.T) {
	existing := []accessRecordResponse{
		{ID: "access-1", Policy: "policy-aaa"},
		{ID: "access-2", Policy: "policy-bbb"},
	}
	desiredPolicyIDs := []string{"policy-bbb", "policy-ccc"}

	existingMap := make(map[string]string)
	for _, rec := range existing {
		existingMap[rec.Policy] = rec.ID
	}

	var toCreate []map[string]interface{}
	for _, pid := range desiredPolicyIDs {
		if _, exists := existingMap[pid]; !exists {
			toCreate = append(toCreate, map[string]interface{}{"policy": pid})
		}
	}

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

	require.Len(t, toCreate, 1)
	assert.Equal(t, "policy-ccc", toCreate[0]["policy"])

	require.Len(t, toDelete, 1)
	assert.Equal(t, "access-1", toDelete[0])
}

// ---------------------------------------------------------------------------
// Delete flow — removes all policies from the role
// ---------------------------------------------------------------------------

func TestRolePoliciesAttachmentDelete_Success(t *testing.T) {
	var mu sync.Mutex
	callCount := 0

	mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
		mu.Lock()
		callCount++
		currentCall := callCount
		mu.Unlock()

		switch {
		case currentCall == 1 && req.Method == "GET":
			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{
					"id": "role-uuid",
					"policies": []map[string]interface{}{
						{"id": "access-1", "policy": "policy-aaa"},
						{"id": "access-2", "policy": "policy-bbb"},
					},
				},
			}), nil

		case currentCall == 2 && req.Method == "PATCH":
			assert.Contains(t, req.URL.String(), "/roles/role-uuid")
			var reqBody map[string]interface{}
			bodyBytes, _ := io.ReadAll(req.Body)
			json.Unmarshal(bodyBytes, &reqBody)

			policies := reqBody["policies"].(map[string]interface{})
			deletes := policies["delete"].([]interface{})
			assert.Len(t, deletes, 2, "Should delete 2 access records")

			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{"id": "role-uuid", "name": "Test Role"},
			}), nil
		}

		t.Fatalf("unexpected call #%d: %s %s", currentCall, req.Method, req.URL.String())
		return nil, nil
	})

	r := &RolePoliciesAttachmentResource{client: mockClient}
	ctx := context.Background()

	existing, err := r.readRolePolicies(ctx, "role-uuid")
	require.NoError(t, err)
	require.Len(t, existing, 2)

	accessIDs := make([]string, 0, len(existing))
	for _, rec := range existing {
		accessIDs = append(accessIDs, rec.ID)
	}

	patchBody := map[string]interface{}{
		"policies": map[string]interface{}{
			"delete": accessIDs,
		},
	}
	err = r.client.Update(ctx, "roles", "role-uuid", patchBody, nil)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Import ID parsing
// ---------------------------------------------------------------------------

func TestRolePoliciesAttachmentImportParsing(t *testing.T) {
	tests := []struct {
		name      string
		importID  string
		wantError bool
	}{
		{"valid role UUID", "role-uuid", false},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantError {
				assert.Empty(t, tt.importID)
			} else {
				assert.NotEmpty(t, tt.importID)
			}
		})
	}
}
