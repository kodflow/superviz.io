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
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			_, _ = w.Write([]byte(v))
		case []byte:
			_, _ = w.Write(v)
		case int:
			_, _ = w.Write([]byte(strconv.Itoa(v)))
		case int64:
			_, _ = w.Write([]byte(strconv.FormatInt(v, 10)))
		case bool:
			_, _ = w.Write([]byte(strconv.FormatBool(v)))
		case rune:
			_, _ = w.Write([]byte(string(v)))
		case byte:
			_, _ = w.Write([]byte{v})
		default:
			// fallback (not zero-alloc)
			_, _ = w.Write([]byte(toString(v)))
		}
	}
}

// writeArgsChecked is same as writeArgs but returns error (used in Must*)
func writeArgsChecked(w io.Writer, args []any) error {
	for _, arg := range args {
		var err error
		switch v := arg.(type) {
		case string:
			_, err = w.Write([]byte(v))
		case []byte:
			_, err = w.Write(v)
		case int:
			_, err = w.Write([]byte(strconv.Itoa(v)))
		case int64:
			_, err = w.Write([]byte(strconv.FormatInt(v, 10)))
		case bool:
			_, err = w.Write([]byte(strconv.FormatBool(v)))
		case rune:
			_, err = w.Write([]byte(string(v)))
		case byte:
			_, err = w.Write([]byte{v})
		default:
			_, err = w.Write([]byte(toString(v)))
		}
		if err != nil {
			return err
		}
	}
	return nil
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
