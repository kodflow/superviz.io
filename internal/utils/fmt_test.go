package utils_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/kodflow/superviz.io/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFprintIgnoreErr(t *testing.T) {
	t.Helper()

	tests := []struct {
		name     string
		args     []any
		expected string
	}{
		{
			name:     "empty args",
			args:     []any{},
			expected: "",
		},
		{
			name:     "single string",
			args:     []any{"hello"},
			expected: "hello",
		},
		{
			name:     "multiple strings",
			args:     []any{"hello", " ", "world"},
			expected: "hello world",
		},
		{
			name:     "mixed types",
			args:     []any{"count: ", 42, " active: ", true},
			expected: "count: 42 active: true",
		},
		{
			name:     "int types",
			args:     []any{123, int64(456)},
			expected: "123456",
		},
		{
			name:     "byte slice",
			args:     []any{[]byte("bytes")},
			expected: "bytes",
		},
		{
			name:     "rune and byte",
			args:     []any{'A', byte('B')},
			expected: "AB",
		},
		{
			name:     "boolean values",
			args:     []any{true, false},
			expected: "truefalse",
		},
		{
			name:     "error type fallback",
			args:     []any{errors.New("test error")},
			expected: "test error",
		},
		{
			name:     "unsupported type fallback",
			args:     []any{struct{ name string }{name: "test"}},
			expected: "[unsupported type]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			utils.FprintIgnoreErr(&buf, tt.args...)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestFprintlnIgnoreErr(t *testing.T) {
	t.Helper()

	tests := []struct {
		name     string
		args     []any
		expected string
	}{
		{
			name:     "empty args with newline",
			args:     []any{},
			expected: "\n",
		},
		{
			name:     "single string with newline",
			args:     []any{"hello"},
			expected: "hello\n",
		},
		{
			name:     "multiple args with newline",
			args:     []any{"hello", " ", "world"},
			expected: "hello world\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			utils.FprintlnIgnoreErr(&buf, tt.args...)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestMustFprint(t *testing.T) {
	t.Helper()

	t.Run("successful write", func(t *testing.T) {
		var buf bytes.Buffer
		require.NotPanics(t, func() {
			utils.MustFprint(&buf, "hello", " ", "world")
		})
		assert.Equal(t, "hello world", buf.String())
	})

	t.Run("panics on write error", func(t *testing.T) {
		writer := &failingWriter{failAfter: 0}
		assert.PanicsWithValue(t, "Fprint failed: write error", func() {
			utils.MustFprint(writer, "hello")
		})
	})
}

func TestMustFprintln(t *testing.T) {
	t.Helper()

	t.Run("successful write with newline", func(t *testing.T) {
		var buf bytes.Buffer
		require.NotPanics(t, func() {
			utils.MustFprintln(&buf, "hello", " ", "world")
		})
		assert.Equal(t, "hello world\n", buf.String())
	})

	t.Run("panics on args write error", func(t *testing.T) {
		writer := &failingWriter{failAfter: 0}
		assert.PanicsWithValue(t, "Fprintln failed: write error", func() {
			utils.MustFprintln(writer, "hello")
		})
	})

	t.Run("panics on newline write error", func(t *testing.T) {
		writer := &failingWriter{failAfter: 1}
		assert.PanicsWithValue(t, "Fprintln write failed: write error", func() {
			utils.MustFprintln(writer, "hello")
		})
	})
}

func TestAllTypesConversion(t *testing.T) {
	t.Helper()

	var buf bytes.Buffer
	utils.FprintIgnoreErr(&buf,
		"string",
		[]byte("bytes"),
		42,
		int64(9223372036854775807),
		true,
		false,
		'🚀',
		byte(65),
		errors.New("error message"),
		struct{ field int }{field: 123},
	)

	expected := "stringbytes429223372036854775807truefalse🚀Aerror message[unsupported type]"
	assert.Equal(t, expected, buf.String())
}

func BenchmarkFprintIgnoreErr(b *testing.B) {
	var buf bytes.Buffer
	args := []any{"prefix: ", 12345, " suffix: ", true, " end"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		utils.FprintIgnoreErr(&buf, args...)
	}
}

func BenchmarkFprintIgnoreErrVsStdFmt(b *testing.B) {
	var buf bytes.Buffer
	args := []any{"prefix: ", 12345, " suffix: ", true, " end"}

	b.Run("utils.FprintIgnoreErr", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf.Reset()
			utils.FprintIgnoreErr(&buf, args...)
		}
	})

	b.Run("fmt.Fprint", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf.Reset()
			_, _ = strings.NewReplacer().WriteString(&buf, "prefix: 12345 suffix: true end")
		}
	})
}

// failingWriter is a test helper that fails after a specified number of writes
type failingWriter struct {
	writes    int
	failAfter int
}

func (w *failingWriter) Write(p []byte) (n int, err error) {
	if w.writes >= w.failAfter {
		return 0, errors.New("write error")
	}
	w.writes++
	return len(p), nil
}

func TestWriteArgsChecked_ErrorPerType(t *testing.T) {
	tests := []struct {
		name      string
		arg       any
		wantPanic string
	}{
		{
			name:      "string error",
			arg:       "fail",
			wantPanic: "Fprint failed: write error",
		},
		{
			name:      "[]byte error",
			arg:       []byte("fail"),
			wantPanic: "Fprint failed: write error",
		},
		{
			name:      "int error",
			arg:       42,
			wantPanic: "Fprint failed: write error",
		},
		{
			name:      "int64 error",
			arg:       int64(123456),
			wantPanic: "Fprint failed: write error",
		},
		{
			name:      "bool error",
			arg:       true,
			wantPanic: "Fprint failed: write error",
		},
		{
			name:      "rune error",
			arg:       'x',
			wantPanic: "Fprint failed: write error",
		},
		{
			name:      "byte error",
			arg:       byte('Z'),
			wantPanic: "Fprint failed: write error",
		},
		{
			name:      "fallback error",
			arg:       struct{ foo string }{foo: "bar"},
			wantPanic: "Fprint failed: write error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &typedFailingWriter{failOn: tt.arg}
			assert.PanicsWithValue(t, tt.wantPanic, func() {
				utils.MustFprint(writer, tt.arg)
			})
		})
	}
}

// typedFailingWriter fails only when writing a specific value (based on string content).
type typedFailingWriter struct {
	failOn any
}

func (w *typedFailingWriter) Write(p []byte) (int, error) {
	// match if we're inside the expected failing value
	failStr := convertToExpectedString(w.failOn)
	if string(p) == failStr || (len(p) == 1 && string(p)[0] == failStr[0]) {
		return 0, errors.New("write error")
	}
	return len(p), nil
}

// convertToExpectedString is used to produce the expected string for matching in writer.
func convertToExpectedString(v any) string {
	var buf bytes.Buffer
	utils.FprintIgnoreErr(&buf, v)
	return buf.String()
}
