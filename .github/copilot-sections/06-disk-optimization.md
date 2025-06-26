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
````
