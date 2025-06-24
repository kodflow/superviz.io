// Package utils provides utility functions for efficient I/O operations and formatting
package utils

import (
	"io"
	"strconv"
	"strings"
	"sync"
)

// builderPool is a pool of strings.Builder instances for memory efficiency.
//
// builderPool reduces memory allocations by reusing strings.Builder instances
// across multiple formatting operations, improving performance in high-throughput scenarios.
var builderPool = sync.Pool{
	New: func() any {
		return new(strings.Builder)
	},
}

// FprintIgnoreErr writes all values to the writer, ignoring any errors.
//
// FprintIgnoreErr provides a convenient way to write formatted output when
// error handling is not critical, such as logging or debug output.
//
// Example:
//
//	var buf bytes.Buffer
//	FprintIgnoreErr(&buf, "Hello", " ", "World", 123)
//	fmt.Println(buf.String()) // "Hello World123"
//
// Parameters:
//   - w: io.Writer to output the formatted values
//   - args: ...any values to write (supports strings, numbers, booleans, etc.)
//
// Returns:
//   - None (errors are ignored)
func FprintIgnoreErr(w io.Writer, args ...any) {
	_, _ = writeArgs(w, args)
}

// FprintlnIgnoreErr writes all values followed by a newline, ignoring any errors.
//
// FprintlnIgnoreErr is similar to FprintIgnoreErr but automatically appends
// a newline character to the output.
//
// Example:
//
//	var buf bytes.Buffer
//	FprintlnIgnoreErr(&buf, "Hello", " ", "World")
//	fmt.Printf("%q", buf.String()) // "Hello World\n"
//
// Parameters:
//   - w: io.Writer to output the formatted values
//   - args: ...any values to write (supports strings, numbers, booleans, etc.)
//
// Returns:
//   - None (errors are ignored)
func FprintlnIgnoreErr(w io.Writer, args ...any) {
	_, _ = writeArgs(w, args)
	_, _ = w.Write([]byte("\n"))
}

// MustFprint writes all values to the writer or panics on error.
//
// MustFprint provides a fail-fast approach to output formatting, panicking
// if any write operation fails. Use when write failures are unrecoverable.
//
// Example:
//
//	var buf bytes.Buffer
//	MustFprint(&buf, "Hello", " ", "World")
//	fmt.Println(buf.String()) // "Hello World"
//
// Parameters:
//   - w: io.Writer to output the formatted values
//   - args: ...any values to write (supports strings, numbers, booleans, etc.)
//
// Returns:
//   - None (panics on error)
//
// Panics:
//   - If any write operation fails
func MustFprint(w io.Writer, args ...any) {
	if _, err := writeArgs(w, args); err != nil {
		panic("Fprint failed: " + err.Error())
	}
}

// MustFprintln writes all values with newline to the writer or panics on error.
//
// MustFprintln is similar to MustFprint but automatically appends a newline
// character and panics if any write operation fails.
//
// Example:
//
//	var buf bytes.Buffer
//	MustFprintln(&buf, "Hello", " ", "World")
//	fmt.Printf("%q", buf.String()) // "Hello World\n"
//
// Parameters:
//   - w: io.Writer to output the formatted values
//   - args: ...any values to write (supports strings, numbers, booleans, etc.)
//
// Returns:
//   - None (panics on error)
//
// Panics:
//   - If any write operation fails
func MustFprintln(w io.Writer, args ...any) {
	if _, err := writeArgs(w, args); err != nil {
		panic("Fprintln failed: " + err.Error())
	}
	if _, err := w.Write([]byte("\n")); err != nil {
		panic("Fprintln write failed: " + err.Error())
	}
}

// writeArgs efficiently writes arguments to the writer using pooled string builders.
//
// writeArgs converts various types to strings and writes them efficiently
// using a pooled strings.Builder to minimize memory allocations.
//
// Example:
//
//	var buf bytes.Buffer
//	n, err := writeArgs(&buf, []any{"Hello", 123, true})
//	// buf contains "Hello123true", n is bytes written
//
// Parameters:
//   - w: io.Writer to output the converted arguments
//   - args: []any arguments to convert and write
//
// Returns:
//   - n: int64 number of bytes written
//   - err: error if any write operation fails
func writeArgs(w io.Writer, args []any) (int64, error) {
	builder := builderPool.Get().(*strings.Builder)
	builder.Reset()
	defer builderPool.Put(builder)

	builder.Grow(estimatedSize(len(args)))

	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			builder.WriteString(v)
		case []byte:
			builder.Write(v)
		case int:
			builder.WriteString(strconv.Itoa(v))
		case int64:
			builder.WriteString(strconv.FormatInt(v, 10))
		case bool:
			builder.WriteString(strconv.FormatBool(v))
		case rune:
			builder.WriteRune(v)
		case byte:
			builder.WriteByte(v)
		default:
			builder.WriteString(toString(v))
		}
	}

	// Final write of content
	n, err := io.WriteString(w, builder.String())
	return int64(n), err
}

// estimatedSize calculates an estimated buffer size for the given number of arguments.
//
// estimatedSize provides a heuristic for pre-allocating string builder capacity
// to reduce memory reallocations during string building operations.
//
// Parameters:
//   - n: Number of arguments to estimate size for
//
// Returns:
//   - Estimated size in bytes
//
// estimatedSize calculates the estimated buffer size needed for string conversion.
//
// estimatedSize provides a heuristic for pre-allocating string builder capacity
// to reduce memory reallocations during argument conversion.
//
// Example:
//
//	size := estimatedSize(5) // Returns 80 (5 * 16)
//
// Parameters:
//   - n: int number of arguments to be converted
//
// Returns:
//   - size: int estimated buffer size in bytes
func estimatedSize(n int) int {
	return n * 16
}

// toString is the fallback converter for unsupported types.
//
// toString handles type conversion for types not explicitly supported
// by the main formatting logic, providing graceful degradation.
//
// Example:
//
//	str := toString(errors.New("test error"))
//	// Returns "test error"
//	str = toString(struct{}{})
//	// Returns "[unsupported type]"
//
// Parameters:
//   - v: any value to convert to string
//
// Returns:
//   - str: string representation of the value
func toString(v any) string {
	switch val := v.(type) {
	case error:
		return val.Error()
	default:
		return "[unsupported type]"
	}
}
