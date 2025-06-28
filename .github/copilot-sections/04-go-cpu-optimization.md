# Go CPU Optimization

## Atomic Operations (Mandatory)

### Counter Pattern

```go
// ✅ ALWAYS: Use atomics for shared counters
var (
    requestCount atomic.Uint64
    errorCount   atomic.Uint64
    isReady      atomic.Bool
)

// Usage
requestCount.Add(1)
if isReady.Load() {
    processRequest()
}

// Compare-and-swap pattern
for {
    old := counter.Load()
    new := old + delta
    if counter.CompareAndSwap(old, new) {
        break
    }
}
```

## Concurrency Control

### Goroutine Limiting

```go
// ✅ ALWAYS: Limit concurrent goroutines
type WorkerPool struct {
    sem chan struct{}
}

func NewWorkerPool(size int) *WorkerPool {
    return &WorkerPool{
        sem: make(chan struct{}, size),
    }
}

func (p *WorkerPool) Execute(fn func()) {
    p.sem <- struct{}{} // Acquire
    go func() {
        defer func() { <-p.sem }() // Release
        fn()
    }()
}
```

### Structured Concurrency

```go
// ✅ ALWAYS: Use context and error handling
func ProcessItems(ctx context.Context, items []Item) error {
    g, ctx := errgroup.WithContext(ctx)

    // Limit concurrent operations
    sem := make(chan struct{}, runtime.NumCPU())

    for _, item := range items {
        item := item // Capture loop variable
        g.Go(func() error {
            sem <- struct{}{}
            defer func() { <-sem }()

            return processItem(ctx, item)
        })
    }

    return g.Wait()
}
```

## Branch Prediction Optimization

### Likelihood Ordering

```go
// ✅ ALWAYS: Most likely condition first
func validate(input string) error {
    // 90% of cases - valid input
    if input != "" && len(input) <= maxLength {
        return nil
    }

    // 8% of cases - empty
    if input == "" {
        return ErrEmpty
    }

    // 2% of cases - too long
    return ErrTooLong
}
```

### Branchless Patterns

```go
// ✅ For hot paths: branchless operations
func min(a, b int) int {
    // Branchless min using bit manipulation
    return a + ((b-a)&((b-a)>>31))
}

// Conditional increment without branch
func countMatches(data []int, target int) int {
    count := 0
    for _, v := range data {
        // Branchless: add 1 if equal, 0 if not
        count += (1 - ((v ^ target) >> 31 & 1))
    }
    return count
}
```

## Cache-Friendly Access

### Sequential Memory Access

```go
// ✅ ALWAYS: Prefer sequential access
// Good: Array of structs for sequential processing
type Data []Item

// Process sequentially for cache efficiency
for i := range data {
    process(&data[i])
}

// ❌ BAD: Random pointer chasing
type Node struct {
    Value int
    Next  *Node
}
```

### Data Locality

```go
// ✅ ALWAYS: Group accessed data together
type ProcessingUnit struct {
    // Hot data - accessed frequently together
    id       uint64
    counter  uint32
    flags    uint32

    // Cold data - rarely accessed
    metadata string
    created  time.Time
}
```

## CPU Profiling

```bash
# Profile CPU usage
go test -bench=. -cpuprofile=cpu.prof

# Analyze hot spots
go tool pprof cpu.prof

# Generate flame graph
go tool pprof -http=:8080 cpu.prof
```

## SIMD-Friendly Patterns

```go
// ✅ Enable auto-vectorization with simple loops
func addVectors(a, b, result []float64) {
    // Compiler can vectorize this
    for i := range a {
        result[i] = a[i] + b[i]
    }
}

// Batch operations for efficiency
const batchSize = 1024

func processBatches(data []float64) {
    for i := 0; i < len(data); i += batchSize {
        end := min(i+batchSize, len(data))
        processBatch(data[i:end])
    }
}
```
