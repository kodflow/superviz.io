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
