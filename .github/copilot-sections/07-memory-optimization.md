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

```go
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
```

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
