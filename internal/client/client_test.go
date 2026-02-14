package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// newTestClient creates a Client pointing at the given httptest.Server.
func newTestClient(server *httptest.Server) *Client {
	return &Client{
		BaseURL:    server.URL,
		Token:      "test-token",
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// offlineClient creates a Client that will never make real HTTP calls.
func offlineClient() *Client {
	return &Client{
		BaseURL:    "https://example.com",
		Token:      "test-token",
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// ---------------------------------------------------------------------------
// NewClient
// ---------------------------------------------------------------------------

func TestNewClient_WithToken(t *testing.T) {
	client, err := NewClient(context.Background(), Config{
		BaseURL: "https://example.com",
		Token:   "test-token",
		Timeout: 10 * time.Second,
	})

	require.NoError(t, err)
	assert.Equal(t, "test-token", client.Token)
	assert.Equal(t, "https://example.com", client.BaseURL)
	assert.Equal(t, 10*time.Second, client.HTTPClient.Timeout)
}

func TestNewClient_DefaultTimeout(t *testing.T) {
	client, err := NewClient(context.Background(), Config{
		BaseURL: "https://example.com",
		Token:   "test-token",
	})

	require.NoError(t, err)
	assert.Equal(t, 30*time.Second, client.HTTPClient.Timeout)
}

func TestNewClient_MissingBaseURL(t *testing.T) {
	_, err := NewClient(context.Background(), Config{Token: "test-token"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "base URL is required")
}

func TestNewClient_InvalidURL(t *testing.T) {
	_, err := NewClient(context.Background(), Config{
		BaseURL: "://invalid-url",
		Token:   "test-token",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid base URL")
}

func TestNewClient_MissingToken(t *testing.T) {
	_, err := NewClient(context.Background(), Config{BaseURL: "https://example.com"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token is required")
}

// ---------------------------------------------------------------------------
// buildCollectionPath
// ---------------------------------------------------------------------------

func TestBuildCollectionPath(t *testing.T) {
	c := &Client{}

	tests := []struct {
		name       string
		collection string
		id         string
		expected   string
	}{
		// System collections
		{"system collection with id", "roles", "uuid-1", "/roles/uuid-1"},
		{"system collection without id", "roles", "", "/roles"},
		{"policies with id", "policies", "uuid-2", "/policies/uuid-2"},
		{"policies without id", "policies", "", "/policies"},
		{"collections with id", "collections", "my_table", "/collections/my_table"},
		{"collections without id", "collections", "", "/collections"},
		{"users with id", "users", "user-1", "/users/user-1"},
		{"folders with id", "folders", "folder-1", "/folders/folder-1"},
		{"settings without id", "settings", "", "/settings"},

		// Custom collections (items prefix)
		{"custom collection with id", "articles", "1", "/items/articles/1"},
		{"custom collection without id", "articles", "", "/items/articles"},
		{"custom collection with uuid id", "products", "abc-123", "/items/products/abc-123"},
		{"custom underscore name", "my_custom_table", "42", "/items/my_custom_table/42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.buildCollectionPath(tt.collection, tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ---------------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------------

func TestGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/items/articles/1", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"id": "1", "title": "Test Article"},
		})
	}))
	defer server.Close()

	var result map[string]interface{}
	err := newTestClient(server).Get(context.Background(), "articles", "1", &result)

	require.NoError(t, err)
	data := result["data"].(map[string]interface{})
	assert.Equal(t, "1", data["id"])
}

func TestGet_MissingCollection(t *testing.T) {
	err := offlineClient().Get(context.Background(), "", "1", &map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collection is required")
}

func TestGet_MissingID(t *testing.T) {
	err := offlineClient().Get(context.Background(), "articles", "", &map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

func TestGet_SystemCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/policies/test-uuid", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"id": "test-uuid", "name": "Test Policy"},
		})
	}))
	defer server.Close()

	var result map[string]interface{}
	err := newTestClient(server).Get(context.Background(), "policies", "test-uuid", &result)

	require.NoError(t, err)
	data := result["data"].(map[string]interface{})
	assert.Equal(t, "test-uuid", data["id"])
}

// ---------------------------------------------------------------------------
// GetWithParams
// ---------------------------------------------------------------------------

func TestGetWithParams_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/roles/role-1", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "id,policies.id,policies.policy", r.URL.Query().Get("fields"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id": "role-1",
				"policies": []map[string]interface{}{
					{"id": "access-1", "policy": "policy-a"},
				},
			},
		})
	}))
	defer server.Close()

	params := map[string]string{"fields": "id,policies.id,policies.policy"}
	var result map[string]interface{}
	err := newTestClient(server).GetWithParams(context.Background(), "roles", "role-1", params, &result)

	require.NoError(t, err)
	data := result["data"].(map[string]interface{})
	assert.Equal(t, "role-1", data["id"])
}

func TestGetWithParams_NoParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/roles/role-1", r.URL.Path)
		assert.Empty(t, r.URL.RawQuery)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"id": "role-1"},
		})
	}))
	defer server.Close()

	var result map[string]interface{}
	err := newTestClient(server).GetWithParams(context.Background(), "roles", "role-1", nil, &result)
	require.NoError(t, err)
}

func TestGetWithParams_MissingCollection(t *testing.T) {
	err := offlineClient().GetWithParams(context.Background(), "", "1", nil, &map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collection is required")
}

func TestGetWithParams_MissingID(t *testing.T) {
	err := offlineClient().GetWithParams(context.Background(), "roles", "", nil, &map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/items/articles", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "1", "title": "Article 1"},
				{"id": "2", "title": "Article 2"},
			},
		})
	}))
	defer server.Close()

	var result map[string]interface{}
	err := newTestClient(server).List(context.Background(), "articles", &result)

	require.NoError(t, err)
	data := result["data"].([]interface{})
	assert.Len(t, data, 2)
}

func TestList_MissingCollection(t *testing.T) {
	err := offlineClient().List(context.Background(), "", &map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collection is required")
}

func TestList_SystemCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/roles", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "r-1", "name": "Admin"},
			},
		})
	}))
	defer server.Close()

	var result map[string]interface{}
	err := newTestClient(server).List(context.Background(), "roles", &result)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestCreate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/items/articles", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var payload map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "New Article", payload["title"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"id": "3", "title": "New Article"},
		})
	}))
	defer server.Close()

	var result map[string]interface{}
	err := newTestClient(server).Create(context.Background(), "articles", map[string]interface{}{"title": "New Article"}, &result)

	require.NoError(t, err)
	data := result["data"].(map[string]interface{})
	assert.Equal(t, "3", data["id"])
}

func TestCreate_MissingCollection(t *testing.T) {
	err := offlineClient().Create(context.Background(), "", map[string]interface{}{"k": "v"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collection is required")
}

func TestCreate_MissingData(t *testing.T) {
	err := offlineClient().Create(context.Background(), "articles", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "data is required")
}

func TestCreate_NilResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"id": "new-uuid"},
		})
	}))
	defer server.Close()

	err := newTestClient(server).Create(context.Background(), "policies", map[string]interface{}{"name": "P"}, nil)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestUpdate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/items/articles/1", r.URL.Path)
		assert.Equal(t, http.MethodPatch, r.Method)

		var payload map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "Updated Article", payload["title"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"id": "1", "title": "Updated Article"},
		})
	}))
	defer server.Close()

	var result map[string]interface{}
	err := newTestClient(server).Update(context.Background(), "articles", "1", map[string]interface{}{"title": "Updated Article"}, &result)

	require.NoError(t, err)
	data := result["data"].(map[string]interface{})
	assert.Equal(t, "Updated Article", data["title"])
}

func TestUpdate_MissingCollection(t *testing.T) {
	err := offlineClient().Update(context.Background(), "", "1", map[string]interface{}{"k": "v"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collection is required")
}

func TestUpdate_MissingID(t *testing.T) {
	err := offlineClient().Update(context.Background(), "articles", "", map[string]interface{}{"k": "v"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

func TestUpdate_MissingData(t *testing.T) {
	err := offlineClient().Update(context.Background(), "articles", "1", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "data is required")
}

func TestUpdate_NilResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"id": "uuid", "name": "Updated"},
		})
	}))
	defer server.Close()

	err := newTestClient(server).Update(context.Background(), "policies", "uuid", map[string]interface{}{"name": "Updated"}, nil)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestDelete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/items/articles/1", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := newTestClient(server).Delete(context.Background(), "articles", "1")
	require.NoError(t, err)
}

func TestDelete_MissingCollection(t *testing.T) {
	err := offlineClient().Delete(context.Background(), "", "1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collection is required")
}

func TestDelete_MissingID(t *testing.T) {
	err := offlineClient().Delete(context.Background(), "articles", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

func TestDelete_SystemCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/policies/uuid-1", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := newTestClient(server).Delete(context.Background(), "policies", "uuid-1")
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Ping
// ---------------------------------------------------------------------------

func TestPing_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/server/ping", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		w.Write([]byte("pong"))
	}))
	defer server.Close()

	err := newTestClient(server).Ping(context.Background())
	require.NoError(t, err)
}

func TestPing_UnexpectedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("unexpected"))
	}))
	defer server.Close()

	err := newTestClient(server).Ping(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected ping response")
}

// ---------------------------------------------------------------------------
// handleErrorResponse
// ---------------------------------------------------------------------------

func TestHandleErrorResponse(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedError string
	}{
		{
			name:          "directus v11 error with code",
			statusCode:    400,
			responseBody:  `{"errors":[{"message":"Invalid request","extensions":{"code":"INVALID_PAYLOAD"}}]}`,
			expectedError: "HTTP 400 [INVALID_PAYLOAD]: Invalid request",
		},
		{
			name:          "directus v11 error without code",
			statusCode:    403,
			responseBody:  `{"errors":[{"message":"You don't have permission to access this."}]}`,
			expectedError: "HTTP 403: You don't have permission to access this.",
		},
		{
			name:          "unstructured error response",
			statusCode:    500,
			responseBody:  `Internal Server Error`,
			expectedError: "HTTP 500: Internal Server Error",
		},
		{
			name:          "empty errors array",
			statusCode:    422,
			responseBody:  `{"errors":[]}`,
			expectedError: "HTTP 422:",
		},
		{
			name:          "error with empty message",
			statusCode:    400,
			responseBody:  `{"errors":[{"message":"","extensions":{"code":"UNKNOWN"}}]}`,
			expectedError: "HTTP 400:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			resp, httpErr := http.Get(server.URL)
			require.NoError(t, httpErr)
			defer resp.Body.Close()

			err := (&Client{}).handleErrorResponse(resp)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// ---------------------------------------------------------------------------
// HTTP error propagation for CRUD methods
// ---------------------------------------------------------------------------

func TestGet_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"errors":[{"message":"Not found","extensions":{"code":"NOT_FOUND"}}]}`))
	}))
	defer server.Close()

	var result map[string]interface{}
	err := newTestClient(server).Get(context.Background(), "articles", "999", &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "NOT_FOUND")
}

func TestCreate_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`{"errors":[{"message":"Validation failed","extensions":{"code":"INVALID_PAYLOAD"}}]}`))
	}))
	defer server.Close()

	err := newTestClient(server).Create(context.Background(), "articles", map[string]interface{}{"x": 1}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "400")
}

func TestUpdate_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte(`{"errors":[{"message":"Forbidden"}]}`))
	}))
	defer server.Close()

	err := newTestClient(server).Update(context.Background(), "articles", "1", map[string]interface{}{"x": 1}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestDelete_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"errors":[{"message":"Not found"}]}`))
	}))
	defer server.Close()

	err := newTestClient(server).Delete(context.Background(), "articles", "999")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

// ---------------------------------------------------------------------------
// Context handling
// ---------------------------------------------------------------------------

func TestContext_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		BaseURL:    server.URL,
		Token:      "test-token",
		HTTPClient: &http.Client{Timeout: 100 * time.Millisecond},
	}

	err := c.Get(context.Background(), "articles", "1", &map[string]interface{}{})
	require.Error(t, err)
}

func TestContext_Cancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := newTestClient(server).Get(ctx, "articles", "1", &map[string]interface{}{})
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Authorization header
// ---------------------------------------------------------------------------

func TestAuthorizationHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer my-secret-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": []interface{}{}})
	}))
	defer server.Close()

	c := &Client{
		BaseURL:    server.URL,
		Token:      "my-secret-token",
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}

	err := c.List(context.Background(), "test_collection", &map[string]interface{}{})
	require.NoError(t, err)
}
