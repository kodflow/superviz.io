package utils

import (
	"io"
	"strconv"
	"strings"
	"sync"
)

var builderPool = sync.Pool{
	New: func() any {
		return new(strings.Builder)
	},
}

// FprintIgnoreErr writes all values to the writer, ignoring errors.
func FprintIgnoreErr(w io.Writer, args ...any) {
	_, _ = writeArgs(w, args)
}

// FprintlnIgnoreErr writes all values followed by a newline, ignoring errors.
func FprintlnIgnoreErr(w io.Writer, args ...any) {
	_, _ = writeArgs(w, args)
	_, _ = w.Write([]byte("\n"))
}

// MustFprint writes all values or panics on error.
func MustFprint(w io.Writer, args ...any) {
	if _, err := writeArgs(w, args); err != nil {
		panic("Fprint failed: " + err.Error())
	}
}

// MustFprintln writes all values with newline or panics on error.
func MustFprintln(w io.Writer, args ...any) {
	if _, err := writeArgs(w, args); err != nil {
		panic("Fprintln failed: " + err.Error())
	}
	if _, err := w.Write([]byte("\n")); err != nil {
		panic("Fprintln write failed: " + err.Error())
	}
}

// writeArgs efficiently writes arguments to the writer.
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

	// Écriture finale du contenu
	n, err := io.WriteString(w, builder.String())
	return int64(n), err
}

func estimatedSize(n int) int {
	return n * 16
}

// toString is the fallback converter for unsupported types.
func toString(v any) string {
	switch val := v.(type) {
	case error:
		return val.Error()
	default:
		return "[unsupported type]"
	}
}
