## 100% Test Coverage with Timeouts

### Mandatory Test Structure

```go
// FunctionName is the function we're testing
func FunctionName(ctx context.Context, input string) (string, error) {
    if input == "" {
        return "", errors.New("empty input")
    }
    return "processed_" + input, nil
}

func TestFunctionName(t *testing.T) {
    // ALWAYS set test timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
        timeout time.Duration // Per-test timeout
    }{
        {
            name:    "valid_input",
            input:   "test",
            want:    "processed_test",
            wantErr: false,
            timeout: time.Second,
        },
        {
            name:    "empty_input",
            input:   "",
            want:    "",
            wantErr: true,
            timeout: 100 * time.Millisecond,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Helper()
            
            // Per-test timeout
            testCtx, testCancel := context.WithTimeout(ctx, tt.timeout)
            defer testCancel()
            
            // Use require for fatal assertions
            require.NotNil(t, testCtx)
            
            got, err := FunctionName(testCtx, tt.input)
            
            if tt.wantErr {
                require.Error(t, err)
                assert.Empty(t, got)
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Concurrent Test Safety
```go
// Service represents our service for testing
type Service struct {
    storage Storage
}

func NewService(storage Storage) *Service {
    return &Service{storage: storage}
}

func (s *Service) Process(ctx context.Context, data string) error {
    if data == "" {
        return errors.New("empty data")
    }
    // Simulate processing
    return nil
}

func (s *Service) ProcessData(ctx context.Context, key string, data []byte) error {
    return s.storage.Save(ctx, key, data)
}

func TestConcurrentAccess(t *testing.T) {
    t.Parallel()
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    const numGoroutines = 100
    var wg sync.WaitGroup
    errors := make(chan error, numGoroutines)
    
    service := NewService(nil) // Using nil storage for this test
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            select {
            case <-ctx.Done():
                errors <- ctx.Err()
                return
            default:
            }
            
            if err := service.Process(ctx, fmt.Sprintf("data-%d", id)); err != nil {
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
            t.Errorf("goroutine error: %v", err)
        }
    case <-ctx.Done():
        t.Fatal("test timeout exceeded")
    }
}
}
```

### Mock Generation (100% Coverage)
```go
//go:generate mockgen -source=interfaces.go -destination=mocks/mock_interfaces.go

type Storage interface {
    Save(ctx context.Context, key string, data []byte) error
    Load(ctx context.Context, key string) ([]byte, error)
    Delete(ctx context.Context, key string) error
}

// Test with mocks
func TestServiceWithMock(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    mockStorage := mocks.NewMockStorage(ctrl)
    service := NewService(mockStorage)
    
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    
    // Setup expectations
    mockStorage.EXPECT().
        Save(gomock.Any(), "test-key", gomock.Any()).
        Return(nil).
        Times(1)
    
    err := service.ProcessData(ctx, "test-key", []byte("test-data"))
    require.NoError(t, err)
}
```
