# Go Disk Optimization

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
