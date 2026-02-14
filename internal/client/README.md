# Directus API Client

A robust Go client for interacting with the Directus API using static token authentication.

## Features

- **Static Token Authentication**: Simple and secure authentication using Bearer tokens
- **CRUD Operations**: Implements all base operations (Get, List, Create, Update, Delete)
- **Error Handling**: Clear error messages with proper HTTP status code handling
- **Context Support**: Full context support for timeouts and cancellation
- **Type Safety**: Generic methods that work with any data structure
- **Server Health Check**: Includes a Ping method to verify server connectivity

## Installation

```go
import "github.com/kylindc/terraform-provider-directus/internal/client"
```

## Usage

### Creating a Client

```go
config := client.Config{
    BaseURL: "https://your-directus-instance.com",
    Token:   "your-static-token",
    Timeout: 30 * time.Second, // optional, defaults to 30s
}

apiClient, err := client.NewClient(context.Background(), config)
if err != nil {
    log.Fatal(err)
}
```

The client automatically adds the `Authorization: Bearer {token}` header to all requests.

### Get a Single Item

```go
var result map[string]interface{}
err := apiClient.Get(context.Background(), "articles", "123", &result)
if err != nil {
    log.Fatal(err)
}
```

### List Items

```go
var result map[string]interface{}
err := apiClient.List(context.Background(), "articles", &result)
if err != nil {
    log.Fatal(err)
}
```

### Create an Item

```go
data := map[string]interface{}{
    "title":   "New Article",
    "content": "Article content",
    "status":  "published",
}

var result map[string]interface{}
err := apiClient.Create(context.Background(), "articles", data, &result)
if err != nil {
    log.Fatal(err)
}
```

### Update an Item

```go
data := map[string]interface{}{
    "title": "Updated Article Title",
}

var result map[string]interface{}
err := apiClient.Update(context.Background(), "articles", "123", data, &result)
if err != nil {
    log.Fatal(err)
}
```

### Delete an Item

```go
err := apiClient.Delete(context.Background(), "articles", "123")
if err != nil {
    log.Fatal(err)
}
```

### Health Check

```go
err := apiClient.Ping(context.Background())
if err != nil {
    log.Fatal("Directus server is not reachable:", err)
}
```

## Error Handling

The client provides clear error messages for all failure scenarios:

- **Authentication Errors**: Invalid or missing token
- **Validation Errors**: Missing required parameters (collection, ID, data)
- **HTTP Errors**: Structured error responses from the API with status codes
- **Network Errors**: Connection failures, timeouts
- **Context Errors**: Cancellation and timeout handling

Example error handling:

```go
err := apiClient.Get(ctx, "articles", "123", &result)
if err != nil {
    // Error messages include HTTP status codes and API error messages
    // Example: "HTTP 404: Item not found"
    log.Printf("Failed to get article: %v", err)
    return err
}
```

## Context Support

All methods accept a `context.Context` parameter for proper timeout and cancellation handling:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

err := apiClient.Get(ctx, "articles", "123", &result)

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Cancel the request programmatically
cancel()
```

## Testing

The package includes comprehensive unit tests covering:

- Static token authentication
- All CRUD operations
- Error handling scenarios
- Context timeout and cancellation
- Input validation
- Authorization header formatting

Run tests with:

```bash
go test -v ./internal/client/...
```

## Thread Safety

The client is safe for concurrent use. Multiple goroutines can share the same client instance.

## Best Practices

1. **Reuse Client Instances**: Create one client and reuse it across your application
2. **Always Use Context**: Pass appropriate contexts for timeout and cancellation control
3. **Handle Errors**: Always check and handle errors returned by client methods
4. **Set Appropriate Timeouts**: Configure timeouts based on your use case
5. **Validate Input**: The client validates required parameters, but validate your data before passing it

## API Reference

### Config

```go
type Config struct {
    BaseURL string        // Required: Base URL of the Directus instance
    Token   string        // Required: Static authentication token
    Timeout time.Duration // Optional: HTTP client timeout (default: 30s)
}
```

### Client Methods

- `NewClient(ctx context.Context, config Config) (*Client, error)`: Create a new client
- `Get(ctx context.Context, collection, id string, result interface{}) error`: Get a single item
- `List(ctx context.Context, collection string, result interface{}) error`: List items
- `Create(ctx context.Context, collection string, data interface{}, result interface{}) error`: Create an item
- `Update(ctx context.Context, collection, id string, data interface{}, result interface{}) error`: Update an item
- `Delete(ctx context.Context, collection, id string) error`: Delete an item
- `Ping(ctx context.Context) error`: Check server connectivity
