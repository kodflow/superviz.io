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
