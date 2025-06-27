## CPU Optimization

### Branch Prediction Optimization

- **Likelihood-ordered conditionals**: Place most likely conditions first
- **Avoid unpredictable branches**: Use branchless programming when possible
- **Profile-guided optimization**: Use `go build -pgo` when available
- **Early cheap checks**: Put fast checks before expensive ones
- **Unroll small loops**: When iterations are small and predictable

```go
// Error types for validation
var (
    ErrEmpty         = errors.New("input is empty")
    ErrTooLong       = errors.New("input too long")
    ErrInvalidFormat = errors.New("invalid format")
)

const maxLength = 1000

// Compiled regex pattern (compile once, use many times)
var complexPattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// Good: Most common case first
func validate(input string) error {
    if input == "" { // Most common case first (80% of calls)
        return ErrEmpty
    }
    if len(input) > maxLength { // Second most common (15% of calls)
        return ErrTooLong
    }
    if !isValidFormat(input) { // Least likely (5% of calls)
        return ErrInvalidFormat
    }
    return nil
}

func isValidFormat(input string) bool {
    return complexPattern.MatchString(input)
}

// Branchless optimization for simple cases
func max(a, b int) int {
    return a + ((b-a)&((b-a)>>63)) // Branchless max for 64-bit
}

// Array-based conditional to avoid branches
func countEvenOdd(data []int) (evens, odds int) {
    var counts [2]int
    for _, value := range data {
        // Instead of: if value%2 == 0 { evens++ } else { odds++ }
        counts[value&1]++ // Branchless counting
    }
    return counts[0], counts[1] // evens, odds
}

// Early cheap checks pattern
func expensiveValidation(input string) error {
    // Fast checks first
    if len(input) == 0 {
        return ErrEmpty
    }
    if len(input) > maxLength {
        return ErrTooLong
    }

    // Expensive regex check last
    if !complexPattern.MatchString(input) {
        return ErrInvalidFormat
    }
    return nil
}
```

### Cache Locality Optimization

- **Sequential memory access**: Prefer slices over maps for iteration
- **Data structure alignment**: Group related fields together
- **Cache-line aware design**: Pad atomic fields to prevent false sharing
- **Sort data for locality**: Sorted data improves both cache and branch prediction
- **Batch processing**: Process data in cache-line sized chunks

```go
// Item represents a data item to process
type Item struct {
    ID       uint64
    Name     string
    Priority int
}

func (i *Item) Process() {
    // Process the item
}

// User represents a user with location information
type User struct {
    ID         uint64
    Name       string
    LocationID int
}

// Cache-line padding for atomic fields (64 bytes = cache line size)
type Counter struct {
    value atomic.Uint64
    _     [7]uint64 // Padding to prevent false sharing

    // Related non-atomic fields grouped together
    name     string
    category string
    enabled  bool
    _        [3]byte // Explicit padding for alignment
}

// Sequential access optimization
func processItems(items []Item) {
    // Good: Sequential memory access
    for i := range items {
        items[i].Process()
    }

    // Avoid: Random memory access through pointers
    // for _, item := range itemPointers {
    //     item.Process()
    // }
}

// Sorting for better cache locality
func processUsersByLocation(users []User) {
    // Sort by location to improve cache locality
    sort.Slice(users, func(i, j int) bool {
        return users[i].LocationID < users[j].LocationID
    })

    // Now process sequentially - much better cache usage
    for _, user := range users {
        processUserByLocation(user)
    }
}

func processUserByLocation(user User) {
    // Process user based on location
}
```

### SIMD and Vectorization

- **Use slices for vectorizable operations**: Enable Go's auto-vectorization
- **Batch operations**: Process data in chunks that fit CPU vector units
- **Aligned memory access**: Use properly aligned data structures
- **Simple loops**: Go compiler can auto-vectorize simple arithmetic loops

```go
// Vectorizable operations (Go compiler can auto-vectorize)
func addVectors(a, b, result []float64) {
    // Ensure slices have same length for vectorization
    if len(a) != len(b) || len(a) != len(result) {
        panic("slice length mismatch")
    }

    // Simple loop - Go compiler will vectorize this
    for i := range a {
        result[i] = a[i] + b[i]
    }
}

// Batch processing for better cache utilization
const batchSize = 1024 // Tune based on L1 cache size

func processBatches(data []Item) {
    for i := 0; i < len(data); i += batchSize {
        end := i + batchSize
        if end > len(data) {
            end = len(data)
        }

        // Process batch - better cache locality
        processBatch(data[i:end])
    }
}

func processBatch(batch []Item) {
    for i := range batch {
        batch[i].Process()
    }
}

// Unroll loops for known small iterations
func sumArray4(data [4]int) int {
    // Manually unrolled for exactly 4 elements
    return data[0] + data[1] + data[2] + data[3]
}
```

### CPU-Specific Optimizations

- **Minimize system calls**: Batch I/O operations
- **Reduce context switches**: Use goroutine pools instead of unlimited goroutines
- **CPU affinity awareness**: Consider NUMA topology for large applications
- **Avoid expensive operations in hot paths**: Move to initialization or background
- **Use fast algorithms for known problem sizes**: Choose algorithm based on input size

```go
// Work represents a unit of work for the worker pool
type Work interface {
    Execute() error
}

// Result represents the result of work execution
type Result struct {
    ID    uint64
    Data  []byte
    Error error
}

// Goroutine pool to prevent excessive context switching
type WorkerPool struct {
    workers    int
    workChan   chan Work
    resultChan chan Result
    wg         sync.WaitGroup
}

func NewWorkerPool(numWorkers int) *WorkerPool {
    // Typically: runtime.NumCPU() or runtime.NumCPU() * 2
    if numWorkers <= 0 {
        numWorkers = runtime.NumCPU()
    }

    return &WorkerPool{
        workers:    numWorkers,
        workChan:   make(chan Work, numWorkers*2), // Buffered to prevent blocking
        resultChan: make(chan Result, numWorkers*2),
    }
}

// Batch system calls
func writeFiles(files map[string][]byte) error {
    // Bad: Multiple syscalls
    // for name, data := range files {
    //     os.WriteFile(name, data, 0644)
    // }

    // Good: Batch with async I/O
    var wg sync.WaitGroup
    errChan := make(chan error, len(files))

    for name, data := range files {
        wg.Add(1)
        go func(name string, data []byte) {
            defer wg.Done()
            if err := os.WriteFile(name, data, 0644); err != nil {
                errChan <- err
            }
        }(name, data)
    }

    wg.Wait()
    close(errChan)

    for err := range errChan {
        if err != nil {
            return err
        }
    }

    return nil
}
```
