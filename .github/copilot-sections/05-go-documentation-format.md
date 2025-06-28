# Go Documentation Format

## Mandatory Godoc Format

### Function Documentation

```go
// FunctionName performs the main operation described here
// Code block:
//
//  result, err := FunctionName(ctx, "input", 42)
//  if err != nil {
//      log.Fatal(err)
//  }
//  fmt.Println(result)
//
// Parameters:
//   - 1 ctx: context.Context - context for cancellation and timeout
//   - 2 input: string - the input to process (must not be empty)
//   - 3 count: int - number of iterations (must be positive)
//
// Returns:
//   - 1 result: string - the processed output
//   - 2 error - non-nil if validation fails or processing errors
func FunctionName(ctx context.Context, input string, count int) (string, error) {
    // Implementation
}
```

### Type Documentation

```go
// ServiceManager handles all service operations with thread-safe access
type ServiceManager struct {
    mu       sync.RWMutex        // Protects concurrent access
    services map[string]*Service // Active services
    counter  atomic.Uint64       // Request counter
}

// Config defines service configuration options
type Config struct {
    Host     string        `json:"host"`     // Server hostname
    Port     int           `json:"port"`     // Server port (1-65535)
    Timeout  time.Duration `json:"timeout"`  // Request timeout
}
```

### Interface Documentation

```go
// Storage defines methods for data persistence
type Storage interface {
    // Save stores data with the given key
    Save(ctx context.Context, key string, data []byte) error

    // Load retrieves data for the given key
    Load(ctx context.Context, key string) ([]byte, error)

    // Delete removes data for the given key
    Delete(ctx context.Context, key string) error
}
```

### Method Documentation

```go
// Start initializes and starts the service
// Code block:
//
//  service := NewService(config)
//  if err := service.Start(ctx); err != nil {
//      log.Fatal(err)
//  }
//  defer service.Stop()
//
// Parameters:
//   - 1 ctx: context.Context - startup context
//
// Returns:
//   - 1 error - nil if successful, error if startup fails
func (s *Service) Start(ctx context.Context) error {
    // Implementation
}
```

### Constants and Variables

```go
// MaxRetries defines maximum retry attempts for failed operations
const MaxRetries = 3

// DefaultTimeout is the default operation timeout
const DefaultTimeout = 30 * time.Second

// ErrNotFound indicates the requested item was not found
var ErrNotFound = errors.New("item not found")
```

### Package Documentation

```go
// Package cache provides high-performance caching with zero-allocation patterns.
//
// The cache package implements thread-safe caching with atomic operations
// and memory pooling for optimal performance.
//
// Example usage:
//
//  cache := cache.New(cache.Config{
//      MaxSize: 1000,
//      TTL:     time.Hour,
//  })
//
//  cache.Set("key", data)
//  data, found := cache.Get("key")
package cache
```

## Documentation Rules

1. **First line**: Always start with the name being documented
2. **Code blocks**: Always include practical usage example
3. **Parameters**: Number each parameter with type and constraints
4. **Returns**: Number each return value with type and meaning
5. **English only**: All documentation must be in English
6. **No redundancy**: Don't repeat the obvious from code
