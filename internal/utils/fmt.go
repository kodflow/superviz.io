package utils

import (
	"io"
	"strconv"
)

// FprintIgnoreErr writes all values to the writer using their default format.
// It avoids fmt.Fprint to prevent allocations on formatting and interface wrapping.
func FprintIgnoreErr(w io.Writer, args ...any) {
	writeArgs(w, args)
}

// FprintlnIgnoreErr writes all values followed by a newline.
func FprintlnIgnoreErr(w io.Writer, args ...any) {
	writeArgs(w, args)
	_, _ = w.Write([]byte("\n"))
}

// MustFprint writes all values or panics on error.
func MustFprint(w io.Writer, args ...any) {
	if err := writeArgsChecked(w, args); err != nil {
		panic("Fprint failed: " + err.Error())
	}
}

// MustFprintln writes all values with newline or panics on error.
func MustFprintln(w io.Writer, args ...any) {
	if err := writeArgsChecked(w, args); err != nil {
		panic("Fprintln failed: " + err.Error())
	}
	if _, err := w.Write([]byte("\n")); err != nil {
		panic("Fprintln write failed: " + err.Error())
	}
}

// writeArgs writes a slice of values with minimal allocation.
// It handles only basic types efficiently (strings, ints, bools, etc.).
func writeArgs(w io.Writer, args []any) {
	var builder strings.Builder
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
	_, _ = w.Write([]byte(builder.String()))
}

// writeArgsChecked is same as writeArgs but returns error (used in Must*)
func writeArgsChecked(w io.Writer, args []any) error {
	var builder strings.Builder
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
	_, err := w.Write([]byte(builder.String()))
	return err
}

// toString is the fallback converter using the fmt logic.
func toString(v any) string {
	// not using fmt.Sprintf to avoid importing fmt
	switch val := v.(type) {
	case error:
		return val.Error()
	default:
		return "[unsupported type]"
	}
}
