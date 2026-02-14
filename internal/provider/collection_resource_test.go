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

func TestCollectionResourceSchema(t *testing.T) {
	r := &CollectionResource{}
	schemaResp := fwresource.SchemaResponse{}
	r.Schema(context.Background(), fwresource.SchemaRequest{}, &schemaResp)

	require.False(t, schemaResp.Diagnostics.HasError())

	expectedAttrs := []string{"collection", "icon", "note", "hidden", "singleton", "sort_field", "archive_field", "color"}
	for _, attr := range expectedAttrs {
		assert.NotNil(t, schemaResp.Schema.Attributes[attr], "%s attribute should exist", attr)
	}
}

func TestCollectionResourceMetadata(t *testing.T) {
	r := &CollectionResource{}
	metadataResp := &fwresource.MetadataResponse{}
	r.Metadata(context.Background(), fwresource.MetadataRequest{ProviderTypeName: "directus"}, metadataResp)

	assert.Equal(t, "directus_collection", metadataResp.TypeName)
}

// ---------------------------------------------------------------------------
// buildCollectionInput
// ---------------------------------------------------------------------------

func TestBuildCollectionCreateInput(t *testing.T) {
	t.Run("minimal fields", func(t *testing.T) {
		input := buildCollectionInput(CollectionResourceModel{
			Collection: types.StringValue("test_collection"),
		}, true)

		assert.Equal(t, "test_collection", input["collection"])
		assert.Contains(t, input, "schema", "create input must include schema")
		assert.NotContains(t, input, "meta")
	})

	t.Run("with meta fields", func(t *testing.T) {
		input := buildCollectionInput(CollectionResourceModel{
			Collection: types.StringValue("test_collection"),
			Icon:       types.StringValue("list"),
			Note:       types.StringValue("Test collection"),
			Hidden:     types.BoolValue(false),
			Singleton:  types.BoolValue(true),
			Color:      types.StringValue("#6644FF"),
		}, true)

		assert.Equal(t, "test_collection", input["collection"])
		meta := input["meta"].(map[string]interface{})
		assert.Equal(t, "list", meta["icon"])
		assert.Equal(t, "Test collection", meta["note"])
		assert.Equal(t, false, meta["hidden"])
		assert.Equal(t, true, meta["singleton"])
		assert.Equal(t, "#6644FF", meta["color"])
	})

	t.Run("with sort and archive fields", func(t *testing.T) {
		input := buildCollectionInput(CollectionResourceModel{
			Collection: types.StringValue("test_collection"),
			SortField:  types.StringValue("sort_order"),
			Archive:    types.StringValue("status"),
		}, true)

		meta := input["meta"].(map[string]interface{})
		assert.Equal(t, "sort_order", meta["sort_field"])
		assert.Equal(t, "status", meta["archive_field"])
	})
}

func TestBuildCollectionUpdateInput(t *testing.T) {
	t.Run("partial update", func(t *testing.T) {
		input := buildCollectionInput(CollectionResourceModel{
			Collection: types.StringValue("test_collection"),
			Icon:       types.StringValue("new_icon"),
		}, false)

		assert.NotContains(t, input, "collection", "update should not include collection name")
		assert.NotContains(t, input, "schema", "update should not include schema")
		meta := input["meta"].(map[string]interface{})
		assert.Equal(t, "new_icon", meta["icon"])
	})

	t.Run("update multiple fields", func(t *testing.T) {
		input := buildCollectionInput(CollectionResourceModel{
			Collection: types.StringValue("test_collection"),
			Note:       types.StringValue("Updated note"),
			Hidden:     types.BoolValue(true),
			Color:      types.StringValue("#FF0000"),
		}, false)

		meta := input["meta"].(map[string]interface{})
		assert.Equal(t, "Updated note", meta["note"])
		assert.Equal(t, true, meta["hidden"])
		assert.Equal(t, "#FF0000", meta["color"])
	})
}

// ---------------------------------------------------------------------------
// collectionAPIResponse.toModel
// ---------------------------------------------------------------------------

func TestCollectionAPIResponseToModel(t *testing.T) {
	t.Run("minimal response (no meta)", func(t *testing.T) {
		model := (&collectionAPIResponse{Collection: "test_collection"}).toModel()

		assert.Equal(t, "test_collection", model.Collection.ValueString())
		assert.True(t, model.Icon.IsNull())
		assert.True(t, model.Note.IsNull())
		assert.False(t, model.Hidden.ValueBool())
		assert.False(t, model.Singleton.ValueBool())
	})

	t.Run("full response", func(t *testing.T) {
		model := (&collectionAPIResponse{
			Collection: "test_collection",
			Meta: &collectionMetaResponse{
				Icon: "list", Note: "Test collection", Hidden: true,
				Singleton: false, SortField: "sort_order",
				ArchiveField: "status", Color: "#6644FF",
			},
		}).toModel()

		assert.Equal(t, "test_collection", model.Collection.ValueString())
		assert.Equal(t, "list", model.Icon.ValueString())
		assert.Equal(t, "Test collection", model.Note.ValueString())
		assert.True(t, model.Hidden.ValueBool())
		assert.False(t, model.Singleton.ValueBool())
		assert.Equal(t, "sort_order", model.SortField.ValueString())
		assert.Equal(t, "status", model.Archive.ValueString())
		assert.Equal(t, "#6644FF", model.Color.ValueString())
	})

	t.Run("response with partial meta", func(t *testing.T) {
		model := (&collectionAPIResponse{
			Collection: "test_collection",
			Meta:       &collectionMetaResponse{Icon: "list", Hidden: false},
		}).toModel()

		assert.Equal(t, "list", model.Icon.ValueString())
		assert.True(t, model.Note.IsNull())
		assert.True(t, model.SortField.IsNull())
		assert.False(t, model.Hidden.ValueBool())
	})
}

// ---------------------------------------------------------------------------
// CRUD mock tests
// ---------------------------------------------------------------------------

func TestCollectionResource_Create(t *testing.T) {
	t.Run("success with minimal fields", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "POST", req.Method)
			assert.Equal(t, "http://example.com/collections", req.URL.String())
			assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))

			var reqBody map[string]interface{}
			bodyBytes, _ := io.ReadAll(req.Body)
			json.Unmarshal(bodyBytes, &reqBody)
			assert.Equal(t, "test_collection", reqBody["collection"])

			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{"collection": "test_collection"},
			}), nil
		})

		r := &CollectionResource{client: mockClient}
		input := buildCollectionInput(CollectionResourceModel{
			Collection: types.StringValue("test_collection"),
		}, true)

		var result struct {
			Data collectionAPIResponse `json:"data"`
		}
		err := r.client.Create(context.Background(), "collections", input, &result)

		require.NoError(t, err)
		assert.Equal(t, "test_collection", result.Data.Collection)
	})

	t.Run("success with meta fields", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			var reqBody map[string]interface{}
			bodyBytes, _ := io.ReadAll(req.Body)
			json.Unmarshal(bodyBytes, &reqBody)

			assert.Equal(t, "articles", reqBody["collection"])
			meta := reqBody["meta"].(map[string]interface{})
			assert.Equal(t, "article", meta["icon"])

			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{
					"collection": "articles",
					"meta":       map[string]interface{}{"icon": "article", "note": "Blog articles"},
				},
			}), nil
		})

		r := &CollectionResource{client: mockClient}
		input := buildCollectionInput(CollectionResourceModel{
			Collection: types.StringValue("articles"),
			Icon:       types.StringValue("article"),
			Note:       types.StringValue("Blog articles"),
		}, true)

		var result struct {
			Data collectionAPIResponse `json:"data"`
		}
		err := r.client.Create(context.Background(), "collections", input, &result)

		require.NoError(t, err)
		assert.Equal(t, "articles", result.Data.Collection)
		assert.Equal(t, "article", result.Data.Meta.Icon)
	})

	t.Run("API error", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			return mockErrorResponse(403, "Forbidden"), nil
		})

		r := &CollectionResource{client: mockClient}
		input := buildCollectionInput(CollectionResourceModel{
			Collection: types.StringValue("test"),
		}, true)

		var result struct {
			Data collectionAPIResponse `json:"data"`
		}
		err := r.client.Create(context.Background(), "collections", input, &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "403")
	})
}

func TestCollectionResource_Read(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "GET", req.Method)
			assert.Equal(t, "http://example.com/collections/test_collection", req.URL.String())

			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{
					"collection": "test_collection",
					"meta":       map[string]interface{}{"icon": "list", "note": "Test collection", "hidden": false},
				},
			}), nil
		})

		r := &CollectionResource{client: mockClient}
		var result struct {
			Data collectionAPIResponse `json:"data"`
		}
		err := r.client.Get(context.Background(), "collections", "test_collection", &result)

		require.NoError(t, err)
		assert.Equal(t, "test_collection", result.Data.Collection)
		assert.Equal(t, "list", result.Data.Meta.Icon)
	})

	t.Run("not found", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			return mockErrorResponse(404, "Collection not found"), nil
		})

		r := &CollectionResource{client: mockClient}
		var result struct {
			Data collectionAPIResponse `json:"data"`
		}
		err := r.client.Get(context.Background(), "collections", "nonexistent", &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})
}

func TestCollectionResource_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "PATCH", req.Method)

			var reqBody map[string]interface{}
			bodyBytes, _ := io.ReadAll(req.Body)
			json.Unmarshal(bodyBytes, &reqBody)
			meta := reqBody["meta"].(map[string]interface{})
			assert.Equal(t, "Updated note", meta["note"])

			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{
					"collection": "test_collection",
					"meta":       map[string]interface{}{"note": "Updated note"},
				},
			}), nil
		})

		r := &CollectionResource{client: mockClient}
		input := buildCollectionInput(CollectionResourceModel{
			Collection: types.StringValue("test_collection"),
			Note:       types.StringValue("Updated note"),
		}, false)

		var result struct {
			Data collectionAPIResponse `json:"data"`
		}
		err := r.client.Update(context.Background(), "collections", "test_collection", input, &result)

		require.NoError(t, err)
		assert.Equal(t, "Updated note", result.Data.Meta.Note)
	})

	t.Run("update visibility", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			var reqBody map[string]interface{}
			bodyBytes, _ := io.ReadAll(req.Body)
			json.Unmarshal(bodyBytes, &reqBody)
			meta := reqBody["meta"].(map[string]interface{})
			assert.Equal(t, true, meta["hidden"])

			return mockJSONResponse(200, map[string]interface{}{
				"data": map[string]interface{}{
					"collection": "test_collection",
					"meta":       map[string]interface{}{"hidden": true},
				},
			}), nil
		})

		r := &CollectionResource{client: mockClient}
		input := buildCollectionInput(CollectionResourceModel{
			Collection: types.StringValue("test_collection"),
			Hidden:     types.BoolValue(true),
		}, false)

		var result struct {
			Data collectionAPIResponse `json:"data"`
		}
		err := r.client.Update(context.Background(), "collections", "test_collection", input, &result)

		require.NoError(t, err)
		assert.True(t, result.Data.Meta.Hidden)
	})
}

func TestCollectionResource_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "DELETE", req.Method)
			assert.Equal(t, "http://example.com/collections/test_collection", req.URL.String())

			return mockJSONResponse(204, nil), nil
		})

		r := &CollectionResource{client: mockClient}
		err := r.client.Delete(context.Background(), "collections", "test_collection")
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mockClient := newMockClient(func(req *http.Request) (*http.Response, error) {
			return mockErrorResponse(404, "Collection not found"), nil
		})

		r := &CollectionResource{client: mockClient}
		err := r.client.Delete(context.Background(), "collections", "nonexistent")
		require.Error(t, err)
	})
}
