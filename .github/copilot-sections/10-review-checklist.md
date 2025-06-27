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
