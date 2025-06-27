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
