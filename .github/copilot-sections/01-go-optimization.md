# Go Optimization Rules (.go files ONLY)

## Zero-Allocation Patterns (Mandatory)

### Pre-allocation with Exact Capacity

```go
// ✅ ALWAYS: Pre-allocate slices
items := make([]Item, 0, len(source))

// ✅ ALWAYS: Pre-allocate maps
cache := make(map[string]*Item, estimatedCount)

// ✅ ALWAYS: Pre-grow strings.Builder
var builder strings.Builder
builder.Grow(estimatedSize)
```

### Buffer Pool Management

```go
// ✅ ALWAYS: Reuse buffers with sync.Pool
var bufferPool = sync.Pool{
    New: func() any {
        return make([]byte, 0, 4096) // 4KB initial
    },
}

func ProcessData() []byte {
    buf := bufferPool.Get().([]byte)
    defer func() {
        buf = buf[:0] // Reset length, keep capacity
        bufferPool.Put(buf)
    }()

    // Use buffer...
    return append([]byte(nil), buf...) // Return copy
}
```

## Struct Optimization

### Field Ordering (Size Descending)

```go
// ✅ ALWAYS: Order fields by size
type OptimizedStruct struct {
    // 8-byte fields first
    id        int64
    timestamp int64

    // 4-byte fields
    count     int32
    flags     uint32

    // 1-byte fields + padding
    active    bool
    status    byte
    _         [2]byte // Explicit padding

    // Pointer fields last
    name      string
    data      []byte
}
```

### Cache-Line Padding

```go
// ✅ ALWAYS: Pad atomic fields
type Counter struct {
    value atomic.Uint64
    _     [7]uint64 // 64-byte cache line
}
```

## Allocation Avoidance

### Value vs Pointer Decision

```go
// ✅ Small structs (<= 64 bytes): use values
type Point struct {
    X, Y, Z float64 // 24 bytes - pass by value
}

// ✅ Large structs or optional fields: use pointers
type User struct {
    ID      int64
    Name    string
    Avatar  *ImageData // Large/optional - use pointer
}
```

### String Building

```go
// ❌ BAD: Concatenation in loops
var s string
for _, part := range parts {
    s += part // Allocation each iteration
}

// ✅ GOOD: strings.Builder
var builder strings.Builder
builder.Grow(len(parts) * avgPartSize)
for _, part := range parts {
    builder.WriteString(part)
}
s := builder.String()
```

## Zero I/O Strategy (Primary Goal)

### Memory-First Approach

```go
// ✅ ALWAYS: Keep everything in memory when possible
type Service struct {
    cache sync.Map // No disk I/O
}

// ✅ ALWAYS: Pre-load at startup
func NewService(dataDir string) (*Service, error) {
    svc := &Service{}

    // Load all data into memory once
    files, err := os.ReadDir(dataDir)
    if err != nil {
        return nil, err
    }

    for _, file := range files {
        if file.IsDir() {
            continue
        }

        data, err := os.ReadFile(filepath.Join(dataDir, file.Name()))
        if err != nil {
            continue
        }

        svc.cache.Store(file.Name(), data)
    }

    return svc, nil
}
```

## Buffered I/O (When I/O Required)

### Large Buffer Sizes

```go
// ✅ ALWAYS: Use large buffers to reduce syscalls
const bufferSize = 64 * 1024 // 64KB

writer := bufio.NewWriterSize(file, bufferSize)
reader := bufio.NewReaderSize(file, bufferSize)
```

### Batch Operations

```go
// ✅ ALWAYS: Batch multiple operations
type BatchWriter struct {
    file     *os.File
    buffer   *bufio.Writer
    pending  [][]byte
    maxBatch int
}

func (bw *BatchWriter) Write(data []byte) error {
    bw.pending = append(bw.pending, data)

    if len(bw.pending) >= bw.maxBatch {
        return bw.Flush()
    }

    return nil
}

func (bw *BatchWriter) Flush() error {
    for _, data := range bw.pending {
        if _, err := bw.buffer.Write(data); err != nil {
            return err
        }
    }

    bw.pending = bw.pending[:0]
    return bw.buffer.Flush()
}
```

## File System Optimization

### Minimize Stat Calls

```go
// ❌ BAD: Multiple stat calls
if _, err := os.Stat(filename); err == nil {
    data, err := os.ReadFile(filename)
}

// ✅ GOOD: Direct operation with error handling
data, err := os.ReadFile(filename)
if err != nil {
    if os.IsNotExist(err) {
        // Handle missing file
    }
    return err
}
```

### Efficient Directory Reading

```go
// ✅ ALWAYS: Use ReadDir for multiple files
entries, err := os.ReadDir(dirPath)
if err != nil {
    return err
}

// Process all entries with single syscall
for _, entry := range entries {
    if !entry.IsDir() {
        info, _ := entry.Info()
        processFile(info)
    }
}
```

## Memory-Mapped Files

### For Large Read-Only Files

```go
// ✅ For files > 10MB, consider mmap
func ReadLargeFile(filename string) ([]byte, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    stat, err := file.Stat()
    if err != nil {
        return nil, err
    }

    if stat.Size() > 10*1024*1024 { // > 10MB
        // Use mmap for zero-copy access
        return mmap.Map(file, mmap.RDONLY, 0)
    }

    // Small file - regular read
    return io.ReadAll(file)
}
```

## Async I/O Pattern

```go
// ✅ ALWAYS: Async writes with buffering
type AsyncWriter struct {
    writeChan chan writeRequest
    errorChan chan error
}

type writeRequest struct {
    filename string
    data     []byte
    response chan error
}

func (aw *AsyncWriter) Start() {
    go func() {
        for req := range aw.writeChan {
            err := os.WriteFile(req.filename, req.data, 0644)
            req.response <- err
        }
    }()
}
```

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

## Structured Concurrency (Mandatory)

### No Fire-and-Forget

```go
// ❌ NEVER: Unsupervised goroutines
func BadPattern() {
    go func() {
        // No error handling
        // No cancellation
        // No monitoring
        doWork()
    }()
}

// ✅ ALWAYS: Supervised concurrency
func GoodPattern(ctx context.Context) error {
    errCh := make(chan error, 1)

    go func() {
        defer func() {
            if r := recover(); r != nil {
                errCh <- fmt.Errorf("panic: %v", r)
            }
        }()

        errCh <- doWork(ctx)
    }()

    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### Worker Pool Pattern

```go
// ✅ ALWAYS: Bounded concurrency
type WorkerPool struct {
    workers  int
    queue    chan Job
    results  chan Result
    wg       sync.WaitGroup

    // Metrics
    processed atomic.Uint64
    errors    atomic.Uint64
}

func NewWorkerPool(workers int, queueSize int) *WorkerPool {
    return &WorkerPool{
        workers: workers,
        queue:   make(chan Job, queueSize),
        results: make(chan Result, queueSize),
    }
}

func (p *WorkerPool) Start(ctx context.Context) {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker(ctx, i)
    }
}

func (p *WorkerPool) worker(ctx context.Context, id int) {
    defer p.wg.Done()

    for {
        select {
        case job := <-p.queue:
            result := p.processJob(ctx, job)

            select {
            case p.results <- result:
                if result.Err != nil {
                    p.errors.Add(1)
                } else {
                    p.processed.Add(1)
                }
            case <-ctx.Done():
                return
            }

        case <-ctx.Done():
            return
        }
    }
}

func (p *WorkerPool) Stop() {
    close(p.queue)
    p.wg.Wait()
    close(p.results)
}
```

## Channel Patterns

### Buffered Channels

```go
// ✅ ALWAYS: Use buffered channels for async operations
ch := make(chan Message, 100) // Prevent blocking

// ✅ ALWAYS: Handle channel closing
func (w *Worker) Start(ctx context.Context) {
    for {
        select {
        case msg, ok := <-w.input:
            if !ok {
                return // Channel closed
            }
            w.process(msg)

        case <-ctx.Done():
            return
        }
    }
}
```

### Fan-Out/Fan-In

```go
// ✅ Distribute work across multiple workers
func FanOut(ctx context.Context, in <-chan Job, workers int) []<-chan Result {
    outs := make([]<-chan Result, workers)

    for i := 0; i < workers; i++ {
        out := make(chan Result)
        outs[i] = out

        go func() {
            defer close(out)
            for job := range in {
                select {
                case out <- process(job):
                case <-ctx.Done():
                    return
                }
            }
        }()
    }

    return outs
}

// ✅ Merge results from multiple channels
func FanIn(ctx context.Context, inputs ...<-chan Result) <-chan Result {
    out := make(chan Result)
    var wg sync.WaitGroup

    for _, in := range inputs {
        wg.Add(1)
        go func(ch <-chan Result) {
            defer wg.Done()
            for {
                select {
                case res, ok := <-ch:
                    if !ok {
                        return
                    }
                    select {
                    case out <- res:
                    case <-ctx.Done():
                        return
                    }
                case <-ctx.Done():
                    return
                }
            }
        }(in)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

## Error Group Pattern

### Concurrent Operations with Error Handling

```go
// ✅ ALWAYS: Use errgroup for multiple concurrent operations
func ProcessItems(ctx context.Context, items []Item) error {
    g, ctx := errgroup.WithContext(ctx)

    // Limit concurrency
    sem := make(chan struct{}, runtime.NumCPU())

    for _, item := range items {
        item := item // Capture loop variable

        g.Go(func() error {
            // Acquire semaphore
            select {
            case sem <- struct{}{}:
                defer func() { <-sem }()
            case <-ctx.Done():
                return ctx.Err()
            }

            return processItem(ctx, item)
        })
    }

    return g.Wait()
}
```

## Timeout Patterns

### Operation Timeouts

```go
// ✅ ALWAYS: Set timeouts for operations
func CallService(ctx context.Context, data []byte) ([]byte, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    resultCh := make(chan []byte, 1)
    errCh := make(chan error, 1)

    go func() {
        result, err := slowOperation(data)
        if err != nil {
            errCh <- err
            return
        }
        resultCh <- result
    }()

    select {
    case result := <-resultCh:
        return result, nil
    case err := <-errCh:
        return nil, err
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

## Rate Limiting

### Token Bucket Pattern

```go
// ✅ Rate limit concurrent operations
type RateLimiter struct {
    tokens chan struct{}
    ticker *time.Ticker
}

func NewRateLimiter(rps int) *RateLimiter {
    rl := &RateLimiter{
        tokens: make(chan struct{}, rps),
        ticker: time.NewTicker(time.Second / time.Duration(rps)),
    }

    // Fill bucket
    for i := 0; i < rps; i++ {
        rl.tokens <- struct{}{}
    }

    // Refill tokens
    go func() {
        for range rl.ticker.C {
            select {
            case rl.tokens <- struct{}{}:
            default: // Bucket full
            }
        }
    }()

    return rl
}

func (rl *RateLimiter) Allow() bool {
    select {
    case <-rl.tokens:
        return true
    default:
        return false
    }
}
```

## Test Structure (Mandatory)

### Table-Driven Tests

```go
// ✅ ALWAYS: Use table-driven tests
func TestCalculate(t *testing.T) {
    tests := []struct {
        name    string
        input   int
        want    int
        wantErr bool
    }{
        {
            name:    "positive_number",
            input:   5,
            want:    25,
            wantErr: false,
        },
        {
            name:    "zero",
            input:   0,
            want:    0,
            wantErr: false,
        },
        {
            name:    "negative_number",
            input:   -5,
            want:    0,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Calculate(tt.input)

            if tt.wantErr {
                require.Error(t, err)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Timeout Patterns

### Test Timeouts

```go
// ✅ ALWAYS: Set test timeouts
func TestWithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Run test with timeout
    done := make(chan bool)
    go func() {
        // Test logic here
        result := performOperation(ctx)
        assert.NotNil(t, result)
        done <- true
    }()

    select {
    case <-done:
        // Test completed
    case <-ctx.Done():
        t.Fatal("test timeout exceeded")
    }
}

// ✅ Per-test timeouts in table tests
func TestOperations(t *testing.T) {
    tests := []struct {
        name    string
        timeout time.Duration
        fn      func(context.Context) error
    }{
        {
            name:    "fast_operation",
            timeout: 100 * time.Millisecond,
            fn:      fastOperation,
        },
        {
            name:    "slow_operation",
            timeout: 5 * time.Second,
            fn:      slowOperation,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
            defer cancel()

            err := tt.fn(ctx)
            require.NoError(t, err)
        })
    }
}
```

## Parallel Testing

### Concurrent Test Safety

```go
// ✅ ALWAYS: Use t.Parallel() for independent tests
func TestParallel(t *testing.T) {
    t.Parallel() // Mark test as parallel-safe

    // Test implementation
}

// ✅ Test concurrent access
func TestConcurrentAccess(t *testing.T) {
    service := NewService()

    const numGoroutines = 100
    var wg sync.WaitGroup
    errors := make(chan error, numGoroutines)

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            if err := service.Process(ctx, id); err != nil {
                errors <- err
            }
        }(i)
    }

    // Wait with timeout
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        close(errors)
        for err := range errors {
            t.Errorf("concurrent error: %v", err)
        }
    case <-ctx.Done():
        t.Fatal("test timeout")
    }
}
```

## Mock Patterns

### Interface Mocking

```go
//go:generate mockgen -source=storage.go -destination=mocks/mock_storage.go

// Storage interface for mocking
type Storage interface {
    Save(ctx context.Context, key string, data []byte) error
    Load(ctx context.Context, key string) ([]byte, error)
}

// ✅ Test with mocks
func TestServiceWithMock(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockStorage := mocks.NewMockStorage(ctrl)
    service := NewService(mockStorage)

    ctx := context.Background()
    testData := []byte("test")

    // Set expectations
    mockStorage.EXPECT().
        Save(ctx, "key", testData).
        Return(nil).
        Times(1)

    // Execute test
    err := service.Store(ctx, "key", testData)
    require.NoError(t, err)
}
```

## Benchmark Patterns

### Performance Testing

```go
// ✅ Benchmark with memory stats
func BenchmarkOperation(b *testing.B) {
    // Setup
    data := generateTestData(1000)

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        _ = processData(data)
    }
}

// ✅ Comparative benchmarks
func BenchmarkAlgorithms(b *testing.B) {
    sizes := []int{10, 100, 1000, 10000}

    for _, size := range sizes {
        b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
            data := generateTestData(size)

            b.Run("algorithm_v1", func(b *testing.B) {
                b.ResetTimer()
                for i := 0; i < b.N; i++ {
                    _ = algorithmV1(data)
                }
            })

            b.Run("algorithm_v2", func(b *testing.B) {
                b.ResetTimer()
                for i := 0; i < b.N; i++ {
                    _ = algorithmV2(data)
                }
            })
        })
    }
}
```

## Test Helpers

### Reusable Test Functions

```go
// ✅ Test helper functions
func setupTest(t *testing.T) (*Service, func()) {
    t.Helper()

    // Setup
    tmpDir := t.TempDir()
    service := NewService(tmpDir)

    // Cleanup function
    cleanup := func() {
        service.Close()
    }

    return service, cleanup
}

// Usage
func TestFeature(t *testing.T) {
    service, cleanup := setupTest(t)
    defer cleanup()

    // Test implementation
}
```

## Coverage Requirements

### Achieving 100% Coverage

```go
// ✅ Test all paths including errors
func TestCompleteCodePath(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        mockSetup func(*mocks.MockStorage)
        wantErr   bool
        errMsg    string
    }{
        {
            name:  "success_path",
            input: "valid",
            mockSetup: func(m *mocks.MockStorage) {
                m.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
            },
            wantErr: false,
        },
        {
            name:  "validation_error",
            input: "",
            mockSetup: func(m *mocks.MockStorage) {
                // No mock calls expected
            },
            wantErr: true,
            errMsg:  "input is empty",
        },
        {
            name:  "storage_error",
            input: "valid",
            mockSetup: func(m *mocks.MockStorage) {
                m.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any()).
                    Return(errors.New("storage failed"))
            },
            wantErr: true,
            errMsg:  "storage failed",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mock := mocks.NewMockStorage(ctrl)
            tt.mockSetup(mock)

            service := NewService(mock)
            err := service.Process(tt.input)

            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

## Error Handling Pattern

```go
// ALWAYS wrap errors with context
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// NEVER ignore errors
// Bad: _ = someFunction()
// Good: if err := someFunction(); err != nil { return err }
```

## Resource Management

```go
// ALWAYS use defer for cleanup
file, err := os.Open(filename)
if err != nil {
    return err
}
defer file.Close()

// ALWAYS use context for cancellation
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

## Memory Profiling Commands

```bash
# Run with memory profiling
go test -bench=. -memprofile=mem.prof

# Analyze allocations
go tool pprof -alloc_space mem.prof

# Check escape analysis
go build -gcflags="-m -m" ./...
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

## Documentation Rules

1. **First line**: Always start with the name being documented
2. **Code blocks**: Always include practical usage example
3. **Parameters**: Number each parameter with type and constraints
4. **Returns**: Number each return value with type and meaning
5. **English only**: All documentation must be in English
6. **No redundancy**: Don't repeat the obvious from code
