## Mandatory Documentation Format

Every exported symbol MUST use this exact format:

```go
// FunctionName Description of what the function does
// Code block:
//
//  result, err := FunctionName("input", 42)
//  if err != nil {
//      log.Fatal(err)
//  }
//  fmt.Println(result)
//
// Parameters:
//   - 1 input: string - the input string to process (must not be empty)
//   - 2 count: int - the number of iterations (must be positive)
//   - 3 opts: *Options - optional configuration (can be nil)
//
// Returns:
//   - 1 result: string - the processed output
//   - 2 error - non-nil if validation fails or processing errors occur
func FunctionName(input string, count int, opts *Options) (string, error) {
    // Implementation
}

// TypeName Description of the type and its purpose
type TypeName struct {
    Field1 string       // Description of field 1
    Field2 atomic.Int64 // Description of field 2
    mu     sync.RWMutex // Description of field 3
}

// constantName Description of what this constant represents
const constantName = 30

// variableName Description of what this variable holds
var variableName = "default value"

// InterfaceName Description of what this interface defines
type InterfaceName interface {
    Method1(param string) error        // Description of method 1
    Method2() (string, error)          // Description of method 2
    Close() error                      // Description of cleanup method
}

// ErrorName Description of this error type
type ErrorName struct {
    Code    int    // Error code
    Message string // Error message
}

// Error implements the error interface
func (e ErrorName) Error() string {
    return e.Message
}

// FunctionNoParams Description of function with no parameters
// Code block:
//
//  result := FunctionNoParams()
//  fmt.Println(result)
//
// Returns:
//   - 1 result: string - the result value
func FunctionNoParams() string {
    // Implementation
}

// FunctionNoReturns Description of function with no returns
// Code block:
//
//  FunctionNoReturns("config")
//  fmt.Println("Done")
//
// Parameters:
//   - 1 config: string - configuration value (must not be empty)
func FunctionNoReturns(config string) {
    // Implementation
}

// Start Description of method with receiver
// Code block:
//
//  service := &ServiceManager{}
//  err := service.Start(ctx)
//  if err != nil {
//      log.Fatal(err)
//  }
//
// Parameters:
//   - 1 ctx: context.Context - context for cancellation and timeout
//
// Returns:
//   - 1 error - nil if successful, error if startup fails
func (s *ServiceManager) Start(ctx context.Context) error {
    // Implementation
}

// UserID Description of type alias
type UserID string

// StatusCode Description of grouped constants
const (
    StatusOK    StatusCode = 200 // Request successful
    StatusError StatusCode = 500 // Internal server error
    StatusRetry StatusCode = 503 // Service temporarily unavailable
)

// ProcessFiles Description of variadic function
// Code block:
//
//  err := ProcessFiles("file1.txt", "file2.txt", "file3.txt")
//  if err != nil {
//      log.Fatal(err)
//  }
//
// Parameters:
//   - 1 files: ...string - list of file paths to process (must not be empty)
//
// Returns:
//   - 1 error - nil if all files processed successfully
func ProcessFiles(files ...string) error {
    // Implementation
}

// NewService Description of constructor function
// Code block:
//
//  service := NewService(config, logger)
//  defer service.Close()
//
// Parameters:
//   - 1 config: *Config - service configuration (cannot be nil)
//   - 2 logger: Logger - logging interface (cannot be nil)
//
// Returns:
//   - 1 service: *ServiceManager - configured service instance
func NewService(config *Config, logger Logger) *ServiceManager {
    // Implementation
}
```
