# Go Memory Optimization

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

## Memory Profiling Commands

```bash
# Run with memory profiling
go test -bench=. -memprofile=mem.prof

# Analyze allocations
go tool pprof -alloc_space mem.prof

# Check escape analysis
go build -gcflags="-m -m" ./...
```
