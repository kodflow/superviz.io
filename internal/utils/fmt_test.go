package utils_test

import (
	"bytes"
	"errors"
	"fmt"
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

func TestFprint(t *testing.T) {
	t.Helper()

	t.Run("successful write", func(t *testing.T) {
		var buf bytes.Buffer
		err := utils.Fprint(&buf, "hello", " ", "world")
		require.NoError(t, err)
		assert.Equal(t, "hello world", buf.String())
	})

	t.Run("returns error on write failure", func(t *testing.T) {
		writer := &failingWriter{failAfter: 0}
		err := utils.Fprint(writer, "hello")
		require.Error(t, err)
		assert.Equal(t, "write error", err.Error())
	})
}

func TestFprintln(t *testing.T) {
	t.Helper()

	t.Run("successful write with newline", func(t *testing.T) {
		var buf bytes.Buffer
		err := utils.Fprintln(&buf, "hello", " ", "world")
		require.NoError(t, err)
		assert.Equal(t, "hello world\n", buf.String())
	})

	t.Run("returns error on args write failure", func(t *testing.T) {
		writer := &failingWriter{failAfter: 0}
		err := utils.Fprintln(writer, "hello")
		require.Error(t, err)
		assert.Equal(t, "write error", err.Error())
	})

	t.Run("returns error on newline write failure", func(t *testing.T) {
		writer := &failingWriter{failAfter: 1}
		err := utils.Fprintln(writer, "hello")
		require.Error(t, err)
		assert.Equal(t, "write error", err.Error())
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
			_, _ = fmt.Fprint(&buf, args...)
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

func TestFprint_ErrorPerType(t *testing.T) {
	tests := []struct {
		name    string
		arg     any
		wantErr string
	}{
		{
			name:    "string error",
			arg:     "fail",
			wantErr: "write error",
		},
		{
			name:    "[]byte error",
			arg:     []byte("fail"),
			wantErr: "write error",
		},
		{
			name:    "int error",
			arg:     42,
			wantErr: "write error",
		},
		{
			name:    "int64 error",
			arg:     int64(123456),
			wantErr: "write error",
		},
		{
			name:    "bool error",
			arg:     true,
			wantErr: "write error",
		},
		{
			name:    "rune error",
			arg:     'x',
			wantErr: "write error",
		},
		{
			name:    "byte error",
			arg:     byte('Z'),
			wantErr: "write error",
		},
		{
			name:    "fallback error",
			arg:     struct{ foo string }{foo: "bar"},
			wantErr: "write error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &typedFailingWriter{failOn: tt.arg}
			err := utils.Fprint(writer, tt.arg)
			require.Error(t, err)
			assert.Equal(t, tt.wantErr, err.Error())
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
