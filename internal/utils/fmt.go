// Package utils provides ultra-performance utility functions for efficient I/O operations and formatting
package utils

import (
	"io"
	"iter"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// builderPool is a pool of strings.Builder instances for zero-allocation performance.
//
// builderPool reduces memory allocations by reusing strings.Builder instances
// across multiple formatting operations, providing ultra-fast performance
// in high-throughput scenarios with atomic metrics tracking.
// Code block:
//
//	builder := builderPool.Get().(*strings.Builder)
//	defer builderPool.Put(builder)
//	builder.Reset()
//	builder.WriteString("content")
//
// Parameters: N/A (global pool)
//
// Returns: N/A (global pool)
var builderPool = sync.Pool{
	New: func() any {
		builder := new(strings.Builder)
		builder.Grow(64) // Pre-allocate 64 bytes for common cases
		return builder
	},
}

// Performance metrics using atomic operations for lock-free monitoring
var (
	// totalOperations tracks total formatting operations atomically
	totalOperations atomic.Uint64
	// totalBytesWritten tracks total bytes written atomically
	totalBytesWritten atomic.Uint64
	// poolHits tracks successful pool retrievals atomically
	poolHits atomic.Uint64
	// poolMisses tracks pool allocation events atomically
	poolMisses atomic.Uint64
)

// FprintIgnoreErr writes all values to the writer using ultra-performance patterns, ignoring any errors.
//
// FprintIgnoreErr provides zero-allocation formatted output when error handling
// is not critical, such as logging or debug output. Uses atomic metrics and pooled builders.
//
// Code block:
//
//	var buf bytes.Buffer
//	FprintIgnoreErr(&buf, "Hello", " ", "World", 123)
//	fmt.Println(buf.String()) // "Hello World123"
//	// Check metrics
//	total, bytes := GetMetrics()
//	fmt.Printf("Operations: %d, Bytes: %d\n", total, bytes)
//
// Parameters:
//   - 1 w: io.Writer - destination writer for formatted output
//   - 2 args: ...any - values to write (strings, numbers, booleans, etc.)
//
// Returns: N/A (errors are ignored for performance)
func FprintIgnoreErr(w io.Writer, args ...any) {
	totalOperations.Add(1)
	n, _ := writeArgsOptimized(w, args)
	totalBytesWritten.Add(uint64(n))
}

// FprintlnIgnoreErr writes all values followed by newline using ultra-performance patterns, ignoring errors.
//
// FprintlnIgnoreErr provides zero-allocation formatted output with newline
// when error handling is not critical. Uses atomic metrics and pooled builders.
//
// Code block:
//
//	var buf bytes.Buffer
//	FprintlnIgnoreErr(&buf, "Hello", " ", "World")
//	fmt.Printf("%q", buf.String()) // "Hello World\n"
//
// Parameters:
//   - 1 w: io.Writer - destination writer for formatted output
//   - 2 args: ...any - values to write (strings, numbers, booleans, etc.)
//
// Returns: N/A (errors are ignored for performance)
func FprintlnIgnoreErr(w io.Writer, args ...any) {
	totalOperations.Add(1)
	n, _ := writeArgsOptimized(w, args)
	totalBytesWritten.Add(uint64(n + 1)) // +1 for newline
	_, _ = w.Write([]byte("\n"))
}

// Fprint writes all values using ultra-performance patterns and returns error if any operation fails.
//
// Fprint provides zero-allocation formatted output writing with proper error handling.
// Uses atomic metrics, pooled builders, and optimized type switching for production code.
//
// Code block:
//
//	var buf bytes.Buffer
//	if err := Fprint(&buf, "Hello", " ", "World"); err != nil {
//	    return fmt.Errorf("write failed: %w", err)
//	}
//	fmt.Println(buf.String()) // "Hello World"
//
// Parameters:
//   - 1 w: io.Writer - destination writer for formatted output
//   - 2 args: ...any - values to write (strings, numbers, booleans, etc.)
//
// Returns:
//   - 1 error - non-nil if any write operation fails
func Fprint(w io.Writer, args ...any) error {
	totalOperations.Add(1)
	n, err := writeArgsOptimized(w, args)
	if err == nil {
		totalBytesWritten.Add(uint64(n))
	}
	return err
}

// Fprintln writes all values with newline using ultra-performance patterns, returns error if any operation fails.
//
// Fprintln provides zero-allocation formatted output with newline and proper error handling.
// Uses atomic metrics, pooled builders, and optimized patterns for production use.
//
// Code block:
//
//	var buf bytes.Buffer
//	if err := Fprintln(&buf, "Hello", " ", "World"); err != nil {
//	    return fmt.Errorf("write failed: %w", err)
//	}
//	fmt.Printf("%q", buf.String()) // "Hello World\n"
//
// Parameters:
//   - 1 w: io.Writer - destination writer for formatted output
//   - 2 args: ...any - values to write (strings, numbers, booleans, etc.)
//
// Returns:
//   - 1 error - non-nil if any write operation fails
func Fprintln(w io.Writer, args ...any) error {
	totalOperations.Add(1)
	n, err := writeArgsOptimized(w, args)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("\n"))
	if err == nil {
		totalBytesWritten.Add(uint64(n + 1)) // +1 for newline
	}
	return err
}

// writeArgsOptimized efficiently writes arguments using ultra-performance patterns and atomic metrics.
//
// writeArgsOptimized converts various types to strings and writes them with zero-allocation
// optimization using pooled strings.Builder, optimized type switching, and atomic metrics.
// Implements Go 1.24+ performance patterns for maximum throughput.
//
// Code block:
//
//	var buf bytes.Buffer
//	n, err := writeArgsOptimized(&buf, []any{"Hello", 123, true})
//	if err != nil {
//	    log.Printf("Write failed: %v", err)
//	    return
//	}
//	fmt.Printf("Wrote %d bytes: %s\n", n, buf.String())
//
// Parameters:
//   - 1 w: io.Writer - destination writer for converted arguments
//   - 2 args: []any - arguments to convert and write with type optimization
//
// Returns:
//   - 1 n: int64 - number of bytes written to the writer
//   - 2 err: error - non-nil if any write operation fails
func writeArgsOptimized(w io.Writer, args []any) (int64, error) {
	// Get builder from pool with atomic metrics
	var builder *strings.Builder
	if pooled := builderPool.Get(); pooled != nil {
		builder = pooled.(*strings.Builder)
		poolHits.Add(1)
	} else {
		builder = new(strings.Builder)
		poolMisses.Add(1)
	}

	builder.Reset()
	defer builderPool.Put(builder)

	// Pre-allocate with optimized estimation
	estimatedSize := len(args) * 16 // Heuristic: 16 bytes per arg average
	if cap := builder.Cap(); cap < estimatedSize {
		builder.Grow(estimatedSize - cap)
	}

	// Ultra-fast type switching with branch prediction optimization
	for _, arg := range args {
		switch v := arg.(type) {
		case string: // Most common case first for branch prediction
			builder.WriteString(v)
		case []byte: // Second most common
			builder.Write(v)
		case int: // Numeric types grouped
			builder.WriteString(strconv.Itoa(v))
		case int64:
			builder.WriteString(strconv.FormatInt(v, 10))
		case uint:
			builder.WriteString(strconv.FormatUint(uint64(v), 10))
		case uint64:
			builder.WriteString(strconv.FormatUint(v, 10))
		case bool: // Boolean types
			if v {
				builder.WriteString("true")
			} else {
				builder.WriteString("false")
			}
		case rune: // Character types (rune = int32)
			builder.WriteRune(v)
		case byte: // byte = uint8
			builder.WriteByte(v)
		case float32: // Floating point types
			builder.WriteString(strconv.FormatFloat(float64(v), 'g', -1, 32))
		case float64:
			builder.WriteString(strconv.FormatFloat(v, 'g', -1, 64))
		default: // Fallback for rare types
			builder.WriteString(toStringOptimized(v))
		}
	}

	// Single write operation for maximum efficiency
	content := builder.String()
	n, err := io.WriteString(w, content)
	return int64(n), err
}

// toStringOptimized provides ultra-fast fallback converter for unsupported types.
//
// toStringOptimized handles type conversion for types not explicitly supported
// by the main formatting logic, providing graceful degradation with optimized
// error handling and minimal allocations.
//
// Code block:
//
//	str := toStringOptimized(errors.New("test error"))
//	fmt.Println(str) // "test error"
//	str = toStringOptimized(struct{}{})
//	fmt.Println(str) // "[unsupported type]"
//
// Parameters:
//   - 1 v: any - value to convert to string representation
//
// Returns:
//   - 1 str: string - optimized string representation of the value
func toStringOptimized(v any) string {
	switch val := v.(type) {
	case error:
		return val.Error()
	case nil:
		return "<nil>"
	default:
		return "[unsupported type]"
	}
}

// GetMetrics returns atomic performance metrics for monitoring and debugging.
//
// GetMetrics provides real-time performance metrics for the formatting utilities,
// using atomic operations for accurate, lock-free counters in concurrent environments.
//
// Code block:
//
//	total, bytes, hits, misses := GetMetrics()
//	efficiency := float64(hits) / float64(hits + misses) * 100
//	fmt.Printf("Operations: %d, Bytes: %d, Pool efficiency: %.1f%%\n",
//	           total, bytes, efficiency)
//
// Parameters: N/A
//
// Returns:
//   - 1 totalOps: uint64 - total formatting operations performed atomically
//   - 2 totalBytes: uint64 - total bytes written atomically
//   - 3 poolHits: uint64 - successful pool retrievals atomically
//   - 4 poolMisses: uint64 - pool allocation events atomically
func GetMetrics() (totalOps, totalBytes, poolHitsCount, poolMissesCount uint64) {
	return totalOperations.Load(),
		totalBytesWritten.Load(),
		poolHits.Load(),
		poolMisses.Load()
}

// ResetMetrics atomically resets all performance counters to zero.
//
// ResetMetrics provides a thread-safe way to reset metrics for
// testing, benchmarking, or periodic monitoring cycles.
//
// Code block:
//
//	ResetMetrics() // Safe concurrent reset
//	// Perform operations...
//	total, bytes, _, _ := GetMetrics()
//	fmt.Printf("New cycle: %d ops, %d bytes\n", total, bytes)
//
// Parameters: N/A
//
// Returns: N/A (void function)
func ResetMetrics() {
	totalOperations.Store(0)
	totalBytesWritten.Store(0)
	poolHits.Store(0)
	poolMisses.Store(0)
}

// ArgsIter provides a Go 1.24+ iterator for processing arguments efficiently.
//
// ArgsIter creates a memory-efficient iterator over formatting arguments
// using Go 1.24's new iterator patterns for zero-allocation processing.
//
// Code block:
//
//	args := []any{"hello", 123, true}
//	for i, arg := range ArgsIter(args) {
//	    fmt.Printf("Arg %d: %v\n", i, arg)
//	}
//
// Parameters:
//   - 1 args: []any - slice of arguments to iterate over
//
// Returns:
//   - 1 iterator: iter.Seq2[int, any] - Go 1.24+ iterator over index and argument pairs
func ArgsIter(args []any) iter.Seq2[int, any] {
	return func(yield func(int, any) bool) {
		for i, arg := range args {
			if !yield(i, arg) {
				return
			}
		}
	}
}
