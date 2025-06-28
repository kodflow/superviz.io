# Go Testing Patterns

## Test Structure (Mandatory)

### Table-Driven Tests

```go
// ✅ ALWAYS: Use table-driven tests
func TestCalculate(t *testing.T) {
    tests := []struct {
        name    string
        input   int
        want    int
        wantErr bool
    }{
        {
            name:    "positive_number",
            input:   5,
            want:    25,
            wantErr: false,
        },
        {
            name:    "zero",
            input:   0,
            want:    0,
            wantErr: false,
        },
        {
            name:    "negative_number",
            input:   -5,
            want:    0,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Calculate(tt.input)

            if tt.wantErr {
                require.Error(t, err)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Timeout Patterns

### Test Timeouts

```go
// ✅ ALWAYS: Set test timeouts
func TestWithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Run test with timeout
    done := make(chan bool)
    go func() {
        // Test logic here
        result := performOperation(ctx)
        assert.NotNil(t, result)
        done <- true
    }()

    select {
    case <-done:
        // Test completed
    case <-ctx.Done():
        t.Fatal("test timeout exceeded")
    }
}

// ✅ Per-test timeouts in table tests
func TestOperations(t *testing.T) {
    tests := []struct {
        name    string
        timeout time.Duration
        fn      func(context.Context) error
    }{
        {
            name:    "fast_operation",
            timeout: 100 * time.Millisecond,
            fn:      fastOperation,
        },
        {
            name:    "slow_operation",
            timeout: 5 * time.Second,
            fn:      slowOperation,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
            defer cancel()

            err := tt.fn(ctx)
            require.NoError(t, err)
        })
    }
}
```

## Parallel Testing

### Concurrent Test Safety

```go
// ✅ ALWAYS: Use t.Parallel() for independent tests
func TestParallel(t *testing.T) {
    t.Parallel() // Mark test as parallel-safe

    // Test implementation
}

// ✅ Test concurrent access
func TestConcurrentAccess(t *testing.T) {
    service := NewService()

    const numGoroutines = 100
    var wg sync.WaitGroup
    errors := make(chan error, numGoroutines)

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            if err := service.Process(ctx, id); err != nil {
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
            t.Errorf("concurrent error: %v", err)
        }
    case <-ctx.Done():
        t.Fatal("test timeout")
    }
}
```

## Mock Patterns

### Interface Mocking

```go
//go:generate mockgen -source=storage.go -destination=mocks/mock_storage.go

// Storage interface for mocking
type Storage interface {
    Save(ctx context.Context, key string, data []byte) error
    Load(ctx context.Context, key string) ([]byte, error)
}

// ✅ Test with mocks
func TestServiceWithMock(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockStorage := mocks.NewMockStorage(ctrl)
    service := NewService(mockStorage)

    ctx := context.Background()
    testData := []byte("test")

    // Set expectations
    mockStorage.EXPECT().
        Save(ctx, "key", testData).
        Return(nil).
        Times(1)

    // Execute test
    err := service.Store(ctx, "key", testData)
    require.NoError(t, err)
}
```

## Benchmark Patterns

### Performance Testing

```go
// ✅ Benchmark with memory stats
func BenchmarkOperation(b *testing.B) {
    // Setup
    data := generateTestData(1000)

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        _ = processData(data)
    }
}

// ✅ Comparative benchmarks
func BenchmarkAlgorithms(b *testing.B) {
    sizes := []int{10, 100, 1000, 10000}

    for _, size := range sizes {
        b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
            data := generateTestData(size)

            b.Run("algorithm_v1", func(b *testing.B) {
                b.ResetTimer()
                for i := 0; i < b.N; i++ {
                    _ = algorithmV1(data)
                }
            })

            b.Run("algorithm_v2", func(b *testing.B) {
                b.ResetTimer()
                for i := 0; i < b.N; i++ {
                    _ = algorithmV2(data)
                }
            })
        })
    }
}
```

## Test Helpers

### Reusable Test Functions

```go
// ✅ Test helper functions
func setupTest(t *testing.T) (*Service, func()) {
    t.Helper()

    // Setup
    tmpDir := t.TempDir()
    service := NewService(tmpDir)

    // Cleanup function
    cleanup := func() {
        service.Close()
    }

    return service, cleanup
}

// Usage
func TestFeature(t *testing.T) {
    service, cleanup := setupTest(t)
    defer cleanup()

    // Test implementation
}
```

## Coverage Requirements

### Achieving 100% Coverage

```go
// ✅ Test all paths including errors
func TestCompleteCodePath(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        mockSetup func(*mocks.MockStorage)
        wantErr   bool
        errMsg    string
    }{
        {
            name:  "success_path",
            input: "valid",
            mockSetup: func(m *mocks.MockStorage) {
                m.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
            },
            wantErr: false,
        },
        {
            name:  "validation_error",
            input: "",
            mockSetup: func(m *mocks.MockStorage) {
                // No mock calls expected
            },
            wantErr: true,
            errMsg:  "input is empty",
        },
        {
            name:  "storage_error",
            input: "valid",
            mockSetup: func(m *mocks.MockStorage) {
                m.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any()).
                    Return(errors.New("storage failed"))
            },
            wantErr: true,
            errMsg:  "storage failed",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mock := mocks.NewMockStorage(ctrl)
            tt.mockSetup(mock)

            service := NewService(mock)
            err := service.Process(tt.input)

            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```
