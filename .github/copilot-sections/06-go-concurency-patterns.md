# Go Concurrency Patterns

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
