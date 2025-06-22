package version_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kodflow/superviz.io/internal/cli/commands/version"
	"github.com/kodflow/superviz.io/internal/providers"
	"github.com/kodflow/superviz.io/internal/services"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func newVersionCmdWithBuffer(buf *bytes.Buffer) *cobra.Command {
	cmd := version.GetCommand()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	return cmd
}

func TestGetCommand_Singleton(t *testing.T) {
	t.Parallel()

	cmd1 := version.GetCommand()
	cmd2 := version.GetCommand()

	// Should return the same singleton instance
	require.Same(t, cmd1, cmd2)
}

func TestGetCommand_Basic(t *testing.T) {
	t.Parallel()

	cmd := version.GetCommand()

	require.NotNil(t, cmd)
	require.Equal(t, "version", cmd.Use)
	require.Equal(t, "Print version information", cmd.Short)
}

func TestGetCommandWithService_NilService(t *testing.T) {
	t.Parallel()

	cmd1 := version.GetCommandWithService(nil)
	cmd2 := version.GetCommand()

	// Should return the same singleton when service is nil
	require.Same(t, cmd1, cmd2)
}

func TestGetCommandWithService_CustomService(t *testing.T) {
	t.Parallel()

	// Reset singleton state for this test
	providers.Reset()

	// Create mock provider with custom values
	mockProvider := &mockVersionProvider{
		info: providers.VersionInfo{
			Version:   "custom-version",
			Commit:    "custom-commit",
			BuiltAt:   "custom-date",
			BuiltBy:   "custom-user",
			GoVersion: "go1.21.0",
			OSArch:    "linux/amd64",
		},
	}

	// Create service with mock provider
	customService := services.NewVersionService(mockProvider)

	// Get command with custom service
	cmd := version.GetCommandWithService(customService)
	defaultCmd := version.GetCommand()

	// Should return different instances
	require.NotSame(t, cmd, defaultCmd)

	// Test execution
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{}) // Ensure no arguments

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	require.NotEmpty(t, output)

	// Verify custom values are present
	require.Contains(t, output, "custom-version", "should contain custom version")
	require.Contains(t, output, "custom-commit", "should contain custom commit")
	require.Contains(t, output, "custom-date", "should contain custom date")
	require.Contains(t, output, "custom-user", "should contain custom user")
	require.Contains(t, output, "go1.21.0", "should contain custom go version")
	require.Contains(t, output, "linux/amd64", "should contain custom OS/Arch")

	// Verify it's not using default values
	require.NotContains(t, output, "dev", "should not contain default version")
	require.NotContains(t, output, "none", "should not contain default commit")
	require.NotContains(t, output, "unknown", "should not contain default unknown values")
}

// Add this test to verify the mock provider works correctly
func TestMockVersionProvider(t *testing.T) {
	t.Parallel()

	mockProvider := &mockVersionProvider{
		info: providers.VersionInfo{
			Version:   "test-version",
			Commit:    "test-commit",
			BuiltAt:   "test-date",
			BuiltBy:   "test-user",
			GoVersion: "go1.21.0",
			OSArch:    "linux/amd64",
		},
	}

	info := mockProvider.GetVersionInfo()
	require.Equal(t, "test-version", info.Version)
	require.Equal(t, "test-commit", info.Commit)
	require.Equal(t, "test-date", info.BuiltAt)
	require.Equal(t, "test-user", info.BuiltBy)
	require.Equal(t, "go1.21.0", info.GoVersion)
	require.Equal(t, "linux/amd64", info.OSArch)
}

// Fix the NewVersionCommand test to handle nil service properly
func TestNewVersionCommand(t *testing.T) {
	t.Parallel()

	// Test with non-nil service
	customService := services.NewVersionService(nil)
	cmd1 := version.NewVersionCommand(customService)
	cmd2 := version.NewVersionCommand(customService)
	defaultCmd := version.GetCommand()

	// NewVersionCommand should create new instances
	require.NotSame(t, cmd1, cmd2)
	require.NotSame(t, cmd1, defaultCmd)

	// But should have same properties
	require.Equal(t, cmd1.Use, cmd2.Use)
	require.Equal(t, cmd1.Short, cmd2.Short)

	// Test execution works
	var buf bytes.Buffer
	cmd1.SetOut(&buf)
	cmd1.SetErr(&buf)
	cmd1.SetArgs([]string{})

	err := cmd1.Execute()
	require.NoError(t, err)
	require.NotEmpty(t, buf.String())
}

func TestVersionCommand_ContainsAllFields(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	cmd := newVersionCmdWithBuffer(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	require.NotEmpty(t, output)

	expectedFields := []string{
		"Version:",
		"Commit:",
		"Built at:",
		"Built by:",
		"Go version:",
		"OS/Arch:",
	}

	for _, field := range expectedFields {
		require.Contains(t, output, field, "output must contain field %q", field)
	}
}

func TestVersionCommand_DefaultValues(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	cmd := newVersionCmdWithBuffer(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()

	expected := map[string]string{
		"Version:":  "dev",
		"Commit:":   "none",
		"Built at:": "unknown",
		"Built by:": "unknown",
		// Go version and OS/Arch will be validated separately with regex
	}

	// Check for basic field presence and expected static values
	for field, expect := range expected {
		require.Contains(t, output, field, "missing field: %q", field)
		require.Contains(t, output, expect, "field %q should contain %q", field, expect)
	}

	// Validate dynamic fields with proper format checks
	require.Regexp(t, `Go version:\s+go\d+\.\d+(?:\.\d+)?`, output,
		"Go version should match format 'go1.21.0' or 'go1.21'")

	require.Regexp(t, `OS/Arch:\s+\w+/\w+`, output,
		"OS/Arch should match format 'linux/amd64'")

	// Optional: More specific OS/Arch validation
	require.Regexp(t, `OS/Arch:\s+(linux|darwin|windows|freebsd)/(amd64|arm64|386|arm)`, output,
		"OS/Arch should be a valid OS and architecture combination")
}

func TestVersionCommand_ValidFormat(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	cmd := newVersionCmdWithBuffer(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.GreaterOrEqual(t, len(lines), 6)

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		require.Contains(t, line, ":", "line %d should contain a colon: %q", i+1, line)
		parts := strings.SplitN(line, ":", 2)
		require.Len(t, parts, 2, "line %d should contain field and value", i+1)
		require.NotEmpty(t, strings.TrimSpace(parts[0]), "line %d has empty field name", i+1)
		require.NotEmpty(t, strings.TrimSpace(parts[1]), "line %d has empty field value", i+1)
	}
}

func TestVersionCommand_Performance(t *testing.T) {
	t.Parallel()

	// Test that multiple calls to GetCommand return the same cached instance
	commands := make([]*cobra.Command, 100)
	for i := 0; i < 100; i++ {
		commands[i] = version.GetCommand()
	}

	// All should be the same instance
	for i := 1; i < len(commands); i++ {
		require.Same(t, commands[0], commands[i])
	}
}

func TestVersionCommand_ThreadSafety(t *testing.T) {
	t.Parallel()

	// Test concurrent access to GetCommand
	done := make(chan *cobra.Command, 10)

	for i := 0; i < 10; i++ {
		go func() {
			done <- version.GetCommand()
		}()
	}

	var commands []*cobra.Command
	for i := 0; i < 10; i++ {
		commands = append(commands, <-done)
	}

	// All should be the same instance
	for i := 1; i < len(commands); i++ {
		require.Same(t, commands[0], commands[i])
	}
}

// Mock implementation for tests
type mockVersionProvider struct {
	info providers.VersionInfo
}

func (m *mockVersionProvider) GetVersionInfo() providers.VersionInfo {
	return m.info
}
