## Code Quality and Security

### Required Tools (Must Pass)

- `gofmt -s` (format and simplify)
- `goimports` (organize imports)
- `golangci-lint run --config .golangci.yml` (comprehensive linting)
- `go test -race -timeout=30s ./...` (race detection with timeout)
- `go test -coverprofile=coverage.out ./...` (100% coverage requirement)
- `go vet ./...` (static analysis)
- `staticcheck ./...` (additional static analysis)
- `gosec ./...` (security scanning)

### Coverage Requirements

```bash
# Minimum 100% coverage for all packages
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep "total:" | awk '{if ($3+0 < 100) exit 1}'
```

### Input Validation (Zero-Trust)

```go
// ALWAYS validate ALL inputs
func ProcessUser(name string, age int, email string) error {
    // String validation
    name = strings.TrimSpace(name)
    if name == "" || len(name) > 100 {
        return fmt.Errorf("invalid name: length must be 1-100 chars, got %d", len(name))
    }
    if !isAlphaNumeric(name) {
        return errors.New("name contains invalid characters")
    }

    // Numeric validation with overflow protection
    if age < 0 || age > 150 {
        return fmt.Errorf("invalid age: must be 0-150, got %d", age)
    }

    // Email validation (basic)
    if !emailRegex.MatchString(email) {
        return errors.New("invalid email format")
    }

    return nil
}

// Regex compilation (compile once, use many times)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isAlphaNumeric(s string) bool {
    for _, r := range s {
        if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
            return false
        }
    }
    return true
}
```

### Cryptographic Security

```go
// NEVER use math/rand for security
func GenerateToken() (string, error) {
    bytes := make([]byte, 32)
    if _, err := crypto_rand.Read(bytes); err != nil {
        return "", fmt.Errorf("failed to generate random bytes: %w", err)
    }
    return base64.URLEncoding.EncodeToString(bytes), nil
}

// Time-constant comparison for secrets
func ValidateToken(provided, expected string) bool {
    return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

// Secure memory cleanup
func ProcessSecret(secret []byte) error {
    defer func() {
        // Zero out sensitive data
        for i := range secret {
            secret[i] = 0
        }
    }()

    // Process secret...
    return nil
}
```

### Command Injection Prevention

```go
// NEVER trust user input in shell commands
func ExecuteCommand(userInput string) error {
    // Whitelist validation
    if !isValidCommand(userInput) {
        return errors.New("invalid command format")
    }

    // Use exec.Command with separate args (not shell)
    cmd := exec.Command("safe-binary", sanitizeArg(userInput))
    cmd.Env = []string{} // Empty environment

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("command failed: %w", err)
    }

    log.Printf("Command output: %s", output)
    return nil
}

func isValidCommand(input string) bool {
    // Whitelist only alphanumeric characters and safe symbols
    return len(input) > 0 && len(input) < 100 && isAlphaNumeric(input)
}

func sanitizeArg(input string) string {
    // Remove any potentially dangerous characters
    result := strings.Builder{}
    for _, r := range input {
        if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
            result.WriteRune(r)
        }
    }
    return result.String()
}
```

### Error Handling Best Practices

```go
// ALWAYS wrap errors with context
func ProcessFile(filename string) error {
    data, err := os.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("failed to read file %q: %w", filename, err)
    }

    if err := validateFileData(data); err != nil {
        return fmt.Errorf("invalid data in file %q: %w", filename, err)
    }

    if err := processData(data); err != nil {
        return fmt.Errorf("failed to process data from file %q: %w", filename, err)
    }

    return nil
}

func validateFileData(data []byte) error {
    // Basic validation (placeholder)
    if len(data) == 0 {
        return errors.New("file is empty")
    }
    if len(data) > 10*1024*1024 { // 10MB limit
        return errors.New("file too large")
    }
    return nil
}

func processData(data []byte) error {
    // Process the data (placeholder)
    if len(data) < 10 {
        return errors.New("insufficient data")
    }
    return nil
}

// Error aggregation
func ProcessMultipleFiles(filenames []string) error {
    var errs []error

    for _, filename := range filenames {
        if err := ProcessFile(filename); err != nil {
            errs = append(errs, err)
        }
    }

    if len(errs) > 0 {
        return errors.Join(errs...)
    }

    return nil
}
```

### Optimization Workflow (go-perfbook Integration)

- **The Three Questions**: Do we have to do this? Best algorithm? Best implementation?
- **Profile-driven**: Always measure before and after optimizations
- **Amdahl's Law**: Focus on bottlenecks - 80% speedup on 5% code = 2.5% total gain
- **Constant factors matter**: Same Big-O doesn't mean same performance
- **Know your input sizes**: Choose algorithms based on realistic data sizes

```go
// Record and Result types for optimization examples
type Record struct {
    ID   uint64
    Data []byte
    Meta map[string]string
}

type Result struct {
    ID        uint64
    Processed []byte
    Status    string
}

// Cache variable for optimization example
var useCache = true

// The Three Optimization Questions pattern
func optimizeDataProcessing(data []Record) []Result {
    // Question 1: Do we have to do this at all?
    if useCache {
        if cached := checkCache(data); cached != nil {
            return cached // Skip processing entirely
        }
    }

    // Question 2: Is this the best algorithm?
    if len(data) < 100 {
        return simpleLinearProcess(data) // O(n) but low constant factor
    } else {
        return efficientDivideConquer(data) // O(n log n) but high constant factor
    }

    // Question 3: Best implementation handled in individual functions
}

func checkCache(data []Record) []Result {
    // Placeholder cache check
    return nil
}

func simpleLinearProcess(data []Record) []Result {
    results := make([]Result, len(data))
    for i, record := range data {
        results[i] = Result{
            ID:        record.ID,
            Processed: record.Data,
            Status:    "linear",
        }
    }
    return results
}

func efficientDivideConquer(data []Record) []Result {
    results := make([]Result, len(data))
    for i, record := range data {
        results[i] = Result{
            ID:        record.ID,
            Processed: record.Data,
            Status:    "divide-conquer",
        }
    }
    return results
}

// Constant factor optimization example
func fastStringEquals(a, b string) bool {
    // Quick length check (constant factor improvement)
    if len(a) != len(b) {
        return false
    }

    // Early exit for same pointer (constant factor improvement)
    if len(a) > 0 && len(b) > 0 && &a[0] == &b[0] {
        return true
    }

    return a == b
}

// Input-size aware algorithm selection
func sortOptimal(data []int) {
    switch {
    case len(data) < 12:
        insertionSort(data) // Fastest for tiny arrays
    case len(data) < 1000:
        quickSort(data) // Good general purpose
    default:
        parallelSort(data) // Worth the overhead for large data
    }
}

func insertionSort(data []int) {
    for i := 1; i < len(data); i++ {
        key := data[i]
        j := i - 1
        for j >= 0 && data[j] > key {
            data[j+1] = data[j]
            j--
        }
        data[j+1] = key
    }
}

func quickSort(data []int) {
    if len(data) < 2 {
        return
    }
    // Basic quicksort implementation
    sort.Ints(data)
}

func parallelSort(data []int) {
    // For large data, use parallel sorting
    sort.Ints(data) // Go's sort is already optimized
}
```
