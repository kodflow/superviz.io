````instructions
[0;34m[INFO][0m Processing section: 01-prime-directive.md
# Copilot Operational Instructions - Go 1.24 Ultra-Performance Expert Mode

## Prime Directive

Copilot operates as a senior Go 1.24+ engineer with extreme performance focus. ALL suggestions must be:

- **Go 1.24+ features**: leverage yield iterators, enhanced atomic operations, and latest optimizations
- **Zero-allocation first**: every line optimized for minimal CPU/RAM consumption
- **Atomic operations**: use `sync/atomic` for all counters, flags, and lock-free operations
- **100% test coverage**: comprehensive, mockable tests with mandatory timeouts
- **Proactive security**: anticipate and prevent vulnerabilities before they occur
- **Holistic integration**: consider entire project context for optimal integration

## Performance Measurement Policy

- **On-demand profiling**: Only create benchmarks when explicitly requested
- **Profile-driven optimization**: Measure only when user asks for performance analysis
- **Default behavior**: Focus on clean, optimized code without automatic benchmarking
- **Benchmark triggers**: Create benchmarks only when user mentions "benchmark", "profile", or "measure performance"

## The Three Optimization Questions (go-perfbook Framework)

Before suggesting ANY optimization, ALWAYS apply this framework in order:

1. **Do we have to do this at all?** - The fastest code is code never executed
2. **If yes, is this the best algorithm?** - Focus on algorithmic improvements first
3. **If yes, is this the best implementation?** - Only then optimize implementation details

This framework prevents premature optimization while ensuring we address bottlenecks at the right level.

## Optimization Workflow (go-perfbook Integration)

- **Amdahl's Law**: Focus on bottlenecks - 80% speedup on 5% code = 2.5% total gain
- **Constant factors matter**: Same Big-O doesn't mean same performance
- **Know your input sizes**: Choose algorithms based on realistic data sizes
- **Space-time trade-offs**: Understand where you are on the memory/performance curve

## File Edit Strategy

- **Single file focus**: Never edit more than one file at a time
- **Large file handling**: For files >300 lines, propose detailed edit plan first
- **Context awareness**: Always analyze entire project structure before suggesting changes

### Mandatory Edit Plan Format

```text
## PROPOSED EDIT PLAN
Target file: [filename]
Project impact analysis: [how this affects other files/packages]
Total planned edits: [number]
Performance impact: [expected CPU/memory improvements]

Edit sequence:
1. [Change description] - Purpose: [performance/security/testability reason]
2. [Change description] - Purpose: [performance/security/testability reason]
...

Dependencies affected: [list of files that may need updates]
Test files to update: [corresponding test files]
```

Wait for explicit user approval before executing ANY edits.

---

[0;34m[INFO][0m Processing section: 02-go-1-24-performance-standards.md
## Go 1.24+ Ultra-Performance Standards

### Go 1.24 Core Features (Mandatory)

```go
// Generic type aliases - NEW in 1.24
type DataProcessor[T any] interface {
    Process(T) (T, error)
}

// Create generic alias for common patterns
type StringProcessor[T ~string] = DataProcessor[T]
type NumberProcessor[T ~int | ~float64] = DataProcessor[T]

// Directory-limited filesystem access - NEW in 1.24
func processFilesSecurely(rootPath string) error {
    root, err := os.OpenRoot(rootPath)
    if err != nil {
        return fmt.Errorf("failed to open root directory %q: %w", rootPath, err)
    }
    defer root.Close()

    file, err := root.Create("output.dat")
    if err != nil {
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()

    return nil
}

// New benchmark Loop method - NEW in 1.24
func BenchmarkProcessor(b *testing.B) {
    processor := NewProcessor()
    data := generateTestData()

    for b.Loop() {
        processor.Process(data)
    }
}

// Built-in min/max functions - NEW in 1.24 (eliminate boilerplate)
func clampValue(value, minVal, maxVal int) int {
    return min(max(value, minVal), maxVal) // No more custom min/max functions
}

// Weak pointers for smart caching - NEW in 1.24
import "runtime/arena"

type CacheItem struct {
    Data      []byte
    Timestamp time.Time
}

type SmartCache struct {
    arena *arena.Arena
    cache map[string]arena.Weak[CacheItem]
    mutex sync.RWMutex
}

func NewSmartCache() *SmartCache {
    return &SmartCache{
        arena: arena.NewArena(),
        cache: make(map[string]arena.Weak[CacheItem]),
    }
}

func (c *SmartCache) Set(key string, item CacheItem) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    // Allocate in arena for memory efficiency
    allocated := arena.New[CacheItem](c.arena)
    *allocated = item

    // Store weak reference - allows GC when memory pressure high
    c.cache[key] = arena.MakeWeak[CacheItem](allocated)
}

func (c *SmartCache) Get(key string) (CacheItem, bool) {
    c.mutex.RLock()
    weak, exists := c.cache[key]
    c.mutex.RUnlock()

    if !exists {
        return CacheItem{}, false
    }

    // Try to get strong reference
    if strong := weak.Strong(); strong != nil {
        return *strong, true
    }

    // Item was garbage collected
    c.mutex.Lock()
    delete(c.cache, key)
    c.mutex.Unlock()

    return CacheItem{}, false
}

// Enhanced benchmark Loop method - NEW in 1.24
func BenchmarkProcessorOptimal(b *testing.B) {
    processor := NewProcessor()
    data := generateTestData()

    b.ResetTimer()
    for b.Loop() { // More accurate timing than old for-loop
        processor.Process(data)
    }
}

// Improved DWARF debugging - NEW in 1.24
// Use build tags for debug-optimized builds
//go:build debug

func debugOptimizedFunction() {
    // Better debugging information in Go 1.24
    // Improved variable inspection and stack traces
}
```

### Atomic Operations (Mandatory)

```go
// ALWAYS use atomic for counters/flags
var (
    requestCount atomic.Uint64
    isShutdown   atomic.Bool
    lastUpdate   atomic.Int64
)

// Increment pattern
requestCount.Add(1)

// Flag pattern
if isShutdown.Load() {
    return ErrShutdown
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

### Go 1.24 Iterators with Yield

```go
// Use yield for memory-efficient iteration
func (s *Store) Items() iter.Seq2[string, *Item] {
    return func(yield func(string, *Item) bool) {
        s.mu.RLock()
        defer s.mu.RUnlock()

        for k, v := range s.data {
            if !yield(k, v) {
                return
            }
        }
    }
}

// Usage with range-over-func
for key, item := range store.Items() {
    process(key, item)
}
```

### Zero-Allocation Patterns (Enforced)

```go
// Pre-allocate with exact capacity
items := make([]Item, 0, len(source))

// String building with pre-allocation
var builder strings.Builder
builder.Grow(estimatedSize) // ALWAYS pre-grow
builder.WriteString(part1)
builder.WriteString(part2)

// Byte slice reuse with sync.Pool
var bufPool = sync.Pool{
    New: func() any {
        return make([]byte, 0, 4096)
    },
}

func process() []byte {
    buf := bufPool.Get().([]byte)
    defer func() {
        buf = buf[:0] // Reset length
        bufPool.Put(buf)
    }()
    // Use buf
}

// Map pre-allocation
cache := make(map[string]*Item, estimatedCount)
```

### Structured Concurrency Patterns (Anti-Fire-and-Forget)

```go
// ❌ DANGEROUS: Fire-and-forget goroutines (from production incidents)
func BadAsyncOperation() {
    go func() {
        // Unsupervised goroutine - errors disappear, no cleanup, no observability
        riskyOperation()
    }()
    // Function returns immediately, no idea what happened
}

// ✅ GOOD: Supervised concurrency with context and error handling
func GoodAsyncOperation(ctx context.Context) error {
    errChan := make(chan error, 1)

    go func() {
        defer func() {
            if r := recover(); r != nil {
                errChan <- fmt.Errorf("panic in async operation: %v", r)
            }
        }()

        err := riskyOperation()
        errChan <- err
    }()

    select {
    case err := <-errChan:
        return err
    case <-ctx.Done():
        return ctx.Err()
    }
}

// ✅ BETTER: Worker pool with structured lifecycle
type SupervisedWorkerPool struct {
    workers   int
    workChan  chan Work
    resultChan chan Result

    ctx       context.Context
    cancel    context.CancelFunc
    wg        sync.WaitGroup

    // Observability
    activeJobs atomic.Int64
    totalJobs  atomic.Int64
    errors     atomic.Int64
}

func NewSupervisedWorkerPool(workers int) *SupervisedWorkerPool {
    ctx, cancel := context.WithCancel(context.Background())

    pool := &SupervisedWorkerPool{
        workers:    workers,
        workChan:   make(chan Work, workers*2),
        resultChan: make(chan Result, workers*2),
        ctx:        ctx,
        cancel:     cancel,
    }

    // Start supervised workers
    for i := 0; i < workers; i++ {
        pool.wg.Add(1)
        go pool.supervisedWorker(i)
    }

    return pool
}

func (p *SupervisedWorkerPool) supervisedWorker(id int) {
    defer p.wg.Done()

    for {
        select {
        case work := <-p.workChan:
            p.activeJobs.Add(1)
            p.totalJobs.Add(1)

            func() {
                defer func() {
                    p.activeJobs.Add(-1)
                    if r := recover(); r != nil {
                        p.errors.Add(1)
                        log.Printf("Worker %d recovered from panic: %v", id, r)
                    }
                }()

                result := work.Execute()
                if result.Error != nil {
                    p.errors.Add(1)
                }

                select {
                case p.resultChan <- result:
                case <-p.ctx.Done():
                }
            }()

        case <-p.ctx.Done():
            return
        }
    }
}

func (p *SupervisedWorkerPool) Submit(work Work) error {
    select {
    case p.workChan <- work:
        return nil
    case <-p.ctx.Done():
        return p.ctx.Err()
    default:
        return fmt.Errorf("worker pool queue full")
    }
}

func (p *SupervisedWorkerPool) Shutdown(timeout time.Duration) error {
    p.cancel()

    done := make(chan struct{})
    go func() {
        p.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        return nil
    case <-time.After(timeout):
        return fmt.Errorf("shutdown timeout exceeded")
    }
}

func (p *SupervisedWorkerPool) Stats() (active, total, errors int64) {
    return p.activeJobs.Load(), p.totalJobs.Load(), p.errors.Load()
}
```

### Channel Safety (Proactive Bug Prevention)

```go
// ALWAYS use buffered channels with explicit capacity
ch := make(chan Message, 100) // Never unbuffered unless proven necessary

// Mandatory channel patterns
func (w *Worker) Start(ctx context.Context) error {
    // Always check if already started
    if !w.started.CompareAndSwap(false, true) {
        return ErrAlreadyStarted
    }

    go func() {
        defer w.started.Store(false)
        defer close(w.done) // Signal completion

        for {
            select {
            case msg := <-w.input:
                w.process(msg)
            case <-ctx.Done():
                return // Clean shutdown
            case <-w.shutdown:
                return // Explicit shutdown
            }
        }
    }()

    return nil
}

// Channel closing pattern
func (w *Worker) Stop() error {
    select {
    case w.shutdown <- struct{}{}:
        // Signal sent successfully
    case <-time.After(time.Second):
        return ErrShutdownTimeout
    }

    select {
    case <-w.done:
        return nil // Clean shutdown
    case <-time.After(5 * time.Second):
        return ErrForceShutdown
    }
}
```

### The Three Optimization Questions (go-perfbook Core Framework)

Before any optimization, ALWAYS ask these three questions in order:

```go
// Question 1: Do we have to do this at all?
func optimizeDataProcessing(data []Record) []Result {
    // Check cache first - fastest code is code never run
    if cached := checkCache(data); cached != nil {
        return cached // Skip processing entirely
    }

    // Question 2: Is this the best algorithm?
    if len(data) < 100 {
        return simpleLinearProcess(data) // O(n) but low constant factor
    } else {
        return efficientDivideConquer(data) // O(n log n) but higher constant factor
    }

    // Question 3: Best implementation handled in individual functions
}
```

### Big-O Awareness and Algorithm Selection (go-perfbook)

```go
// O(1): Field access, array/map lookup - don't worry about it
func getUser(users map[string]*User, id string) *User {
    return users[id] // O(1) - acceptable anywhere
}

// O(log n): Binary search - only a problem if in tight loop
func findInSorted(data []string, target string) int {
    return sort.SearchStrings(data, target) // O(log n) - usually fine
}

// O(n): Simple loop - very common, usually acceptable
func sumSlice(data []int) int {
    var sum int
    for _, v := range data { // O(n) - standard operation
        sum += v
    }
    return sum
}

// O(n log n): Divide-and-conquer, sorting - still fairly fast
func processAndSort(data []Item) []Item {
    // O(n) processing + O(n log n) sorting = O(n log n) total
    for i := range data {
        data[i] = process(data[i])
    }
    sort.Slice(data, func(i, j int) bool {
        return data[i].Priority < data[j].Priority
    })
    return data
}

// O(n²): Nested loops - be careful with large datasets
func findDuplicates(data []string) []string {
    var duplicates []string
    for i := 0; i < len(data); i++ {
        for j := i + 1; j < len(data); j++ { // O(n²) - constrain dataset size
            if data[i] == data[j] {
                duplicates = append(duplicates, data[i])
            }
        }
    }
    return duplicates
}

// Size-aware algorithm selection
func searchStrategy(data []Item, target string) int {
    switch {
    case len(data) <= 10:
        return linearSearch(data, target) // Better cache locality
    case len(data) <= 1000:
        return binarySearch(data, target) // Needs sorted data
    default:
        return hashLookup(data, target) // Build map first
    }
}

// Polyalgorithm pattern - detect input characteristics
func sortAdaptive(data []int) {
    if isAlreadySorted(data) {
        return // O(n) check saves O(n log n) work
    }

    if len(data) < 12 {
        insertionSort(data) // Fastest for small arrays
    } else if isAlmostSorted(data) {
        insertionSort(data) // Excellent for nearly sorted data
    } else {
        quickSort(data) // General purpose
    }
}

// Constant factor optimization - same Big-O doesn't mean same performance
func fastContains(slice []string, target string) bool {
    // Manual unrolling for small slices (constant factor improvement)
    switch len(slice) {
    case 0:
        return false
    case 1:
        return slice[0] == target
    case 2:
        return slice[0] == target || slice[1] == target
    case 3:
        return slice[0] == target || slice[1] == target || slice[2] == target
    default:
        // Fall back to loop for larger slices
        for _, s := range slice {
            if s == target {
                return true
            }
        }
        return false
    }
}

// Branchless optimization for predictable patterns
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
    return counts[0], counts[1]
}
```

### Production-Scale Optimization Patterns (2M+ Users)

```go
// Global HTTP client with connection pooling (avoid creating per request)
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        MaxConnsPerHost:     50,
        IdleConnTimeout:     90 * time.Second,
        TLSHandshakeTimeout: 10 * time.Second,
        DisableKeepAlives:   false, // Enable keep-alive
    },
    Timeout: 30 * time.Second,
}

// Production HTTP server configuration
func NewProductionServer(handler http.Handler, port string) *http.Server {
    return &http.Server{
        Addr:           ":" + port,
        Handler:        handler,
        ReadTimeout:    10 * time.Second,  // Prevent slow clients
        WriteTimeout:   10 * time.Second,  // Prevent slow responses
        IdleTimeout:    120 * time.Second, // Keep-alive timeout
        MaxHeaderBytes: 1 << 20,           // 1MB max headers
    }
}

// Structured JSON unmarshaling (10x faster than map[string]interface{})
type User struct {
    ID    uint64 `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Role  string `json:"role"`
}

// Bad: Using untyped map (slow, runtime overhead)
// var data map[string]interface{}
// json.Unmarshal(rawBytes, &data)

// Good: Direct unmarshaling to struct
func parseUser(rawBytes []byte) (*User, error) {
    user := userPool.Get().(*User)
    if err := json.Unmarshal(rawBytes, user); err != nil {
        userPool.Put(user)
        return nil, err
    }
    return user, nil
}

// Worker pool to cap goroutines (prevent memory bloat)
type WorkerPool struct {
    workers    chan struct{}
    workerFunc func(interface{})
    jobQueue   chan interface{}
    quit       chan struct{}
}

func NewWorkerPool(maxWorkers int, workerFunc func(interface{})) *WorkerPool {
    return &WorkerPool{
        workers:    make(chan struct{}, maxWorkers), // Semaphore pattern
        workerFunc: workerFunc,
        jobQueue:   make(chan interface{}, maxWorkers*2), // Buffered queue
        quit:       make(chan struct{}),
    }
}

func (wp *WorkerPool) Submit(job interface{}) error {
    select {
    case wp.jobQueue <- job:
        return nil
    case <-wp.quit:
        return ErrPoolClosed
    default:
        return ErrPoolFull
    }
}

func (wp *WorkerPool) Start() {
    for {
        select {
        case job := <-wp.jobQueue:
            // Acquire worker slot
            wp.workers <- struct{}{}
            go func(j interface{}) {
                defer func() { <-wp.workers }() // Release worker slot
                wp.workerFunc(j)
            }(job)
        case <-wp.quit:
            return
        }
    }
}

// Database batch operations (3-5x faster than individual inserts)
type BatchInserter struct {
    db        *sql.DB
    stmt      *sql.Stmt
    batchSize int
    pending   []User
    timer     *time.Timer
}

func NewBatchInserter(db *sql.DB, batchSize int) *BatchInserter {
    stmt, _ := db.Prepare("INSERT INTO users (id, name, email) VALUES ($1, $2, $3)")

    bi := &BatchInserter{
        db:        db,
        stmt:      stmt,
        batchSize: batchSize,
        pending:   make([]User, 0, batchSize),
        timer:     time.NewTimer(time.Second), // Flush every second
    }

    go bi.flushPeriodically()
    return bi
}

func (bi *BatchInserter) Add(user User) error {
    bi.pending = append(bi.pending, user)

    if len(bi.pending) >= bi.batchSize {
        return bi.flush()
    }

    // Reset timer for periodic flush
    bi.timer.Reset(time.Second)
    return nil
}

func (bi *BatchInserter) flush() error {
    if len(bi.pending) == 0 {
        return nil
    }

    tx, err := bi.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    for _, user := range bi.pending {
        if _, err := tx.Stmt(bi.stmt).Exec(user.ID, user.Name, user.Email); err != nil {
            return err
        }
    }

    if err := tx.Commit(); err != nil {
        return err
    }

    bi.pending = bi.pending[:0] // Reset slice, keep capacity
    return nil
}

// Avoid interface{} in hot paths - use concrete types
// Bad: func processData(data interface{}) - runtime overhead
// Good: Use generics or concrete types
func processUser(user *User) error {
    // Direct field access, no type assertions
    if user.ID == 0 {
        return ErrInvalidID
    }
    return validateUser(user)
}

// Byte buffer pooling for request/response processing
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024) // Start with 1KB capacity
    },
}

func processRequest(w http.ResponseWriter, r *http.Request) {
    buf := bufferPool.Get().([]byte)
    defer func() {
        buf = buf[:0] // Reset length, keep capacity
        bufferPool.Put(buf)
    }()

    // Use buf for processing
    buf = append(buf, r.Header.Get("Content-Type")...)
    // ... processing logic
}
```

---

[0;34m[INFO][0m Processing section: 03-documentation-format.md
## Mandatory Documentation Format

### Language Requirement

**ALL comments, documentation, and code must be written in English ONLY.**

- **Godoc comments**: MUST be in English
- **Inline comments**: MUST be in English
- **Variable names**: MUST use English words
- **Function names**: MUST use English words
- **Error messages**: MUST be in English
- **Log messages**: MUST be in English
- **Test names**: MUST be in English

**No exceptions.** French, or any other language, is strictly forbidden in code.

### Documentation Format

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

---

[0;34m[INFO][0m Processing section: 04-optimization-workflow.md
# Optimization Workflow (go-perfbook Integration)

## Decision Framework

### The Three Questions Pattern

Every optimization suggestion must follow this framework:

```go
// Question 1: Do we have to do this at all?
func optimizeProcess(data []Record) []Result {
    // Eliminate work first
    if len(data) == 0 {
        return nil // Fast path for empty input
    }

    // Check cache - fastest code is never run
    if cached := checkCache(data); cached != nil {
        return cached
    }

    // Question 2: Best algorithm for this input size?
    if len(data) < 100 {
        return linearProcess(data) // O(n) but low constant factor
    } else {
        return divideConquer(data) // O(n log n) but justified for large inputs
    }

    // Question 3: Best implementation handled in specific functions
}
```

### Input Size-Aware Algorithm Selection

```go
// Adaptive algorithms based on realistic input characteristics
func searchOptimal(data []Item, target string) int {
    switch {
    case len(data) <= 10:
        // Linear search: better cache locality, no setup cost
        for i, item := range data {
            if item.Name == target {
                return i
            }
        }
        return -1

    case len(data) <= 1000:
        // Binary search: requires sorted data, O(log n)
        return sort.Search(len(data), func(i int) bool {
            return data[i].Name >= target
        })

    default:
        // Hash lookup: O(1) average, worth setup cost for large datasets
        return buildHashAndSearch(data, target)
    }
}

// Polyalgorithm pattern - detect input characteristics
func sortAdaptive(data []int) {
    // Question 1: Do we need to sort at all?
    if isAlreadySorted(data) {
        return // O(n) check saves O(n log n) work
    }

    // Question 2: Best algorithm for this data?
    switch {
    case len(data) < 12:
        insertionSort(data) // Fastest for tiny arrays
    case isAlmostSorted(data):
        insertionSort(data) // Excellent for nearly sorted
    case len(data) < 1000:
        quickSort(data) // Good general purpose
    default:
        parallelSort(data) // Worth overhead for large data
    }
}
```

## Constant Factor Optimization

### Branch Prediction Patterns

```go
// Likelihood-ordered conditions
func validateInput(input string) error {
    // Most common case first (80% of calls)
    if input != "" {
        // Fast path for valid input
        return nil
    }

    // Less common cases
    if len(input) > maxLength {
        return ErrTooLong
    }

    return ErrEmpty
}

// Branchless optimization for predictable patterns
func minMax(a, b int) (min, max int) {
    // Branchless comparison using bit manipulation
    diff := a - b
    mask := diff >> 31 // -1 if a < b, 0 if a >= b
    min = b + (diff & mask)
    max = a - (diff & mask)
    return min, max
}

// Array-based state machine to avoid branches
func processStates(states []State) {
    // Pre-computed action table
    var actions [StateCount]func(State) State
    actions[StateInit] = handleInit
    actions[StateProcess] = handleProcess
    actions[StateDone] = handleDone

    for i, state := range states {
        states[i] = actions[state.Type](state) // No branch prediction needed
    }
}
```

### Cache-Aware Data Layout

```go
// Structure field ordering for cache efficiency
type OptimizedStruct struct {
    // Hot data first (frequently accessed together)
    counter atomic.Uint64   // 8 bytes
    flags   uint32          // 4 bytes
    active  bool           // 1 byte + 3 padding = 8 bytes total

    // Separate cache line for atomic to prevent false sharing
    _       [7]uint64      // Padding to 64-byte boundary

    // Cold data last (rarely accessed)
    name        string     // 16 bytes
    description string     // 16 bytes
    metadata    map[string]interface{} // 8 bytes pointer
}

// Memory layout for sequential access
type VectorizedData struct {
    // Structure of Arrays (SoA) instead of Array of Structures (AoS)
    ids        []uint64    // All IDs together for vectorization
    timestamps []uint64    // All timestamps together
    values     []float64   // All values together for SIMD

    // Single allocation for better cache locality
    storage []byte         // Backing store for all arrays
}

func NewVectorizedData(capacity int) *VectorizedData {
    // Calculate total memory needed
    totalSize := capacity * (8 + 8 + 8) // uint64 + uint64 + float64
    storage := make([]byte, totalSize)

    // Slice the backing store
    ids := (*[1 << 30]uint64)(unsafe.Pointer(&storage[0]))[:capacity:capacity]
    timestamps := (*[1 << 30]uint64)(unsafe.Pointer(&storage[capacity*8]))[:capacity:capacity]
    values := (*[1 << 30]float64)(unsafe.Pointer(&storage[capacity*16]))[:capacity:capacity]

    return &VectorizedData{
        ids:        ids,
        timestamps: timestamps,
        values:     values,
        storage:    storage,
    }
}
```

## Specialization Strategies

### Context-Aware Optimization

```go
// Specialized implementations based on usage patterns
type TimeParser struct {
    // Single-item cache for log parsing (temporal locality)
    lastFormat string
    lastTime   time.Time
    lastLayout string

    // Statistics for adaptive behavior
    cacheHits   atomic.Uint64
    cacheMisses atomic.Uint64
}

func (tp *TimeParser) Parse(value, format string) (time.Time, error) {
    // Question 1: Can we avoid parsing entirely?
    if format == tp.lastFormat {
        tp.cacheHits.Add(1)
        return tp.lastTime, nil
    }

    // Question 2: Is there a faster algorithm for this specific format?
    if isStandardFormat(format) {
        return parseOptimized(value, format)
    }

    // Question 3: Best implementation of general parser
    tp.cacheMisses.Add(1)
    layout := compileFormat(format)
    tp.lastFormat = format
    tp.lastLayout = layout

    result, err := time.Parse(layout, value)
    if err == nil {
        tp.lastTime = result
    }

    return result, err
}

// Custom parser for known format - 50x faster than general parser
func parseLogTimestamp(s string) (time.Time, error) {
    // Specialized for: "2006-01-02 15:04:05"
    if len(s) != 19 {
        return time.Time{}, ErrInvalidFormat
    }

    // Extract components using fixed offsets (no regex, no general parsing)
    year := parseInt4(s[0:4])
    month := parseInt2(s[5:7])
    day := parseInt2(s[8:10])
    hour := parseInt2(s[11:13])
    minute := parseInt2(s[14:16])
    second := parseInt2(s[17:19])

    return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC), nil
}
```

### Workload-Specific Data Structures

```go
// Multi-level optimization based on access patterns
type AdaptiveCache struct {
    // Level 1: Single-item cache (90% hit rate for temporal locality)
    lastKey   string
    lastValue interface{}

    // Level 2: Small LRU cache (8% additional hit rate)
    recent map[string]*cacheEntry
    lru    *list.List

    // Level 3: Bloom filter for negative lookups (remaining 2%)
    bloom *BloomFilter

    // Statistics for monitoring effectiveness
    l1Hits atomic.Uint64
    l2Hits atomic.Uint64
    misses atomic.Uint64
}

func (ac *AdaptiveCache) Get(key string) (interface{}, bool) {
    // Level 1: Check single-item cache first
    if key == ac.lastKey {
        ac.l1Hits.Add(1)
        return ac.lastValue, true
    }

    // Level 2: Check LRU cache
    if entry, exists := ac.recent[key]; exists {
        ac.l2Hits.Add(1)
        ac.lru.MoveToFront(entry.element)
        ac.lastKey = key
        ac.lastValue = entry.value
        return entry.value, true
    }

    // Level 3: Check bloom filter before expensive lookup
    if !ac.bloom.MightContain([]byte(key)) {
        ac.misses.Add(1)
        return nil, false // Definitely not present
    }

    // Expensive lookup only if bloom filter says "maybe"
    return ac.expensiveLookup(key)
}
```

## Performance Measurement Integration

### Benchmarking Strategy

```go
// Embedded performance monitoring
type PerformanceMetrics struct {
    operationCount atomic.Uint64
    totalDuration  atomic.Uint64 // nanoseconds
    slowCalls      atomic.Uint64 // calls > threshold

    histogram [10]atomic.Uint64 // latency distribution
}

func (pm *PerformanceMetrics) RecordOperation(duration time.Duration) {
    pm.operationCount.Add(1)
    pm.totalDuration.Add(uint64(duration))

    if duration > slowThreshold {
        pm.slowCalls.Add(1)
    }

    // Update histogram
    bucket := min(9, int(duration.Milliseconds()))
    pm.histogram[bucket].Add(1)
}

// Adaptive optimization based on runtime metrics
func (s *Service) processData(data []byte) error {
    start := time.Now()
    defer func() {
        s.metrics.RecordOperation(time.Since(start))

        // Adaptive algorithm selection based on performance
        avgDuration := s.metrics.AverageDuration()
        if avgDuration > degradationThreshold {
            s.switchToFasterAlgorithm()
        }
    }()

    return s.algorithm.Process(data)
}
```

## Summary: Intelligent Optimization Framework

This workflow ensures optimizations are:

1. **Measured**: Profile-driven decisions with concrete metrics
2. **Targeted**: Focus on actual bottlenecks, not perceived ones
3. **Adaptive**: Choose algorithms based on real input characteristics
4. **Maintainable**: Simple solutions preferred when performance is equivalent
5. **Holistic**: Consider entire system impact, not just local optimizations

Every optimization suggestion must demonstrate understanding of:

- Where we are on the space-time trade-off curve
- Input size characteristics and their performance implications
- Constant factors vs algorithmic complexity
- Cache behavior and memory hierarchy effects
- Maintenance cost vs performance gain trade-offs

## Benchmark Creation Policy

### When to Create Benchmarks

**ONLY create benchmarks when explicitly requested by:**

- User mentions "benchmark" in their request
- User asks to "measure performance"
- User requests "profiling" or "performance analysis"
- User asks to "compare performance" between implementations

### Default Behavior

- Focus on writing optimized code without benchmarks
- Apply optimization patterns based on established best practices
- Use go-perfbook principles without requiring measurement
- Create clean, performant code that follows proven patterns

### Example Triggers for Benchmarking

```go
// ✅ CREATE BENCHMARKS - User explicitly requested
// "Can you benchmark this function?"
// "I need performance measurements for this code"
// "Compare the performance of these two approaches"

// ❌ NO BENCHMARKS - Regular optimization request
// "Optimize this function"
// "Make this code faster"
// "Improve performance"
```

---

[0;34m[INFO][0m Processing section: 05-cpu-optimization.md
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

---

[0;34m[INFO][0m Processing section: 06-disk-optimization.md
## Disk Optimization (Zero I/O Priority)

### Zero I/O Strategy

- **Memory-first approach**: Keep everything in memory when possible
- **Lazy loading**: Load data only when absolutely necessary
- **In-memory caching**: Use sync.Map, atomic values, or custom caches
- **Batch operations**: When I/O is unavoidable, batch multiple operations

```go
// Service represents a service with in-memory cache
type Service struct {
    cache *MemoryCache
}

// In-memory cache to avoid disk reads
type MemoryCache struct {
    data   sync.Map
    stats  atomic.Int64 // Cache hits
    misses atomic.Int64 // Cache misses
}

func (c *MemoryCache) Get(key string) ([]byte, bool) {
    if value, ok := c.data.Load(key); ok {
        c.stats.Add(1) // Hit
        return value.([]byte), true
    }
    c.misses.Add(1) // Miss
    return nil, false
}

// Avoid disk I/O by pre-loading everything at startup
func NewServiceWithPreload(dataDir string) (*Service, error) {
    cache := &MemoryCache{}

    // Pre-load all data into memory at startup
    err := filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() {
            return err
        }

        data, err := os.ReadFile(path)
        if err != nil {
            return err
        }

        key := strings.TrimPrefix(path, dataDir)
        cache.data.Store(key, data)
        return nil
    })

    if err != nil {
        return nil, fmt.Errorf("failed to preload data: %w", err)
    }

    return &Service{cache: cache}, nil
}
```

### Memory-Mapped Files (When I/O Required)

- **Use mmap for large files**: Avoid read() syscalls
- **Read-only mappings**: For configuration and static data
- **Sequential access patterns**: Optimize for page faults

```go
// Memory-mapped file access
func readLargeFileOptimal(filename string) ([]byte, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    stat, err := file.Stat()
    if err != nil {
        return nil, err
    }

    // For large files, use mmap to avoid copying data
    if stat.Size() > 1024*1024 { // > 1MB
        return syscall.Mmap(int(file.Fd()), 0, int(stat.Size()),
            syscall.PROT_READ, syscall.MAP_SHARED)
    }

    // For small files, regular read is fine
    return io.ReadAll(file)
}
```

### Buffered I/O Optimization

- **Large buffer sizes**: Reduce syscall frequency
- **Write batching**: Accumulate writes before flushing
- **Async I/O patterns**: Use goroutines for I/O operations

```go
// Optimal buffered writer
type OptimalWriter struct {
    file   *os.File
    buffer *bufio.Writer
    size   int64

    // Async flushing
    flushChan chan struct{}
    doneChan  chan struct{}
}

func NewOptimalWriter(filename string) (*OptimalWriter, error) {
    file, err := os.Create(filename)
    if err != nil {
        return nil, err
    }

    // Large buffer to reduce syscalls (64KB)
    buffer := bufio.NewWriterSize(file, 64*1024)

    w := &OptimalWriter{
        file:      file,
        buffer:    buffer,
        flushChan: make(chan struct{}, 1),
        doneChan:  make(chan struct{}),
    }

    // Async flush goroutine
    go w.asyncFlusher()

    return w, nil
}

func (w *OptimalWriter) Write(data []byte) error {
    n, err := w.buffer.Write(data)
    if err != nil {
        return err
    }

    atomic.AddInt64(&w.size, int64(n))

    // Trigger async flush if buffer is getting full
    if w.buffer.Available() < len(data) {
        select {
        case w.flushChan <- struct{}{}:
        default: // Don't block if flush is already pending
        }
    }

    return nil
}

func (w *OptimalWriter) asyncFlusher() {
    ticker := time.NewTicker(time.Second) // Periodic flush
    defer ticker.Stop()

    for {
        select {
        case <-w.flushChan:
            w.buffer.Flush()
        case <-ticker.C:
            w.buffer.Flush()
        case <-w.doneChan:
            w.buffer.Flush()
            return
        }
    }
}
```

### Directory and File System Optimization

- **Minimize stat() calls**: Cache file information
- **Batch directory operations**: Use ReadDir instead of multiple Stat calls
- **Avoid file existence checks**: Use direct open with error handling

```go
// Avoid multiple stat calls
func processDirectoryOptimal(dirPath string) error {
    // Good: Single ReadDir call gets all file info
    entries, err := os.ReadDir(dirPath)
    if err != nil {
        return err
    }

    for _, entry := range entries {
        info, err := entry.Info()
        if err != nil {
            continue
        }

        // Process file info without additional stat calls
        processFileInfo(info)
    }

    return nil
}

func processFileInfo(info os.FileInfo) {
    // Process the file info (placeholder implementation)
    _ = info.Name()
    _ = info.Size()
    _ = info.ModTime()
}

// Avoid existence checks
func writeFileOptimal(filename string, data []byte) error {
    // Bad: Check if file exists first
    // if _, err := os.Stat(filename); os.IsNotExist(err) { ... }

    // Good: Direct open, handle error appropriately
    file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
    if err != nil {
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()

    _, err = file.Write(data)
    return err
}
}
```

### Batching Strategies (go-perfbook Pattern)

- **Reduce syscall overhead**: Batch multiple operations into single calls
- **Amortize expensive operations**: Spread cost across multiple requests
- **Buffer size optimization**: Find sweet spot between memory usage and I/O efficiency
- **Async batching**: Use background goroutines to batch operations

```go
// Efficient batched writes
type BatchWriter struct {
    file     *os.File
    buffer   []byte
    pending  [][]byte
    maxBatch int
    timeout  time.Duration

    flushChan chan struct{}
    stopChan  chan struct{}
}

func NewBatchWriter(filename string, maxBatch int) (*BatchWriter, error) {
    file, err := os.Create(filename)
    if err != nil {
        return nil, err
    }

    bw := &BatchWriter{
        file:      file,
        buffer:    make([]byte, 0, 64*1024), // 64KB buffer
        pending:   make([][]byte, 0, maxBatch),
        maxBatch:  maxBatch,
        timeout:   time.Millisecond * 100,
        flushChan: make(chan struct{}, 1),
        stopChan:  make(chan struct{}),
    }

    go bw.batchProcessor()
    return bw, nil
}

func (bw *BatchWriter) Write(data []byte) error {
    // Add to pending batch
    bw.pending = append(bw.pending, data)

    // Trigger flush if batch is full
    if len(bw.pending) >= bw.maxBatch {
        select {
        case bw.flushChan <- struct{}{}:
        default: // Don't block if flush already pending
        }
    }

    return nil
}

func (bw *BatchWriter) batchProcessor() {
    ticker := time.NewTicker(bw.timeout)
    defer ticker.Stop()

    for {
        select {
        case <-bw.flushChan:
            bw.flush()
        case <-ticker.C:
            if len(bw.pending) > 0 {
                bw.flush()
            }
        case <-bw.stopChan:
            bw.flush()
            return
        }
    }
}

func (bw *BatchWriter) flush() {
    if len(bw.pending) == 0 {
        return
    }

    // Combine all pending writes into single buffer
    bw.buffer = bw.buffer[:0]
    for _, data := range bw.pending {
        bw.buffer = append(bw.buffer, data...)
    }

    // Single syscall for all writes
    bw.file.Write(bw.buffer)

    // Clear pending
    bw.pending = bw.pending[:0]
}

// Directory traversal optimization - avoid multiple stat calls
func processDirectoryEfficient(dirPath string) error {
    // Single ReadDir call instead of multiple individual stats
    entries, err := os.ReadDir(dirPath)
    if err != nil {
        return err
    }

    // Batch process entries
    batch := make([]os.DirEntry, 0, 100)
    for _, entry := range entries {
        batch = append(batch, entry)

        if len(batch) >= 100 {
            processBatch(batch)
            batch = batch[:0]
        }
    }

    // Process remaining entries
    if len(batch) > 0 {
        processBatch(batch)
    }

    return nil
}

func processBatch(batch []os.DirEntry) {
    for _, entry := range batch {
        // Process each directory entry
        if !entry.IsDir() {
            info, err := entry.Info()
            if err == nil {
                processFileInfo(info)
            }
        }
    }
}
```

---

[0;34m[INFO][0m Processing section: 07-memory-optimization.md
## Memory Optimization

### Space-Time Trade-offs (go-perfbook Pattern)

- **Small memory for significant speed**: Lookup tables, caches, precomputed values
- **Linear trade-offs**: 2x memory usage for 2x performance gain
- **Diminishing returns**: Huge memory for small speedup (avoid this)
- **Understand your position**: Where are you on the memory/performance curve?

```go
// Lookup table example - trading space for speed
var popCountTable = [256]uint8{
    0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4,
    1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
    1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
    2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
    1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
    2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
    2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
    3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
    1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
    2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
    2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
    3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
    2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
    3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
    3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
    4, 5, 5, 6, 5, 6, 6, 7, 5, 6, 6, 7, 6, 7, 7, 8,
}

func popCountFast(x uint64) int {
    return int(popCountTable[x&0xff] +
               popCountTable[(x>>8)&0xff] +
               popCountTable[(x>>16)&0xff] +
               popCountTable[(x>>24)&0xff] +
               popCountTable[(x>>32)&0xff] +
               popCountTable[(x>>40)&0xff] +
               popCountTable[(x>>48)&0xff] +
               popCountTable[(x>>56)&0xff])
}

// Single-item cache - minimal memory for common case speedup
type TimeParser struct {
    lastFormat string
    lastLayout string
}

func (tp *TimeParser) Parse(value, format string) (time.Time, error) {
    if format == tp.lastFormat {
        return time.Parse(tp.lastLayout, value)
    }

    // Expensive format compilation
    layout := compileFormat(format)
    tp.lastFormat = format
    tp.lastLayout = layout

    return time.Parse(layout, value)
}

func compileFormat(format string) string {
    // Simplified format compilation (placeholder)
    return format // In real implementation, would convert format to Go time layout
}

// String interning - trade memory for deduplication
type StringInterner struct {
    strings map[string]string
    stats   atomic.Int64
}

func (si *StringInterner) Intern(s string) string {
    if interned, exists := si.strings[s]; exists {
        si.stats.Add(1) // Deduplication hit
        return interned
    }

    si.strings[s] = s
    return s
}

// Space-efficient bloom filter for negative lookups
type BloomFilter struct {
    bits   []uint64
    hashes int
    size   uint64
}

func (bf *BloomFilter) MightContain(data []byte) bool {
    // Small memory usage for fast negative answers
    hash1, hash2 := hash(data)
    for i := 0; i < bf.hashes; i++ {
        bit := (hash1 + uint64(i)*hash2) % bf.size
        if bf.bits[bit/64]&(1<<(bit%64)) == 0 {
            return false // Definitely not present
        }
    }
    return true // Might be present
}

func hash(data []byte) (uint64, uint64) {
    // Simplified hash function (placeholder)
    h1 := uint64(0)
    h2 := uint64(1)
    for _, b := range data {
        h1 = h1*31 + uint64(b)
        h2 = h2*37 + uint64(b)
    }
    return h1, h2
}
```

### Memory Layout Optimization

- **Struct field ordering**: Place larger fields first, group related fields
- **Cache-line awareness**: Pad atomic fields to prevent false sharing
- **Memory alignment**: Use proper alignment for better CPU cache performance
- **Pool reuse**: Extensive use of sync.Pool for object reuse
- **Pointer reduction**: Minimize pointers to reduce GC scan time

```go
// Good: Optimized field ordering (24 bytes total)
type OptimizedStruct struct {
    name    string  // 16 bytes (pointer + length)
    counter int64   // 8 bytes
    flag    bool    // 1 byte
    active  bool    // 1 byte + 6 padding
}

// Bad: Random field ordering (32 bytes total)
type BadStruct struct {
    flag    bool    // 1 byte + 7 padding
    counter int64   // 8 bytes
    name    string  // 16 bytes
    active  bool    // 1 byte + 7 padding
}

// Cache-line padding for atomic fields
type AtomicCounter struct {
    value atomic.Uint64
    _     [7]uint64 // Padding to prevent false sharing (64 bytes = cache line)
}

// Pointer-free slices for better GC performance
type PointFreeData struct {
    // Pack multiple values into single slice to reduce pointers
    values []uint64 // ID, timestamp, flags all packed
}

func (d *PointFreeData) GetID(idx int) uint32 {
    return uint32(d.values[idx] >> 32)
}

func (d *PointFreeData) GetTimestamp(idx int) uint32 {
    return uint32(d.values[idx])
}
```

### Algorithmic Memory Patterns (go-perfbook)

- **Big-O space complexity**: Choose data structures by space/time complexity
- **Constant factor optimization**: Small changes with big memory impact
- **Input-size dependent**: Different algorithms for different data sizes
- **Hybrid data structures**: Combine algorithms for optimal performance

```go
// Size-dependent algorithm choice
func searchData(data []Item, target string) int {
    if len(data) < 100 {
        // Linear search for small datasets (better cache locality)
        for i, item := range data {
            if item.Name == target {
                return i
            }
        }
        return -1
    }

    // Binary search for larger datasets (sorted data required)
    sort.Slice(data, func(i, j int) bool {
        return data[i].Name < data[j].Name
    })

    return sort.Search(len(data), func(i int) bool {
        return data[i].Name >= target
    })
}

// Hybrid data structure - bucketed approach
type HybridMap struct {
    buckets [][]KeyValue
    size    int
}

func (h *HybridMap) Get(key string) (string, bool) {
    bucket := h.buckets[hash(key)%len(h.buckets)]

    // Linear search within bucket (small buckets = cache friendly)
    for _, kv := range bucket {
        if kv.Key == key {
            return kv.Value, true
        }
    }
    return "", false
}

// Space-time trade-off: specialized for string keys
type StringInterner struct {
    strings map[string]string
    stats   atomic.Int64
}

func (si *StringInterner) Intern(s string) string {
    if interned, exists := si.strings[s]; exists {
        si.stats.Add(1) // Deduplication hit
        return interned
    }

    si.strings[s] = s
    return s
}
```

### Zero-Allocation Patterns

- **Pre-allocation with exact capacity**: Always specify slice/map capacity
- **String building optimization**: Use strings.Builder with pre-growth
- **Byte slice reuse**: Implement sync.Pool for buffer reuse
- **Avoid interface{} boxing**: Use generics instead

```go
// Pre-allocate with exact capacity
func processItems(source []Item) []ProcessedItem {
    // Good: Pre-allocate with known capacity
    results := make([]ProcessedItem, 0, len(source))

    for _, item := range source {
        processed := processItem(item)
        results = append(results, processed)
    }

    return results
}

// String building with pre-allocation
func buildString(parts []string) string {
    var totalLen int
    for _, part := range parts {
        totalLen += len(part)
    }

    var builder strings.Builder
    builder.Grow(totalLen) // ALWAYS pre-grow

    for _, part := range parts {
        builder.WriteString(part)
    }

    return builder.String()
}

// Byte slice reuse with sync.Pool
var bufferPool = sync.Pool{
    New: func() any {
        return make([]byte, 0, 4096) // 4KB initial capacity
    },
}

func processWithBuffer() []byte {
    buf := bufferPool.Get().([]byte)
    defer func() {
        buf = buf[:0] // Reset length, keep capacity
        bufferPool.Put(buf)
    }()

    // Use buf for processing
    buf = append(buf, someData...)

    // Return copy to avoid pool corruption
    result := make([]byte, len(buf))
    copy(result, buf)
    return result
}
```

### Memory Pool Management

- **Object pools**: Reuse expensive-to-create objects
- **Worker pools**: Prevent unlimited goroutine creation
- **Connection pools**: Reuse network connections and database handles
- **Multi-level pooling**: Different pools for different object sizes
- **Pool monitoring**: Track pool effectiveness and adjust sizes

```go
// Multi-level pooling by size
type MultiLevelPool struct {
    small  sync.Pool // < 1KB
    medium sync.Pool // 1KB - 64KB
    large  sync.Pool // > 64KB

    stats struct {
        smallHits  atomic.Uint64
        mediumHits atomic.Uint64
        largeHits  atomic.Uint64
        misses     atomic.Uint64
    }
}

func NewMultiLevelPool() *MultiLevelPool {
    return &MultiLevelPool{
        small: sync.Pool{
            New: func() any { return make([]byte, 0, 1024) },
        },
        medium: sync.Pool{
            New: func() any { return make([]byte, 0, 64*1024) },
        },
        large: sync.Pool{
            New: func() any { return make([]byte, 0, 1024*1024) },
        },
    }
}

func (p *MultiLevelPool) GetBuffer(size int) []byte {
    switch {
    case size <= 1024:
        p.stats.smallHits.Add(1)
        buf := p.small.Get().([]byte)
        return buf[:0] // Reset length
    case size <= 64*1024:
        p.stats.mediumHits.Add(1)
        buf := p.medium.Get().([]byte)
        return buf[:0]
    default:
        p.stats.largeHits.Add(1)
        buf := p.large.Get().([]byte)
        return buf[:0]
    }
}

func (p *MultiLevelPool) PutBuffer(buf []byte, size int) {
    switch {
    case size <= 1024:
        if cap(buf) >= 1024 {
            p.small.Put(buf)
        }
    case size <= 64*1024:
        if cap(buf) >= 64*1024 {
            p.medium.Put(buf)
        }
    default:
        if cap(buf) >= 1024*1024 {
            p.large.Put(buf)
        }
    }
}

// Object pool for expensive structs with monitoring
var processorPool = sync.Pool{
    New: func() any {
        return &ExpensiveProcessor{
            buffer:    make([]byte, 1024*1024), // 1MB buffer
            workspace: make(map[string]interface{}, 1000),
            // Other expensive fields
        }
    },
}

func processData(data []byte) error {
    processor := processorPool.Get().(*ExpensiveProcessor)
    defer func() {
        processor.Reset() // Clear sensitive data
        processorPool.Put(processor)
    }()

    return processor.Process(data)
}

// Worker pool to control memory usage
type WorkerPool struct {
    workers   int
    workChan  chan Work
    wg        sync.WaitGroup
    semaphore chan struct{} // Limit concurrent workers
}

func NewWorkerPool(maxWorkers int) *WorkerPool {
    return &WorkerPool{
        workers:   maxWorkers,
        workChan:  make(chan Work, maxWorkers*2),
        semaphore: make(chan struct{}, maxWorkers),
    }
}

func (wp *WorkerPool) Submit(work Work) {
    wp.semaphore <- struct{}{} // Acquire
    wp.wg.Add(1)

    go func() {
        defer func() {
            <-wp.semaphore // Release
            wp.wg.Done()
        }()

        work.Execute()
    }()
}
```

### Real-Time Memory Allocation Tracking

- **Live allocation monitoring**: Track allocations as they happen
- **Memory pressure detection**: Automatically detect and respond to high memory usage
- **Allocation profiling**: Identify allocation hotspots in real-time
- **Dynamic GC tuning**: Adjust garbage collection based on allocation patterns

```go
// Real-time allocation tracker
type AllocationTracker struct {
    allocations   atomic.Uint64
    deallocations atomic.Uint64
    currentBytes  atomic.Uint64
    peakBytes     atomic.Uint64

    // Allocation pattern tracking
    smallAllocs  atomic.Uint64 // < 1KB
    mediumAllocs atomic.Uint64 // 1KB - 1MB
    largeAllocs  atomic.Uint64 // > 1MB

    // Pressure thresholds
    warningThreshold uint64
    criticalThreshold uint64

    // Callbacks for pressure events
    onWarning  func(stats AllocationStats)
    onCritical func(stats AllocationStats)
}

type AllocationStats struct {
    Current     uint64
    Peak        uint64
    Allocations uint64
    SmallRatio  float64
    MediumRatio float64
    LargeRatio  float64
}

func NewAllocationTracker(warningMB, criticalMB uint64) *AllocationTracker {
    return &AllocationTracker{
        warningThreshold:  warningMB * 1024 * 1024,
        criticalThreshold: criticalMB * 1024 * 1024,
    }
}

func (at *AllocationTracker) TrackAllocation(size uint64) {
    // Update counters
    at.allocations.Add(1)
    newCurrent := at.currentBytes.Add(size)

    // Update peak if necessary
    for {
        currentPeak := at.peakBytes.Load()
        if newCurrent <= currentPeak {
            break
        }
        if at.peakBytes.CompareAndSwap(currentPeak, newCurrent) {
            break
        }
    }

    // Track allocation size pattern
    switch {
    case size < 1024:
        at.smallAllocs.Add(1)
    case size < 1024*1024:
        at.mediumAllocs.Add(1)
    default:
        at.largeAllocs.Add(1)
    }

    // Check pressure thresholds
    at.checkPressure(newCurrent)
}

func (at *AllocationTracker) TrackDeallocation(size uint64) {
    at.deallocations.Add(1)
    at.currentBytes.Add(^uint64(size - 1)) // Atomic subtract
}

func (at *AllocationTracker) checkPressure(current uint64) {
    if current >= at.criticalThreshold && at.onCritical != nil {
        at.onCritical(at.GetStats())
    } else if current >= at.warningThreshold && at.onWarning != nil {
        at.onWarning(at.GetStats())
    }
}

func (at *AllocationTracker) GetStats() AllocationStats {
    totalAllocs := at.smallAllocs.Load() + at.mediumAllocs.Load() + at.largeAllocs.Load()
    if totalAllocs == 0 {
        totalAllocs = 1 // Avoid division by zero
    }

    return AllocationStats{
        Current:     at.currentBytes.Load(),
        Peak:        at.peakBytes.Load(),
        Allocations: at.allocations.Load(),
        SmallRatio:  float64(at.smallAllocs.Load()) / float64(totalAllocs),
        MediumRatio: float64(at.mediumAllocs.Load()) / float64(totalAllocs),
        LargeRatio:  float64(at.largeAllocs.Load()) / float64(totalAllocs),
    }
}

// Tracked buffer pool with automatic monitoring
type TrackedBufferPool struct {
    pool    sync.Pool
    tracker *AllocationTracker
    size    int
}

func NewTrackedBufferPool(size int, tracker *AllocationTracker) *TrackedBufferPool {
    return &TrackedBufferPool{
        pool: sync.Pool{
            New: func() any {
                buf := make([]byte, 0, size)
                tracker.TrackAllocation(uint64(size))
                return buf
            },
        },
        tracker: tracker,
        size:    size,
    }
}

func (tbp *TrackedBufferPool) Get() []byte {
    return tbp.pool.Get().([]byte)[:0]
}

func (tbp *TrackedBufferPool) Put(buf []byte) {
    if cap(buf) == tbp.size {
        tbp.pool.Put(buf)
    } else {
        // Buffer was resized, track deallocation of old size and allocation of new
        tbp.tracker.TrackDeallocation(uint64(tbp.size))
        tbp.tracker.TrackAllocation(uint64(cap(buf)))
    }
}
```

### Advanced Context Management

- **Dynamic timeouts**: Adjust timeouts based on system load
- **Graceful degradation**: Handle resource exhaustion gracefully
- **Operation counting**: Track active operations for clean shutdown
- **Resource pooling**: Share expensive resources across contexts

````go
// Advanced context manager with resource tracking
type ContextManager struct {
    activeOperations atomic.Int64
    maxOperations    int64

    // Dynamic timeout calculation
    baseTimeout     time.Duration
    loadFactor      atomic.Uint64 // 0-100, represents system load percentage
    timeoutMultiplier float64

    // Resource pools
    resourcePools map[string]*ResourcePool
    poolMutex     sync.RWMutex

    // Shutdown coordination
    shutdownChan chan struct{}
    shutdownOnce sync.Once
}

type ResourcePool struct {
    resources chan interface{}
    factory   func() interface{}
    cleanup   func(interface{})

    created atomic.Int64
    inUse   atomic.Int64
    maxSize int
}

func NewContextManager(maxOps int64, baseTimeout time.Duration) *ContextManager {
    return &ContextManager{
        maxOperations:     maxOps,
        baseTimeout:       baseTimeout,
        timeoutMultiplier: 1.5, // 50% additional time under load
        resourcePools:     make(map[string]*ResourcePool),
        shutdownChan:      make(chan struct{}),
    }
}

func (cm *ContextManager) WithOperation(ctx context.Context) (context.Context, func(), error) {
    // Check if we can accept new operations
    current := cm.activeOperations.Add(1)
    if current > cm.maxOperations {
        cm.activeOperations.Add(-1)
        return nil, nil, fmt.Errorf("maximum operations exceeded: %d", cm.maxOperations)
    }

    // Calculate dynamic timeout based on load
    load := float64(cm.loadFactor.Load()) / 100.0
    timeout := time.Duration(float64(cm.baseTimeout) * (1.0 + load*cm.timeoutMultiplier))

    // Create context with dynamic timeout
    ctx, cancel := context.WithTimeout(ctx, timeout)

    // Check for shutdown
    select {
    case <-cm.shutdownChan:
        cancel()
        cm.activeOperations.Add(-1)
        return nil, nil, fmt.Errorf("system is shutting down")
    default:
    }

    // Return cleanup function
    cleanup := func() {
        cancel()
        cm.activeOperations.Add(-1)
    }

    return ctx, cleanup, nil
}

func (cm *ContextManager) GetResource(ctx context.Context, poolName string) (interface{}, func(), error) {
    cm.poolMutex.RLock()
    pool, exists := cm.resourcePools[poolName]
    cm.poolMutex.RUnlock()

    if !exists {
        return nil, nil, fmt.Errorf("resource pool %s not found", poolName)
    }

    select {
    case resource := <-pool.resources:
        pool.inUse.Add(1)

        release := func() {
            pool.inUse.Add(-1)
            if pool.cleanup != nil {
                pool.cleanup(resource)
            }

            select {
            case pool.resources <- resource:
                // Resource returned to pool
            default:
                // Pool is full, discard resource
            }
        }

        return resource, release, nil

    case <-ctx.Done():
        return nil, nil, ctx.Err()

    default:
        // Pool is empty, try to create new resource
        if pool.created.Load() < int64(pool.maxSize) {
            pool.created.Add(1)
            pool.inUse.Add(1)

            resource := pool.factory()

            release := func() {
                pool.inUse.Add(-1)
                if pool.cleanup != nil {
                    pool.cleanup(resource)
                }

                select {
                case pool.resources <- resource:
                    // Resource returned to pool
                default:
                    // Pool is full, discard resource
                    pool.created.Add(-1)
                }
            }

            return resource, release, nil
        }

        return nil, nil, fmt.Errorf("resource pool %s exhausted", poolName)
    }
}

func (cm *ContextManager) RegisterResourcePool(name string, maxSize int, factory func() interface{}, cleanup func(interface{})) {
    cm.poolMutex.Lock()
    defer cm.poolMutex.Unlock()

    cm.resourcePools[name] = &ResourcePool{
        resources: make(chan interface{}, maxSize),
        factory:   factory,
        cleanup:   cleanup,
        maxSize:   maxSize,
    }
}

func (cm *ContextManager) UpdateLoad(loadPercent int) {
    if loadPercent < 0 {
        loadPercent = 0
    } else if loadPercent > 100 {
        loadPercent = 100
    }
    cm.loadFactor.Store(uint64(loadPercent))
}

func (cm *ContextManager) Shutdown(timeout time.Duration) error {
    cm.shutdownOnce.Do(func() {
        close(cm.shutdownChan)
    })

    // Wait for active operations to complete
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if cm.activeOperations.Load() == 0 {
            return nil
        }
        time.Sleep(10 * time.Millisecond)
    }

    remaining := cm.activeOperations.Load()
    if remaining > 0 {
        return fmt.Errorf("shutdown timeout: %d operations still active", remaining)
    }

    return nil
}

### Memory Monitoring and Profiling

- **Runtime memory stats**: Monitor heap usage in production
- **Memory profiling**: Use pprof for memory leak detection
- **GC optimization**: Tune GOGC for your workload

```go
// Memory monitoring
type MemoryMonitor struct {
    maxHeapSize   atomic.Uint64
    currentHeap   atomic.Uint64
    gcCount       atomic.Uint64

    alertThreshold uint64
    alertCallback  func(stats runtime.MemStats)
}

func (m *MemoryMonitor) Start(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            var stats runtime.MemStats
            runtime.ReadMemStats(&stats)

            m.currentHeap.Store(stats.HeapInuse)
            m.gcCount.Store(uint64(stats.NumGC))

            // Update max heap if current is higher
            for {
                currentMax := m.maxHeapSize.Load()
                if stats.HeapInuse <= currentMax {
                    break
                }
                if m.maxHeapSize.CompareAndSwap(currentMax, stats.HeapInuse) {
                    break
                }
            }

            // Alert if threshold exceeded
            if stats.HeapInuse > m.alertThreshold && m.alertCallback != nil {
                m.alertCallback(stats)
            }
        }
    }
}

// Force GC when memory pressure is high
func (m *MemoryMonitor) ForceGCIfNeeded() {
    var stats runtime.MemStats
    runtime.ReadMemStats(&stats)

    // Force GC if heap is above threshold and growing
    if stats.HeapInuse > m.alertThreshold &&
       stats.HeapInuse > stats.HeapReleased*2 {
        runtime.GC()
        runtime.GC() // Double GC to ensure cleanup
    }
}
````

### Memory-Efficient Data Structures

- **Compact representations**: Use byte arrays instead of structs when possible
- **String interning**: Deduplicate common strings
- **Bit packing**: Pack multiple boolean flags into single integers

```go
// Compact data representation
type CompactRecord struct {
    // Pack multiple fields into single byte array
    data [32]byte // Fixed size record
}

func (r *CompactRecord) GetID() uint64 {
    return binary.LittleEndian.Uint64(r.data[0:8])
}

func (r *CompactRecord) SetID(id uint64) {
    binary.LittleEndian.PutUint64(r.data[0:8], id)
}

func (r *CompactRecord) GetFlags() uint32 {
    return binary.LittleEndian.Uint32(r.data[8:12])
}

// String interning for deduplication
type StringInterner struct {
    strings sync.Map
    stats   atomic.Int64
}

func (si *StringInterner) Intern(s string) string {
    if interned, ok := si.strings.Load(s); ok {
        si.stats.Add(1) // Deduplication hit
        return interned.(string)
    }

    // Store and return the original string
    si.strings.Store(s, s)
    return s
}

// Bit packing for flags
type PackedFlags uint64

const (
    FlagEnabled PackedFlags = 1 << iota
    FlagVisible
    FlagReadonly
    FlagArchived
    FlagEncrypted
    // ... up to 64 flags
)

func (f PackedFlags) Has(flag PackedFlags) bool {
    return f&flag != 0
}

func (f *PackedFlags) Set(flag PackedFlags) {
    *f |= flag
}

func (f *PackedFlags) Clear(flag PackedFlags) {
    *f &^= flag
}
```

---

[0;34m[INFO][0m Processing section: 08-code-quality.md
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

---

[0;34m[INFO][0m Processing section: 09-test-coverage.md
## 100% Test Coverage with Timeouts

### Mandatory Test Structure

```go
// FunctionName is the function we're testing
func FunctionName(ctx context.Context, input string) (string, error) {
    if input == "" {
        return "", errors.New("empty input")
    }
    return "processed_" + input, nil
}

func TestFunctionName(t *testing.T) {
    // ALWAYS set test timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
        timeout time.Duration // Per-test timeout
    }{
        {
            name:    "valid_input",
            input:   "test",
            want:    "processed_test",
            wantErr: false,
            timeout: time.Second,
        },
        {
            name:    "empty_input",
            input:   "",
            want:    "",
            wantErr: true,
            timeout: 100 * time.Millisecond,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Helper()

            // Per-test timeout
            testCtx, testCancel := context.WithTimeout(ctx, tt.timeout)
            defer testCancel()

            // Use require for fatal assertions
            require.NotNil(t, testCtx)

            got, err := FunctionName(testCtx, tt.input)

            if tt.wantErr {
                require.Error(t, err)
                assert.Empty(t, got)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Concurrent Test Safety

```go
// Service represents our service for testing
type Service struct {
    storage Storage
}

func NewService(storage Storage) *Service {
    return &Service{storage: storage}
}

func (s *Service) Process(ctx context.Context, data string) error {
    if data == "" {
        return errors.New("empty data")
    }
    // Simulate processing
    return nil
}

func (s *Service) ProcessData(ctx context.Context, key string, data []byte) error {
    return s.storage.Save(ctx, key, data)
}

func TestConcurrentAccess(t *testing.T) {
    t.Parallel()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    const numGoroutines = 100
    var wg sync.WaitGroup
    errors := make(chan error, numGoroutines)

    service := NewService(nil) // Using nil storage for this test

    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            select {
            case <-ctx.Done():
                errors <- ctx.Err()
                return
            default:
            }

            if err := service.Process(ctx, fmt.Sprintf("data-%d", id)); err != nil {
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
            t.Errorf("goroutine error: %v", err)
        }
    case <-ctx.Done():
        t.Fatal("test timeout exceeded")
    }
}
}
```

### Mock Generation (100% Coverage)

```go
//go:generate mockgen -source=interfaces.go -destination=mocks/mock_interfaces.go

type Storage interface {
    Save(ctx context.Context, key string, data []byte) error
    Load(ctx context.Context, key string) ([]byte, error)
    Delete(ctx context.Context, key string) error
}

// Test with mocks
func TestServiceWithMock(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockStorage := mocks.NewMockStorage(ctrl)
    service := NewService(mockStorage)

    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    // Setup expectations
    mockStorage.EXPECT().
        Save(gomock.Any(), "test-key", gomock.Any()).
        Return(nil).
        Times(1)

    err := service.ProcessData(ctx, "test-key", []byte("test-data"))
    require.NoError(t, err)
}
```

---

[0;34m[INFO][0m Processing section: 10-review-checklist.md
## Copilot Review Checklist

### Optimization Framework Review (go-perfbook)

- [ ] **Three Questions Applied**: (1) Eliminate work (2) Best algorithm (3) Best implementation
- [ ] **Input size consideration**: Algorithm choice matches realistic data sizes
- [ ] **Space-time trade-offs**: Understand position on memory/performance curve
- [ ] **Constant factors**: Branch prediction, cache locality, branchless patterns
- [ ] **Specialization justified**: Custom implementations only when demonstrably better

### Performance Review

- [ ] All counters use atomic operations
- [ ] Zero-allocation patterns implemented
- [ ] Memory pre-allocation where possible
- [ ] Efficient string building
- [ ] Optimal struct field ordering
- [ ] Cache-friendly data access patterns
- [ ] Batching strategies for I/O operations
- [ ] Size-aware algorithm selection

### Benchmarking Review (Only When Requested)

- [ ] **Benchmark creation**: Only when user explicitly requests performance measurement
- [ ] **Benchmark quality**: Proper setup, teardown, and realistic data when created
- [ ] **Performance validation**: Memory and CPU measurements when benchmarking is requested

### Advanced Optimization Patterns

- [ ] **Polyalgorithm implementation**: Adaptive algorithm selection based on input
- [ ] **Cache hierarchy**: Single-item cache, LRU, bloom filters where appropriate
- [ ] **Memory layout**: SoA vs AoS choice justified, cache-line padding implemented
- [ ] **Branch optimization**: Likelihood-ordered conditionals, branchless code
- [ ] **Vectorization-friendly**: Data structures enable compiler auto-vectorization

### Production Scale Patterns (2M+ Users Proven)

- [ ] **HTTP client pooling**: Global HTTP client with connection pooling configured
- [ ] **Server timeouts**: ReadTimeout, WriteTimeout, IdleTimeout properly set
- [ ] **Database batching**: Batch operations instead of individual inserts/updates
- [ ] **Worker pool limiting**: Strict limits with circuit breakers and error monitoring
- [ ] **JSON optimization**: Zero-allocation encoding with pooled buffers
- [ ] **Buffer pools**: Multi-tier buffer pools for different data sizes
- [ ] **Structured responses**: Consistent API response format with metadata

### Concurrency Anti-Pattern Prevention

- [ ] **No fire-and-forget goroutines**: All `go func(){}()` have supervision
- [ ] **Structured concurrency**: Context, error handling, panic recovery mandatory
- [ ] **Worker pool patterns**: Bounded concurrency instead of unlimited goroutines
- [ ] **Lifecycle management**: Proper startup, shutdown, and resource cleanup
- [ ] **Observability**: Metrics for active jobs, total processed, error rates
- [ ] **Context propagation**: All long-running operations accept context.Context

### Memory Optimization (Pointer vs Value Awareness)

- [ ] **Conscious pointer decisions**: Pointers only for large structs or optional fields
- [ ] **Value types for small data**: bool, int, time.Time as values not pointers
- [ ] **GC-friendly patterns**: Fewer pointers, stack allocation preferred
- [ ] **Benchmark validation**: Memory decisions backed by actual benchmarks
- [ ] **Optional field semantics**: Pointers used only when nil is meaningful
- [ ] **Connection pool tuning**: MaxOpenConns, MaxIdleConns, ConnMaxLifetime configured
- [ ] **Worker pool capping**: Goroutine limits with semaphores or buffered channels
- [ ] **Structured JSON**: Avoid map[string]interface{}, use typed structs with pools
- [ ] **Buffer pooling**: Multi-size buffer pools for different use cases
- [ ] **Rate limiting**: Token bucket or similar for API endpoints
- [ ] **Graceful shutdown**: Proper cleanup of resources and connections
- [ ] **Health checks**: Database and service health validation

### Security Review

- [ ] All inputs validated
- [ ] No math/rand for security purposes
- [ ] Proper error handling without information leakage
- [ ] Command injection prevention
- [ ] Secure memory handling for sensitive data

### Testing Review

- [ ] 100% test coverage achieved
- [ ] All tests have timeouts
- [ ] Concurrent tests for race conditions
- [ ] Mock interfaces for external dependencies
- [ ] Table-driven test patterns
- [ ] Performance regression tests

### Code Quality Review

- [ ] Proper Godoc format with code blocks
- [ ] Channel safety patterns
- [ ] Context usage for cancellation
- [ ] Error wrapping with context
- [ ] Interface-based design for mockability

---

[0;34m[INFO][0m Processing section: 11-summary.md
## Summary

Copilot must operate as an ultra-performance Go 1.24+ expert with **intelligent optimization framework** based on go-perfbook principles and production-proven patterns:

### Core Decision Framework

- **The Three Questions**: (1) Do we have to do this? (2) Best algorithm? (3) Best implementation?
- **Profile-driven decisions**: Always measure before and after optimizations
- **Input-size awareness**: Choose algorithms based on realistic data characteristics
- **Space-time trade-offs**: Understand position on memory/performance curve
- **Anti-pattern prevention**: Proactively avoid dangerous patterns like fire-and-forget goroutines

### Performance Excellence Standards

- **Zero-compromise performance**: Every line optimized for CPU/RAM efficiency using production-proven patterns
- **Go 1.24+ features**: Leverage yield iterators, enhanced atomics, min/max built-ins, weak pointers, arena allocation
- **Algorithmic intelligence**: Size-aware algorithm selection, polyalgorithm patterns, constant factor optimization
- **Production scalability**: HTTP client pooling, database batching, worker pool limiting, structured JSON optimization
- **Memory mastery**: Conscious pointer vs value decisions, GC-friendly patterns, multi-tier buffer pools

### Concurrency & Reliability

- **Structured concurrency**: No fire-and-forget goroutines, supervised workers with lifecycle management
- **Production reliability**: Context propagation, panic recovery, error handling, observability metrics
- **Resource boundaries**: Strict worker limits, circuit breakers, clean shutdown patterns
- **Latency optimization**: Predictive caching, branch prediction, sub-second response targets

### Security & Quality

- **Proactive security**: Anticipate and prevent vulnerabilities before they occur
- **100% test coverage**: Comprehensive, timeout-protected, mockable tests
- **Holistic integration**: Consider entire project impact for optimal solutions
- **Production-grade reliability**: Channel safety, atomic operations, proper error handling

### Advanced Optimization Patterns

- **Specialization vs generalization**: Context-aware optimization for specific use cases
- **Cache hierarchy optimization**: Multi-level caches, bloom filters, single-item caches
- **Memory layout optimization**: SoA vs AoS, cache-line awareness, false sharing prevention
- **Adaptive algorithms**: Runtime metrics-driven algorithm selection

Every suggestion must be measurably better in performance, security, and maintainability, following the **intelligent optimization framework** that prevents premature optimization while ensuring we address bottlenecks at the correct abstraction level.

---

[0;34m[INFO][0m Processing section: 12-go-design-patterns.md
## Go Design Patterns Documentation Standards

### The 5 Essential Patterns for Real Go Projects

Based on production-tested patterns, here are the design patterns that matter most in Go development, documented according to our Godoc standards.

---

### 1. Strategy Pattern - Clean Conditional Logic

Replace 50-line switch-case statements with clean interfaces.

```go
// ShippingStrategy Interface for calculating shipping costs by region
type ShippingStrategy interface {
    CalculateCost(weight float64) float64 // Calculate cost based on weight
}

// IndiaShipping Shipping cost calculation for India
type IndiaShipping struct{}

// CalculateCost Calculate shipping cost for India
// Code block:
//
//  shipping := &IndiaShipping{}
//  cost := shipping.CalculateCost(2.5)
//  fmt.Printf("Cost: %.2f\n", cost)
//
// Parameters:
//   - 1 weight: float64 - package weight in kg (must be positive)
//
// Returns:
//   - 1 cost: float64 - shipping cost in local currency
func (s IndiaShipping) CalculateCost(weight float64) float64 {
    return 50 + weight*10
}

// ShippingCalculator Calculator using appropriate strategy
type ShippingCalculator struct {
    strategy ShippingStrategy // Active calculation strategy
}

// NewShippingCalculator Create calculator with given strategy
// Code block:
//
//  calculator := NewShippingCalculator(&IndiaShipping{})
//  cost := calculator.Calculate(2.5)
//  fmt.Printf("Total: %.2f\n", cost)
//
// Parameters:
//   - 1 strategy: ShippingStrategy - calculation strategy (cannot be nil)
//
// Returns:
//   - 1 calculator: *ShippingCalculator - configured calculator
func NewShippingCalculator(strategy ShippingStrategy) *ShippingCalculator {
    return &ShippingCalculator{strategy: strategy}
}
```

---

### 2. Factory Pattern - Flexible Object Creation

Create different object types based on context.

```go
// DatabaseType Supported database type
type DatabaseType string

const (
    MySQL    DatabaseType = "mysql"    // MySQL database
    Postgres DatabaseType = "postgres" // PostgreSQL database
)

// Database Interface for database operations
type Database interface {
    Connect() error    // Establish connection
    Query(sql string)  // Execute query
    Close() error      // Close connection
}

// NewDatabase Factory to create database instance
// Code block:
//
//  db, err := NewDatabase(MySQL, "localhost:3306")
//  if err != nil {
//      log.Fatal(err)
//  }
//  defer db.Close()
//
// Parameters:
//   - 1 dbType: DatabaseType - database type to create
//   - 2 config: string - connection configuration (cannot be empty)
//
// Returns:
//   - 1 db: Database - configured database instance
//   - 2 error - nil if creation successful, error if unsupported type
func NewDatabase(dbType DatabaseType, config string) (Database, error) {
    switch dbType {
    case MySQL:
        return &MySQLDB{host: config, port: 3306}, nil
    case Postgres:
        return &PostgresDB{connectionString: config}, nil
    default:
        return nil, fmt.Errorf("unsupported database type: %s", dbType)
    }
}
```

---

### 3. Builder Pattern - Complex Object Construction

Build objects with many optional parameters.

```go
// ServerConfig Complete server configuration
type ServerConfig struct {
    host         string        // Listen address
    port         int           // Listen port
    timeout      time.Duration // Request timeout
    enableHTTPS  bool          // HTTPS activation
}

// ServerBuilder Builder for constructing server configuration
type ServerBuilder struct {
    config *ServerConfig // Configuration being built
}

// NewServerBuilder Create new builder with default values
// Code block:
//
//  builder := NewServerBuilder()
//  config := builder.Host("localhost").Port(8080).Build()
//  fmt.Printf("Server: %s:%d\n", config.host, config.port)
//
// Returns:
//   - 1 builder: *ServerBuilder - builder with default configuration
func NewServerBuilder() *ServerBuilder {
    return &ServerBuilder{
        config: &ServerConfig{
            host:    "localhost",
            port:    8080,
            timeout: 30 * time.Second,
        },
    }
}

// Host Configure listen address
// Code block:
//
//  config := NewServerBuilder().Host("0.0.0.0").Port(9000).Build()
//
// Parameters:
//   - 1 host: string - listen address (cannot be empty)
//
// Returns:
//   - 1 builder: *ServerBuilder - builder for chaining
func (b *ServerBuilder) Host(host string) *ServerBuilder {
    b.config.host = host
    return b
}

// Build Construct final configuration
// Code block:
//
//  config := NewServerBuilder().
//      Host("api.example.com").
//      Port(443).
//      Build()
//
// Returns:
//   - 1 config: *ServerConfig - complete server configuration
func (b *ServerBuilder) Build() *ServerConfig {
    return b.config
}
```

---

### 4. Observer Pattern - Event Communication

Notify multiple components on state changes.

```go
// Event Represents system event
type Event struct {
    Type      string      // Event type
    Data      interface{} // Associated data
    Timestamp time.Time   // Event moment
}

// Observer Interface for event observers
type Observer interface {
    HandleEvent(event Event) // Process received event
    GetID() string           // Return unique identifier
}

// EventManager Central event manager
type EventManager struct {
    observers map[string]Observer // Map of registered observers
    mu        sync.RWMutex        // Mutex for concurrent access
}

// NewEventManager Create new event manager
// Code block:
//
//  manager := NewEventManager()
//  observer := NewEmailNotifier("test@example.com")
//  manager.Subscribe(observer)
//
// Returns:
//   - 1 manager: *EventManager - initialized manager
func NewEventManager() *EventManager {
    return &EventManager{
        observers: make(map[string]Observer),
    }
}

// Subscribe Register observer
// Code block:
//
//  manager := NewEventManager()
//  notifier := NewEmailNotifier("admin@example.com")
//  manager.Subscribe(notifier)
//
// Parameters:
//   - 1 observer: Observer - observer to register (cannot be nil)
func (em *EventManager) Subscribe(observer Observer) {
    em.mu.Lock()
    defer em.mu.Unlock()
    em.observers[observer.GetID()] = observer
}

// Publish Publish event to all observers
// Code block:
//
//  event := Event{Type: "order.completed", Data: "Order #123"}
//  manager.Publish(event)
//
// Parameters:
//   - 1 event: Event - event to publish
func (em *EventManager) Publish(event Event) {
    em.mu.RLock()
    defer em.mu.RUnlock()

    for _, observer := range em.observers {
        go observer.HandleEvent(event) // Asynchronous notification
    }
}
```

---

### 5. Dependency Injection - Decoupling Components

Decouple components and facilitate testing.

```go
// Logger Interface for logging
type Logger interface {
    Info(message string)  // Log info level
    Error(message string) // Log error level
}

// Repository Interface for persistence
type Repository interface {
    Save(data interface{}) error // Save data
    Find(id string) interface{}  // Find by ID
}

// UserService Business service with injected dependencies
type UserService struct {
    logger Logger     // Injected logger
    repo   Repository // Injected repository
}

// NewUserService Create service with dependency injection
// Code block:
//
//  logger := &ConsoleLogger{}
//  repo := &DatabaseRepository{}
//  service := NewUserService(logger, repo)
//
//  err := service.CreateUser("John Doe")
//  if err != nil {
//      log.Fatal(err)
//  }
//
// Parameters:
//   - 1 logger: Logger - logger for traces (cannot be nil)
//   - 2 repo: Repository - repository for persistence (cannot be nil)
//
// Returns:
//   - 1 service: *UserService - service configured with dependencies
func NewUserService(logger Logger, repo Repository) *UserService {
    return &UserService{
        logger: logger,
        repo:   repo,
    }
}

// CreateUser Create new user
// Code block:
//
//  err := service.CreateUser("Jane Smith")
//  if err != nil {
//      log.Printf("Creation error: %v", err)
//  }
//
// Parameters:
//   - 1 name: string - user name (cannot be empty)
//
// Returns:
//   - 1 error - nil if creation successful, error otherwise
func (s *UserService) CreateUser(name string) error {
    s.logger.Info(fmt.Sprintf("Creating user: %s", name))

    user := map[string]string{"name": name}
    err := s.repo.Save(user)
    if err != nil {
        s.logger.Error(fmt.Sprintf("Save failed: %v", err))
        return err
    }

    s.logger.Info("User created successfully")
    return nil
}
```

---

### Pattern Usage in superviz.io

These patterns are particularly useful in superviz.io context:

- **Strategy**: Managing different package managers (apt, yum, apk, etc.)
- **Factory**: Creating SSH clients based on configuration
- **Builder**: Building complex installation configurations
- **Observer**: Installation progress notifications
- **DI**: Injecting infrastructure services into CLI commands

All examples follow our Godoc standards with code blocks, parameters, and returns documentation.

---

[0;34m[INFO][0m Processing section: 13-production-scale-patterns.md
# Production-Scale Patterns from Real-World Articles

## Anti-Patterns from Real Production Issues

### The Fire-and-Forget Goroutine Anti-Pattern

```go
// DANGEROUS: Unsupervised goroutines (from "Most Dangerous Line" article)
func SendWelcomeEmail(user User) {
    // ❌ BAD: Fire-and-forget goroutine
    go func() {
        err := smtp.Send(user.Email, "Welcome!")
        if err != nil {
            // Error silently disappears into the void
            // No observability, no recovery, no cleanup
        }
    }()
    // Function returns immediately, no idea if email sent
}

// ✅ GOOD: Supervised concurrency with context and error handling
func SendWelcomeEmailSafe(ctx context.Context, user User) error {
    errChan := make(chan error, 1)

    go func() {
        defer func() {
            if r := recover(); r != nil {
                errChan <- fmt.Errorf("panic in email sender: %v", r)
            }
        }()

        err := smtp.Send(user.Email, "Welcome!")
        errChan <- err
    }()

    select {
    case err := <-errChan:
        return err
    case <-ctx.Done():
        return ctx.Err()
    }
}

// ✅ BETTER: Structured concurrency with worker pool
type EmailService struct {
    workerPool *WorkerPool
    metrics    *EmailMetrics
}

func (s *EmailService) SendWelcomeEmail(ctx context.Context, user User) error {
    task := EmailTask{
        To:      user.Email,
        Template: "welcome",
        Data:     user,
    }

    result := s.workerPool.Submit(ctx, task)
    s.metrics.RecordSend(result.Error == nil)

    return result.Error
}
```

### Pointer vs Value Performance Awareness

```go
// From "Stop Using Pointers" article - conscious decisions about memory

// ❌ BAD: Pointer overuse without justification
type SmallConfig struct {
    Enabled *bool   `json:"enabled"`     // 8 bytes pointer vs 1 byte bool
    Count   *int    `json:"count"`       // 8 bytes pointer vs 8 bytes int
    Name    *string `json:"name"`        // 8 bytes pointer vs 16 bytes string
}

// ✅ GOOD: Value types for small, immutable data
type SmallConfig struct {
    Enabled bool   `json:"enabled"`     // Stack allocation, GC-friendly
    Count   int    `json:"count"`       // No pointer indirection
    Name    string `json:"name"`        // String is already a reference type
}

// ✅ WHEN TO USE POINTERS: Large structs, optional fields, or need mutation
type LargeUserProfile struct {
    ID           int64
    Avatar       *ImageData    // Large struct - use pointer
    Preferences  *UserPrefs    // Optional field - use pointer for nil semantics
    CreatedAt    time.Time     // Small struct - use value
    LastActive   time.Time     // Small struct - use value
}

// Benchmark-driven decision making
func BenchmarkPointerVsValue(b *testing.B) {
    configs := make([]SmallConfig, 1000)
    configPtrs := make([]*SmallConfig, 1000)

    b.Run("Value", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            for _, config := range configs {
                _ = config.Enabled // Direct access
            }
        }
    })

    b.Run("Pointer", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            for _, config := range configPtrs {
                _ = config.Enabled // Pointer indirection
            }
        }
    })
}
```

### Latency Reduction Techniques (80% Improvement Patterns)

```go
// From "We Slashed Latency by 80%" article

// Memory-first caching with predictive loading
type LatencyOptimizedCache struct {
    data        sync.Map
    predictions map[string]float64
    mutex       sync.RWMutex

    // Real-time metrics
    hits        atomic.Uint64
    misses      atomic.Uint64
    avgLatency  atomic.Uint64 // nanoseconds
}

func (c *LatencyOptimizedCache) Get(key string) (interface{}, bool) {
    start := time.Now()
    defer func() {
        latency := uint64(time.Since(start).Nanoseconds())
        c.updateAvgLatency(latency)
    }()

    if value, ok := c.data.Load(key); ok {
        c.hits.Add(1)

        // Predictive prefetching for related data
        go c.prefetchRelated(key)

        return value, true
    }

    c.misses.Add(1)
    return nil, false
}

func (c *LatencyOptimizedCache) prefetchRelated(key string) {
    c.mutex.RLock()
    score, exists := c.predictions[key]
    c.mutex.RUnlock()

    if exists && score > 0.7 { // High probability threshold
        relatedKeys := c.getRelatedKeys(key)
        for _, relKey := range relatedKeys {
            if _, loaded := c.data.Load(relKey); !loaded {
                // Async load in background
                go c.loadAndCache(relKey)
            }
        }
    }
}

// CPU-optimized hot path patterns
func (c *LatencyOptimizedCache) updateAvgLatency(newLatency uint64) {
    // Lock-free moving average
    for {
        current := c.avgLatency.Load()
        // Exponential moving average: new = old * 0.9 + current * 0.1
        updated := (current*9 + newLatency) / 10
        if c.avgLatency.CompareAndSwap(current, updated) {
            break
        }
    }
}

// Branch prediction optimization
func (c *LatencyOptimizedCache) processRequest(req *Request) *Response {
    // Order checks by likelihood (most common cases first)
    if req.Type == "read" { // 80% of requests
        return c.handleRead(req)
    }
    if req.Type == "write" { // 15% of requests
        return c.handleWrite(req)
    }
    if req.Type == "delete" { // 4% of requests
        return c.handleDelete(req)
    }
    // 1% of requests - complex operations
    return c.handleComplex(req)
}
```

### Production HTTP Client/Server (2M+ Users)

```go
// Production HTTP client configuration - handles 2M+ users
func NewProductionHTTPClient() *http.Client {
    transport := &http.Transport{
        MaxIdleConns:        100,              // Connection pool size
        MaxIdleConnsPerHost: 30,               // Per-host connection limit
        IdleConnTimeout:     90 * time.Second, // Keep-alive timeout

        // TCP-level optimizations
        DisableKeepAlives:   false,
        DisableCompression:  false,

        // TLS optimization
        TLSHandshakeTimeout: 10 * time.Second,

        // DNS and connection timeouts
        DialContext: (&net.Dialer{
            Timeout:   30 * time.Second,
            KeepAlive: 30 * time.Second,
        }).DialContext,
    }

    return &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second, // Total request timeout
    }
}

// Production HTTP server configuration
func NewProductionServer(handler http.Handler) *http.Server {
    return &http.Server{
        Addr:    ":8080",
        Handler: handler,

        // Prevent slowloris attacks
        ReadTimeout:       15 * time.Second,
        ReadHeaderTimeout: 10 * time.Second,
        WriteTimeout:      15 * time.Second,
        IdleTimeout:       60 * time.Second,

        // Connection limits
        MaxHeaderBytes: 1 << 20, // 1MB

        // Error logging
        ErrorLog: log.New(os.Stderr, "HTTP: ", log.LstdFlags),
    }
}
```

### Database Batching for High Throughput

```go
// Batched database operations for high throughput
type BatchWriter struct {
    db          *sql.DB
    batchSize   int
    flushPeriod time.Duration

    pending     []BatchItem
    pendingMutex sync.Mutex

    flushChan   chan struct{}
    stopChan    chan struct{}
    wg          sync.WaitGroup
}

type BatchItem struct {
    SQL    string
    Args   []interface{}
    Result chan error
}

func NewBatchWriter(db *sql.DB, batchSize int) *BatchWriter {
    bw := &BatchWriter{
        db:          db,
        batchSize:   batchSize,
        flushPeriod: 100 * time.Millisecond,
        pending:     make([]BatchItem, 0, batchSize),
        flushChan:   make(chan struct{}, 1),
        stopChan:    make(chan struct{}),
    }

    bw.wg.Add(1)
    go bw.flushWorker()

    return bw
}

func (bw *BatchWriter) Execute(sql string, args ...interface{}) error {
    resultChan := make(chan error, 1)

    bw.pendingMutex.Lock()
    bw.pending = append(bw.pending, BatchItem{
        SQL:    sql,
        Args:   args,
        Result: resultChan,
    })

    shouldFlush := len(bw.pending) >= bw.batchSize
    bw.pendingMutex.Unlock()

    if shouldFlush {
        select {
        case bw.flushChan <- struct{}{}:
        default: // Don't block if flush is already queued
        }
    }

    return <-resultChan
}

func (bw *BatchWriter) flushWorker() {
    defer bw.wg.Done()
    ticker := time.NewTicker(bw.flushPeriod)
    defer ticker.Stop()

    for {
        select {
        case <-bw.flushChan:
            bw.flush()
        case <-ticker.C:
            bw.flush()
        case <-bw.stopChan:
            bw.flush()
            return
        }
    }
}

func (bw *BatchWriter) flush() {
    bw.pendingMutex.Lock()
    if len(bw.pending) == 0 {
        bw.pendingMutex.Unlock()
        return
    }

    batch := make([]BatchItem, len(bw.pending))
    copy(batch, bw.pending)
    bw.pending = bw.pending[:0] // Reset slice
    bw.pendingMutex.Unlock()

    // Execute all items in transaction
    tx, err := bw.db.Begin()
    if err != nil {
        // Send error to all waiting items
        for _, item := range batch {
            item.Result <- err
        }
        return
    }

    for _, item := range batch {
        _, execErr := tx.Exec(item.SQL, item.Args...)
        item.Result <- execErr
    }

    if err := tx.Commit(); err != nil {
        tx.Rollback()
    }
}
```

### Production HTTP Server Configuration

```go
// Production-ready server with all timeouts
func NewProductionServer(handler http.Handler, port string) *http.Server {
    return &http.Server{
        Addr:    ":" + port,
        Handler: handler,

        // Prevent slow client attacks
        ReadTimeout:    10 * time.Second,  // Time to read request
        ReadHeaderTimeout: 5 * time.Second, // Time to read headers only

        // Prevent slow response attacks
        WriteTimeout:   10 * time.Second,  // Time to write response

        // Keep-alive timeout
        IdleTimeout:    120 * time.Second, // Keep-alive idle timeout

        // Prevent large header attacks
        MaxHeaderBytes: 1 << 20, // 1MB max headers

        // Error logging
        ErrorLog: log.New(os.Stderr, "HTTP-SERVER ", log.LstdFlags),
    }
}

// Graceful shutdown pattern
func (s *Server) Shutdown(ctx context.Context) error {
    // Stop accepting new connections
    if err := s.httpServer.Shutdown(ctx); err != nil {
        return fmt.Errorf("server shutdown failed: %w", err)
    }

    // Wait for background workers
    s.workerPool.Stop()

    // Close database connections
    if err := s.db.Close(); err != nil {
        return fmt.Errorf("database close failed: %w", err)
    }

    return nil
}
```

## Database Optimization Patterns

### Batch Operations for High Throughput

```go
// High-performance batch inserter (3-5x faster than individual inserts)
type BatchInserter struct {
    db           *sql.DB
    stmt         *sql.Stmt
    batchSize    int
    flushTimeout time.Duration

    mu      sync.Mutex
    pending []BatchItem
    timer   *time.Timer
    closed  atomic.Bool
}

type BatchItem struct {
    ID    uint64
    Data  []byte
    Type  string
}

func NewBatchInserter(db *sql.DB, batchSize int, flushTimeout time.Duration) *BatchInserter {
    stmt, err := db.Prepare(`
        INSERT INTO items (id, data, type, created_at)
        VALUES ($1, $2, $3, $4)
    `)
    if err != nil {
        panic(fmt.Sprintf("failed to prepare batch statement: %v", err))
    }

    bi := &BatchInserter{
        db:           db,
        stmt:         stmt,
        batchSize:    batchSize,
        flushTimeout: flushTimeout,
        pending:      make([]BatchItem, 0, batchSize),
        timer:        time.NewTimer(flushTimeout),
    }

    go bi.flushPeriodically()
    return bi
}

func (bi *BatchInserter) Add(item BatchItem) error {
    if bi.closed.Load() {
        return ErrBatcherClosed
    }

    bi.mu.Lock()
    defer bi.mu.Unlock()

    bi.pending = append(bi.pending, item)

    // Flush when batch is full
    if len(bi.pending) >= bi.batchSize {
        return bi.flushLocked()
    }

    // Reset timer for periodic flush
    if !bi.timer.Stop() {
        <-bi.timer.C
    }
    bi.timer.Reset(bi.flushTimeout)

    return nil
}

func (bi *BatchInserter) flushLocked() error {
    if len(bi.pending) == 0 {
        return nil
    }

    // Use transaction for atomic batch
    tx, err := bi.db.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    txStmt := tx.Stmt(bi.stmt)
    now := time.Now()

    for _, item := range bi.pending {
        if _, err := txStmt.Exec(item.ID, item.Data, item.Type, now); err != nil {
            return fmt.Errorf("failed to execute batch item: %w", err)
        }
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit batch: %w", err)
    }

    // Reset pending slice (keep capacity)
    bi.pending = bi.pending[:0]
    return nil
}

func (bi *BatchInserter) flushPeriodically() {
    for !bi.closed.Load() {
        select {
        case <-bi.timer.C:
            bi.mu.Lock()
            _ = bi.flushLocked()
            bi.mu.Unlock()
        }
    }
}

// PostgreSQL-specific COPY optimization (10x faster for bulk inserts)
func BulkInsertWithCopy(db *sql.DB, items []BatchItem) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    stmt, err := tx.Prepare(pq.CopyIn("items", "id", "data", "type", "created_at"))
    if err != nil {
        return err
    }
    defer stmt.Close()

    now := time.Now()
    for _, item := range items {
        if _, err := stmt.Exec(item.ID, item.Data, item.Type, now); err != nil {
            return err
        }
    }

    if _, err := stmt.Exec(); err != nil {
        return err
    }

    return tx.Commit()
}
```

### Connection Pool Optimization

```go
// Production database configuration
func ConfigureDBPool(db *sql.DB) {
    // Connection pool settings
    db.SetMaxOpenConns(25)                 // Limit concurrent connections
    db.SetMaxIdleConns(10)                 // Keep some idle connections
    db.SetConnMaxLifetime(5 * time.Minute) // Rotate connections
    db.SetConnMaxIdleTime(1 * time.Minute) // Close idle connections
}

// Database health check with connection validation
func (s *Service) HealthCheck(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    if err := s.db.PingContext(ctx); err != nil {
        return fmt.Errorf("database ping failed: %w", err)
    }

    // Validate connection pool state
    stats := s.db.Stats()
    if stats.OpenConnections > stats.MaxOpenConnections*8/10 {
        return fmt.Errorf("connection pool near capacity: %d/%d",
            stats.OpenConnections, stats.MaxOpenConnections)
    }

    return nil
}
```

## JSON & Type Safety Optimization

### Structured Unmarshaling (10x Performance Gain)

```go
// High-performance JSON processing with pools
var userPool = sync.Pool{
    New: func() interface{} {
        return &User{}
    },
}

var jsonDecoderPool = sync.Pool{
    New: func() interface{} {
        return json.NewDecoder(nil)
    },
}

type User struct {
    ID       uint64    `json:"id"`
    Name     string    `json:"name"`
    Email    string    `json:"email"`
    Role     UserRole  `json:"role"`
    Created  time.Time `json:"created_at"`
    Metadata UserMetadata `json:"metadata,omitempty"`
}

type UserMetadata struct {
    LastLogin    *time.Time `json:"last_login,omitempty"`
    LoginCount   int        `json:"login_count"`
    Preferences  map[string]string `json:"preferences,omitempty"`
}

// Fast JSON decoding with object reuse
func DecodeUser(r io.Reader) (*User, error) {
    user := userPool.Get().(*User)

    decoder := jsonDecoderPool.Get().(*json.Decoder)
    decoder.Reset(r)

    err := decoder.Decode(user)

    jsonDecoderPool.Put(decoder)

    if err != nil {
        userPool.Put(user)
        return nil, fmt.Errorf("failed to decode user: %w", err)
    }

    return user, nil
}

func ReleaseUser(user *User) {
    // Reset struct to zero value
    *user = User{}
    userPool.Put(user)
}

// Avoid interface{} in hot paths - use generics instead
func ProcessData[T any](data T, processor func(T) error) error {
    // No runtime type assertions - compile-time type safety
    return processor(data)
}

// Bad: Runtime overhead with interface{}
// func processGeneric(data interface{}) error {
//     user, ok := data.(*User)
//     if !ok {
//         return ErrInvalidType
//     }
//     return processUser(user)
// }

// Good: Compile-time type safety
func processUser(user *User) error {
    if user.ID == 0 {
        return ErrInvalidUserID
    }

    if user.Email == "" {
        return ErrMissingEmail
    }

    return validateUserData(user)
}
```

## Worker Pool & Concurrency Control

### Production Goroutine Management

```go
// Production worker pool with backpressure
type WorkerPool struct {
    maxWorkers   int
    workerFunc   func(context.Context, interface{}) error
    jobQueue     chan Job
    workerSem    chan struct{} // Semaphore for worker limit

    wg           sync.WaitGroup
    ctx          context.Context
    cancel       context.CancelFunc

    // Metrics
    processed    atomic.Uint64
    errors       atomic.Uint64
    queueSize    atomic.Int64
}

type Job struct {
    ID      string
    Data    interface{}
    Context context.Context
}

func NewWorkerPool(maxWorkers int, queueSize int, workerFunc func(context.Context, interface{}) error) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())

    return &WorkerPool{
        maxWorkers: maxWorkers,
        workerFunc: workerFunc,
        jobQueue:   make(chan Job, queueSize),
        workerSem:  make(chan struct{}, maxWorkers),
        ctx:        ctx,
        cancel:     cancel,
    }
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.maxWorkers; i++ {
        wp.wg.Add(1)
        go wp.worker()
    }
}

func (wp *WorkerPool) worker() {
    defer wp.wg.Done()

    for {
        select {
        case job := <-wp.jobQueue:
            wp.queueSize.Add(-1)

            // Acquire worker semaphore
            wp.workerSem <- struct{}{}

            if err := wp.workerFunc(job.Context, job.Data); err != nil {
                wp.errors.Add(1)
                log.Printf("Worker error for job %s: %v", job.ID, err)
            } else {
                wp.processed.Add(1)
            }

            // Release worker semaphore
            <-wp.workerSem

        case <-wp.ctx.Done():
            return
        }
    }
}

func (wp *WorkerPool) Submit(job Job) error {
    select {
    case wp.jobQueue <- job:
        wp.queueSize.Add(1)
        return nil
    case <-wp.ctx.Done():
        return ErrPoolClosed
    default:
        return ErrPoolFull // Apply backpressure
    }
}

func (wp *WorkerPool) Stop(timeout time.Duration) error {
    wp.cancel()

    // Wait for workers to finish with timeout
    done := make(chan struct{})
    go func() {
        wp.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        return nil
    case <-time.After(timeout):
        return ErrShutdownTimeout
    }
}

func (wp *WorkerPool) Stats() PoolStats {
    return PoolStats{
        MaxWorkers:    wp.maxWorkers,
        ActiveWorkers: len(wp.workerSem),
        QueueSize:     int(wp.queueSize.Load()),
        Processed:     wp.processed.Load(),
        Errors:        wp.errors.Load(),
    }
}

// Rate limiting with token bucket
type RateLimiter struct {
    tokens chan struct{}
    ticker *time.Ticker
    done   chan struct{}
}

func NewRateLimiter(rps int) *RateLimiter {
    rl := &RateLimiter{
        tokens: make(chan struct{}, rps),
        ticker: time.NewTicker(time.Second / time.Duration(rps)),
        done:   make(chan struct{}),
    }

    // Pre-fill bucket
    for i := 0; i < rps; i++ {
        rl.tokens <- struct{}{}
    }

    go rl.refillTokens()
    return rl
}

func (rl *RateLimiter) refillTokens() {
    for {
        select {
        case <-rl.ticker.C:
            select {
            case rl.tokens <- struct{}{}:
            default: // Bucket is full
            }
        case <-rl.done:
            return
        }
    }
}

func (rl *RateLimiter) Allow() bool {
    select {
    case <-rl.tokens:
        return true
    default:
        return false
    }
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
    select {
    case <-rl.tokens:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

## Memory & Buffer Management

### Production Buffer Pooling

```go
// Multi-size buffer pool for different use cases
type BufferPool struct {
    small  sync.Pool // 1KB buffers
    medium sync.Pool // 16KB buffers
    large  sync.Pool // 64KB buffers

    stats struct {
        smallHits  atomic.Uint64
        mediumHits atomic.Uint64
        largeHits  atomic.Uint64
        allocations atomic.Uint64
    }
}

func NewBufferPool() *BufferPool {
    return &BufferPool{
        small: sync.Pool{
            New: func() interface{} {
                return make([]byte, 0, 1024) // 1KB
            },
        },
        medium: sync.Pool{
            New: func() interface{} {
                return make([]byte, 0, 16*1024) // 16KB
            },
        },
        large: sync.Pool{
            New: func() interface{} {
                return make([]byte, 0, 64*1024) // 64KB
            },
        },
    }
}

func (bp *BufferPool) Get(size int) []byte {
    switch {
    case size <= 1024:
        bp.stats.smallHits.Add(1)
        return bp.small.Get().([]byte)[:0]
    case size <= 16*1024:
        bp.stats.mediumHits.Add(1)
        return bp.medium.Get().([]byte)[:0]
    case size <= 64*1024:
        bp.stats.largeHits.Add(1)
        return bp.large.Get().([]byte)[:0]
    default:
        bp.stats.allocations.Add(1)
        return make([]byte, 0, size)
    }
}

func (bp *BufferPool) Put(buf []byte) {
    capacity := cap(buf)
    switch {
    case capacity == 1024:
        bp.small.Put(buf[:0])
    case capacity == 16*1024:
        bp.medium.Put(buf[:0])
    case capacity == 64*1024:
        bp.large.Put(buf[:0])
    // Don't pool non-standard sizes
    }
}

// Usage in HTTP handler
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
    buf := s.bufferPool.Get(1024)
    defer s.bufferPool.Put(buf)

    // Read request body
    buf = buf[:cap(buf)]
    n, err := r.Body.Read(buf)
    if err != nil && err != io.EOF {
        http.Error(w, "Failed to read body", http.StatusBadRequest)
        return
    }
    buf = buf[:n]

    // Process request...
    response := s.processRequest(buf)
    w.Write(response)
}
```

Ces patterns de production sont **éprouvés à 2M+ utilisateurs** et complètent parfaitement notre framework d'optimisation go-perfbook. Ils se concentrent sur les **goulots d'étranglement réels** rencontrés en production : connection pooling, batch operations, type safety, worker pools, et memory management.
````
