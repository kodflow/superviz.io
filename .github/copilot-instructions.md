````instructions

---

# Primitive Rules - Universal Core Principles

## Fundamental Optimization Order

**ALWAYS optimize in this order: Memory → Disk → CPU**

## Core Behaviors

### Analysis Before Action

- **Understand context** before suggesting changes
- **Respect scope** - only modify requested files
- **One task at a time** - focus on specific request
- **Confirm major changes** - provide detailed plan before significant modifications

### Language Requirements

- **ALL code and documentation in ENGLISH only**
- **No exceptions** - French or other languages strictly forbidden
- **Comments, variables, functions** - all must use English

### Build and Test File Management

- **ALL automation scripts created by AI** go in `.tmp/` folder at project root
- **These are helper scripts** for automation, not part of final deliverable
- **Use existing Makefile commands** - never recreate build logic
- **Clean temporary files** after creation - only keep requested files
- **Never create build artifacts** in `.tmp/` - those belong in standard locations

#### .tmp Usage Rules

```
project-root/
├── .tmp/           # AI-created automation scripts only
│   ├── test-runner.sh    # Any test automation script
│   ├── deploy-helper.py  # Any deployment script
│   ├── setup-env.sh      # Any environment setup
│   └── batch-process.js  # Any processing automation
├── Makefile        # Use existing targets
└── dist/           # Build artifacts (standard location)
```

#### What Goes in .tmp/

- **Scripts created to automate tasks** (testing, deployment, processing)
- **Helper utilities** that aren't part of the project deliverable
- **Temporary automation tools** that could be deleted without affecting the project

#### What NEVER Goes in .tmp/

- **Project source code**
- **Build artifacts** (binaries, dist files, compiled assets)
- **Configuration files** needed by the application
- **Documentation** or files requested as deliverables

#### Automation Pattern

```bash
# Create automation script in .tmp
cat > .tmp/run-tests.sh << 'EOF'
#!/bin/bash
make test  # Use existing Makefile targets
EOF
chmod +x .tmp/run-tests.sh

# Execute and clean up if temporary
./.tmp/run-tests.sh
# rm .tmp/run-tests.sh  # Only if not requested to keep
```

#### Integration with Existing Build System

- **ALWAYS check for Makefile** before creating custom logic
- **Use `make test`, `make build`, `make deploy`** instead of custom commands
- **Extend Makefile** if new targets needed, don't bypass it

### Universal Error Handling Pattern

```bash
# Bash/Shell - ALWAYS check return codes
command || { echo "Command failed"; exit 1; }
```

```python
# Python - ALWAYS capture exceptions
try:
    operation()
except Exception as e:
    logger.error(f"Operation failed: {e}")
    raise
```

```javascript
// JavaScript - ALWAYS handle errors
try {
  await operation();
} catch (error) {
  console.error("Operation failed:", error);
  throw error;
}
```

## Decision Framework

### When to Optimize

1. **Hot paths** - code executed frequently (>1000/sec)
2. **Memory pressure** - high allocation rates visible
3. **User complaints** - explicit performance issues mentioned
4. **Production scale** - mentions of "millions of users"

### When NOT to Optimize

1. **Configuration code** - executed once at startup
2. **Test code** - performance not critical
3. **Prototypes/POC** - clarity over performance
4. **Migration scripts** - one-time execution

## Response Template

1. **Acknowledge**: "I see you want to [summary]"
2. **Analyze**: "Here's what I found..."
3. **Propose**: "I suggest these changes..."
4. **Confirm**: "Shall I proceed?"

---

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

---

# Dockerfile Optimization Rules (Dockerfile\* files ONLY)

## Multi-Stage Build Optimization (Mandatory)

### Zero-Waste Image Pattern

```dockerfile
# ✅ ALWAYS: Multi-stage builds for minimal final image
FROM golang:1.21-alpine AS builder

# Build stage optimizations
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main .

# Production stage - minimal image
FROM alpine:3.18
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

### Distroless Pattern for Maximum Security

```dockerfile
# ✅ ALWAYS: Use distroless for production Go binaries
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main .

# Distroless final stage
FROM gcr.io/distroless/static-debian11
COPY --from=builder /app/main /
ENTRYPOINT ["/main"]
```

## Layer Optimization (Memory → Disk → CPU)

### Combine RUN Commands (Disk Optimization)

```dockerfile
# ✅ ALWAYS: Combine RUN commands to reduce layers
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        package1 \
        package2 \
        package3 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    rm -rf /tmp/* && \
    rm -rf /var/tmp/*
```

### Strategic Layer Ordering (Cache Optimization)

```dockerfile
# ✅ ALWAYS: Order by change frequency (least → most frequent)
FROM golang:1.21-alpine

# 1. System dependencies (change rarely)
RUN apk add --no-cache git ca-certificates tzdata

# 2. Go dependencies (change occasionally)
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# 3. Source code (changes frequently) - last
COPY . .
RUN go build -o main .
```

## .dockerignore Optimization

### Essential Exclusions

```dockerfile
# ✅ ALWAYS: Create comprehensive .dockerignore
# Version control
.git
.gitignore
.gitattributes

# Documentation
*.md
README*
CHANGELOG*
LICENSE*

# Development files
.env*
.vscode/
.idea/
*.log
tmp/
temp/

# Build artifacts
target/
dist/
build/
*.exe
*.dll
*.so

# Test files
*_test.go
testdata/
coverage.out

# Docker files
Dockerfile*
docker-compose*
.dockerignore
```

## Security Hardening

### Non-Root User Pattern

```dockerfile
# ✅ ALWAYS: Create and use non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Change ownership of working directory
WORKDIR /app
COPY --chown=appuser:appgroup . .

# Switch to non-root user
USER appuser

# Verify user
RUN id && whoami
```

### Specific Version Pinning

```dockerfile
# ✅ ALWAYS: Pin exact versions for security
FROM golang:1.21.5-alpine3.18

# Pin package versions
RUN apk add --no-cache \
    ca-certificates=20230506-r0 \
    tzdata=2023c-r1
```

## Health Check Implementation

### Comprehensive Health Monitoring

```dockerfile
# ✅ ALWAYS: Implement health checks
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# For Go applications without curl
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ./main -health-check || exit 1
```

## Build Arguments and Environment

### Flexible Build Configuration

```dockerfile
# ✅ ALWAYS: Use build args for flexibility
ARG GO_VERSION=1.21
ARG ALPINE_VERSION=3.18
ARG BUILD_DATE
ARG VERSION
ARG COMMIT_SHA

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

# Build-time labels
LABEL org.opencontainers.image.title="My Application"
LABEL org.opencontainers.image.description="High-performance Go application"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.revision="${COMMIT_SHA}"
LABEL org.opencontainers.image.source="https://github.com/user/repo"
```

## Resource Optimization

### Memory-Efficient Base Images

```dockerfile
# ✅ ALWAYS: Choose minimal base images
# Alpine for small size (5MB)
FROM alpine:3.18

# Distroless for security (static binary)
FROM gcr.io/distroless/static-debian11

# Scratch for minimal possible size (static binary only)
FROM scratch
```

### CPU-Optimized Builds

```dockerfile
# ✅ ALWAYS: Optimize Go builds for production
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s -X main.version=${VERSION} -X main.commit=${COMMIT_SHA}" \
    -o main .
```

## Volume and Data Management

### Proper Volume Handling

```dockerfile
# ✅ ALWAYS: Declare volumes for persistent data
VOLUME ["/data", "/logs"]

# Create directories with proper permissions
RUN mkdir -p /data /logs && \
    chown -R appuser:appgroup /data /logs && \
    chmod 755 /data /logs
```

## Network Configuration

### Port and Protocol Declaration

```dockerfile
# ✅ ALWAYS: Document exposed ports
EXPOSE 8080/tcp
EXPOSE 8081/tcp

# Document port purpose in comments
# 8080: HTTP API
# 8081: Metrics endpoint
```

## Build Cache Optimization

### Dependency Caching Strategy

```dockerfile
# ✅ ALWAYS: Leverage build cache for dependencies
FROM golang:1.21-alpine AS deps
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

FROM deps AS builder
COPY . .
RUN go build -o main .
```

## Production Optimization

### Runtime Environment

```dockerfile
# ✅ ALWAYS: Set production environment variables
ENV GO_ENV=production
ENV GIN_MODE=release
ENV CGO_ENABLED=0

# Timezone configuration
ENV TZ=UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
```

## Docker Compose Integration

### Service Definition Best Practices

```yaml
# docker-compose.yml
version: "3.8"

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        - BUILD_DATE=${BUILD_DATE}
        - VERSION=${VERSION}
        - COMMIT_SHA=${COMMIT_SHA}

    # ✅ ALWAYS: Resource limits
    deploy:
      resources:
        limits:
          cpus: "0.5"
          memory: 512M
        reservations:
          cpus: "0.25"
          memory: 256M

    # ✅ ALWAYS: Restart policy
    restart: unless-stopped

    # ✅ ALWAYS: Health check
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

    # ✅ ALWAYS: Logging configuration
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

    # Environment variables
    environment:
      - GO_ENV=production
      - LOG_LEVEL=info

    # Volume mounts
    volumes:
      - app_data:/data
      - app_logs:/logs

volumes:
  app_data:
  app_logs:
```

## Security Scanning Integration

### Vulnerability Assessment

```dockerfile
# ✅ ALWAYS: Include security scan comments
# Security scanning commands:
# docker scout cve <image>
# docker scout recommendations <image>
# trivy image <image>

# Build with security context
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder
```

## Metadata and Labels

### Comprehensive Image Labeling

```dockerfile
# ✅ ALWAYS: Complete OCI labels
LABEL maintainer="team@company.com"
LABEL org.opencontainers.image.title="Application Name"
LABEL org.opencontainers.image.description="Application description"
LABEL org.opencontainers.image.url="https://company.com"
LABEL org.opencontainers.image.source="https://github.com/company/repo"
LABEL org.opencontainers.image.documentation="https://docs.company.com"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${COMMIT_SHA}"
LABEL org.opencontainers.image.vendor="Company Name"
LABEL org.opencontainers.image.licenses="MIT"
```

## Build Validation Commands

### Quality Assurance

```bash
# ✅ ALWAYS: Validate Dockerfile
hadolint Dockerfile

# ✅ ALWAYS: Security scanning
docker scout cve <image>
trivy image <image>

# ✅ ALWAYS: Size optimization check
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"

# ✅ ALWAYS: Layer analysis
docker history <image>
```

## Error Handling Pattern

```dockerfile
# ✅ ALWAYS: Handle command failures
RUN set -e && \
    command1 && \
    command2 && \
    command3

# ✅ ALWAYS: Verify critical operations
RUN go mod download && \
    go mod verify && \
    test -f go.sum
```

## Documentation Requirements

### Inline Documentation

```dockerfile
# Application: My Go Application
# Description: High-performance microservice
# Version: 1.0.0
# Build: docker build -t myapp:latest .
# Run: docker run -p 8080:8080 myapp:latest

# Build stage
FROM golang:1.21-alpine AS builder
# ... rest of Dockerfile
```

---

# Shell Script Optimization Rules (.sh files ONLY)

## POSIX Shell Priority (Mandatory)

### Default Shell Selection

```bash
#!/bin/sh
# ✅ ALWAYS: Use POSIX /bin/sh as default shebang
# Maximum compatibility across systems
# Minimal resource usage
# Available on all Unix-like systems

# ❌ AVOID: Unless absolutely necessary
#!/bin/bash
#!/bin/zsh
#!/bin/dash
```

### Shell Feature Detection

```bash
# ✅ ALWAYS: Check shell capabilities before using advanced features
check_shell_features() {
    # Test for bash-specific features
    if [ -n "$BASH_VERSION" ]; then
        HAS_BASH=1
    else
        HAS_BASH=0
    fi

    # Test for array support
    if command -v bash >/dev/null 2>&1; then
        HAS_ARRAYS=1
    else
        HAS_ARRAYS=0
    fi
}
```

## Memory Optimization

### Variable Management

```bash
# ✅ ALWAYS: Unset large variables when done
process_large_data() {
    large_data=$(cat large_file.txt)

    # Process data
    echo "$large_data" | process_command

    # Free memory immediately
    unset large_data
}

# ✅ ALWAYS: Use local variables in functions
process_file() {
    local file="$1"
    local temp_data
    local result

    temp_data=$(cat "$file")
    result=$(echo "$temp_data" | transform)
    echo "$result"

    # Variables automatically freed when function exits
}
```

### Efficient String Operations

```bash
# ✅ ALWAYS: Use parameter expansion instead of external commands
filename="/path/to/file.txt"

# Good: Parameter expansion (no external process)
basename="${filename##*/}"
dirname="${filename%/*}"
extension="${filename##*.}"
name="${filename%.*}"

# ❌ BAD: External commands (memory + process overhead)
basename=$(basename "$filename")
dirname=$(dirname "$filename")
extension=$(echo "$filename" | cut -d'.' -f2)
```

## Disk I/O Optimization

### Minimize File Operations

```bash
# ✅ ALWAYS: Read files once and store in memory
read_config_once() {
    if [ -z "$CONFIG_LOADED" ]; then
        CONFIG_DATA=$(cat config.txt)
        CONFIG_LOADED=1
        export CONFIG_DATA CONFIG_LOADED
    fi
}

# ✅ ALWAYS: Use here documents for multi-line output
generate_config() {
    cat > config.txt << 'EOF'
# Configuration file
setting1=value1
setting2=value2
setting3=value3
EOF
}

# ✅ ALWAYS: Batch file operations
process_multiple_files() {
    # Bad: Multiple separate operations
    # for file in *.txt; do
    #     cat "$file" >> combined.txt
    # done

    # Good: Single operation
    cat *.txt > combined.txt
}
```

### Efficient Log Writing

```bash
# ✅ ALWAYS: Buffer log writes
LOG_BUFFER=""
LOG_MAX_SIZE=1024

log_message() {
    local message="$1"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    LOG_BUFFER="${LOG_BUFFER}${timestamp}: ${message}\n"

    # Flush when buffer is full
    if [ ${#LOG_BUFFER} -gt $LOG_MAX_SIZE ]; then
        printf "%s" "$LOG_BUFFER" >> logfile.log
        LOG_BUFFER=""
    fi
}

# ✅ ALWAYS: Flush buffer on exit
cleanup_logs() {
    if [ -n "$LOG_BUFFER" ]; then
        printf "%s" "$LOG_BUFFER" >> logfile.log
    fi
}
trap cleanup_logs EXIT
```

## CPU Optimization

### Command Substitution Efficiency

```bash
# ✅ ALWAYS: Use $() instead of backticks
result=$(command arg1 arg2)

# ❌ AVOID: Backticks (harder to nest, less efficient)
result=`command arg1 arg2`

# ✅ ALWAYS: Minimize subshells
# Bad: Multiple subshells
count=$(echo "$data" | wc -l)
size=$(echo "$data" | wc -c)

# Good: Single operation with multiple outputs
{
    echo "$data" | wc -l
    echo "$data" | wc -c
} | {
    read count
    read size
}
```

### Efficient Loops and Conditions

```bash
# ✅ ALWAYS: Use built-in test conditions
if [ -f "$file" ] && [ -r "$file" ]; then
    process_file "$file"
fi

# ✅ ALWAYS: Avoid unnecessary command substitutions in loops
# Bad: Command substitution in each iteration
for file in $(ls *.txt); do
    process "$file"
done

# Good: Direct globbing
for file in *.txt; do
    [ -f "$file" ] || continue
    process "$file"
done

# ✅ ALWAYS: Use case for multiple string comparisons
check_file_type() {
    local file="$1"
    case "$file" in
        *.txt) echo "text file" ;;
        *.log) echo "log file" ;;
        *.conf|*.cfg) echo "config file" ;;
        *) echo "unknown file" ;;
    esac
}
```

## Error Handling (POSIX Compatible)

### Robust Error Management

```bash
#!/bin/sh
# ✅ ALWAYS: Set strict error handling
set -e  # Exit on error
set -u  # Exit on undefined variable
set -f  # Disable globbing

# ✅ ALWAYS: Define cleanup function
cleanup() {
    local exit_code=$?

    # Remove temporary files
    rm -f "$TEMP_FILE" 2>/dev/null

    # Kill background processes
    [ -n "$BACKGROUND_PID" ] && kill "$BACKGROUND_PID" 2>/dev/null

    exit $exit_code
}
trap cleanup EXIT INT TERM

# ✅ ALWAYS: Check command success explicitly
run_command() {
    local cmd="$1"
    local error_msg="$2"

    if ! $cmd; then
        echo "Error: $error_msg" >&2
        return 1
    fi
}

# ✅ ALWAYS: Validate input parameters
validate_params() {
    if [ $# -lt 1 ]; then
        echo "Usage: $0 <required_param>" >&2
        exit 1
    fi

    if [ ! -f "$1" ]; then
        echo "Error: File '$1' does not exist" >&2
        exit 1
    fi
}
```

## POSIX Compatibility Patterns

### Portable Constructs

```bash
# ✅ ALWAYS: Use POSIX-compliant syntax
# Good: POSIX parameter expansion
remove_extension() {
    local filename="$1"
    echo "${filename%.*}"
}

# Good: POSIX string operations
contains_substring() {
    local string="$1"
    local substring="$2"
    case "$string" in
        *"$substring"*) return 0 ;;
        *) return 1 ;;
    esac
}

# ✅ ALWAYS: Use portable command options
# Good: Portable find
find . -name "*.txt" -type f

# ❌ AVOID: GNU-specific options
# find . -name "*.txt" -type f -printf "%p\n"
```

### Cross-Platform Path Handling

```bash
# ✅ ALWAYS: Handle path separators portably
normalize_path() {
    local path="$1"
    # Remove duplicate slashes
    echo "$path" | sed 's|//*|/|g'
}

# ✅ ALWAYS: Use portable temporary files
create_temp_file() {
    local prefix="$1"
    local temp_dir="${TMPDIR:-/tmp}"
    local temp_file="${temp_dir}/${prefix}.$$"

    # Ensure temp file is unique
    while [ -e "$temp_file" ]; do
        temp_file="${temp_dir}/${prefix}.$$.$(date +%s)"
    done

    touch "$temp_file"
    echo "$temp_file"
}
```

## Performance Monitoring

### Script Profiling

```bash
# ✅ ALWAYS: Add timing for performance-critical sections
time_function() {
    local start_time=$(date +%s)

    # Function logic here
    "$@"

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    echo "Function completed in ${duration}s" >&2
}

# ✅ ALWAYS: Monitor resource usage
monitor_resources() {
    if command -v ps >/dev/null 2>&1; then
        ps -o pid,ppid,rss,vsz,pcpu,comm -p $$
    fi
}
```

## Concurrent Operations

### Safe Background Processing

```bash
# ✅ ALWAYS: Limit concurrent processes
MAX_JOBS=4
CURRENT_JOBS=0

process_file_async() {
    local file="$1"

    # Wait if too many jobs running
    while [ $CURRENT_JOBS -ge $MAX_JOBS ]; do
        wait_for_job_completion
    done

    {
        process_file "$file"
        echo "Completed: $file"
    } &

    CURRENT_JOBS=$((CURRENT_JOBS + 1))
}

wait_for_job_completion() {
    if jobs >/dev/null 2>&1; then
        # Wait for any background job
        wait
        CURRENT_JOBS=0
    fi
}
```

## Security Best Practices

### Input Sanitization

```bash
# ✅ ALWAYS: Sanitize file paths
sanitize_path() {
    local path="$1"

    # Remove dangerous characters
    path=$(echo "$path" | tr -d ';&|`$()')

    # Prevent directory traversal
    case "$path" in
        *../*|*/../*|../*)
            echo "Error: Invalid path" >&2
            return 1
            ;;
    esac

    echo "$path"
}

# ✅ ALWAYS: Quote variables to prevent injection
safe_exec() {
    local command="$1"
    local arg="$2"

    # Always quote arguments
    "$command" "$arg"
}
```

## Configuration Management

### Environment Variable Handling

```bash
# ✅ ALWAYS: Provide defaults for environment variables
CONFIG_FILE="${CONFIG_FILE:-/etc/default/myapp}"
LOG_LEVEL="${LOG_LEVEL:-info}"
MAX_RETRIES="${MAX_RETRIES:-3}"

# ✅ ALWAYS: Validate environment variables
validate_config() {
    # Check required variables
    for var in CONFIG_FILE LOG_LEVEL; do
        eval "value=\$$var"
        if [ -z "$value" ]; then
            echo "Error: $var is not set" >&2
            exit 1
        fi
    done

    # Validate numeric values
    case "$MAX_RETRIES" in
        ''|*[!0-9]*)
            echo "Error: MAX_RETRIES must be a number" >&2
            exit 1
            ;;
    esac
}
```

## Debugging and Maintenance

### Debug Mode Support

```bash
# ✅ ALWAYS: Support debug mode
DEBUG="${DEBUG:-0}"

debug_log() {
    if [ "$DEBUG" = "1" ]; then
        echo "DEBUG: $*" >&2
    fi
}

# ✅ ALWAYS: Provide verbose mode
VERBOSE="${VERBOSE:-0}"

verbose_log() {
    if [ "$VERBOSE" = "1" ]; then
        echo "INFO: $*" >&2
    fi
}

# Enable debug tracing when needed
if [ "$DEBUG" = "1" ]; then
    set -x
fi
```

## Script Template

### Standard Script Structure

```bash
#!/bin/sh
# Script: script_name.sh
# Description: Script description
# Version: 1.0.0
# Author: Your Name
# Usage: script_name.sh [options] <arguments>

# ✅ ALWAYS: Set strict mode
set -e
set -u
set -f

# ✅ ALWAYS: Define constants
readonly SCRIPT_NAME="$(basename "$0")"
readonly SCRIPT_DIR="$(dirname "$0")"
readonly VERSION="1.0.0"

# ✅ ALWAYS: Initialize variables
DEBUG="${DEBUG:-0}"
VERBOSE="${VERBOSE:-0}"
DRY_RUN="${DRY_RUN:-0}"

# ✅ ALWAYS: Define usage function
usage() {
    cat << EOF
Usage: $SCRIPT_NAME [OPTIONS] <argument>

Description of what the script does.

OPTIONS:
    -h, --help      Show this help message
    -v, --verbose   Enable verbose output
    -d, --debug     Enable debug mode
    -n, --dry-run   Show what would be done without executing

EXAMPLES:
    $SCRIPT_NAME file.txt
    $SCRIPT_NAME -v file.txt

EOF
}

# ✅ ALWAYS: Parse command line arguments
parse_args() {
    while [ $# -gt 0 ]; do
        case "$1" in
            -h|--help)
                usage
                exit 0
                ;;
            -v|--verbose)
                VERBOSE=1
                ;;
            -d|--debug)
                DEBUG=1
                set -x
                ;;
            -n|--dry-run)
                DRY_RUN=1
                ;;
            -*)
                echo "Error: Unknown option $1" >&2
                usage >&2
                exit 1
                ;;
            *)
                break
                ;;
        esac
        shift
    done
}

# ✅ ALWAYS: Define cleanup function
cleanup() {
    local exit_code=$?

    # Cleanup temporary files
    [ -n "${TEMP_FILES:-}" ] && rm -f $TEMP_FILES

    # Kill background processes
    [ -n "${BACKGROUND_PIDS:-}" ] && kill $BACKGROUND_PIDS 2>/dev/null || true

    exit $exit_code
}
trap cleanup EXIT INT TERM

# ✅ ALWAYS: Main function
main() {
    parse_args "$@"

    # Validate requirements
    validate_environment

    # Main logic here
    echo "Script execution completed successfully"
}

# ✅ ALWAYS: Validate environment
validate_environment() {
    # Check required commands
    for cmd in cat sed grep; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            echo "Error: Required command '$cmd' not found" >&2
            exit 1
        fi
    done
}

# Execute main function with all arguments
main "$@"
```

## Validation Commands

### Script Quality Checks

```bash
# ✅ ALWAYS: Validate shell syntax
shellcheck script.sh

# ✅ ALWAYS: Test POSIX compliance
checkbashisms script.sh

# ✅ ALWAYS: Performance testing
time ./script.sh test_input

# ✅ ALWAYS: Memory usage monitoring
/usr/bin/time -v ./script.sh test_input
```
````
