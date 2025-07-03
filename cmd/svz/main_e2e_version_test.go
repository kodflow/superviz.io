//go:build e2e

package main

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestE2E_VersionCommand_AllFormats tests all output formats for the version command
func TestE2E_VersionCommand_AllFormats(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping e2e tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	binaryPath := getSvzBinaryPath()

	tests := []struct {
		name         string
		args         []string
		expectOutput []string
		expectJSON   bool
		expectYAML   bool
	}{
		{
			name:         "version_default_format",
			args:         []string{"version"},
			expectOutput: []string{"Version:", "Commit:", "Built at:", "Built by:", "OS/Arch:"},
		},
		{
			name:         "version_explicit_default_format",
			args:         []string{"version", "--format", "default"},
			expectOutput: []string{"Version:", "Commit:", "Built at:", "Built by:", "OS/Arch:"},
		},
		{
			name:         "version_short_flag",
			args:         []string{"version", "-f", "default"},
			expectOutput: []string{"Version:", "Commit:", "Built at:", "Built by:", "OS/Arch:"},
		},
		{
			name:         "version_json_format",
			args:         []string{"version", "--format", "json"},
			expectOutput: []string{"{", "}", "\"version\":", "\"commit\":", "\"built_at\":", "\"built_by\":", "\"go_version\":", "\"os_arch\":"},
			expectJSON:   true,
		},
		{
			name:         "version_json_short_flag",
			args:         []string{"version", "-f", "json"},
			expectOutput: []string{"{", "}", "\"version\":", "\"commit\":", "\"built_at\":", "\"built_by\":", "\"go_version\":", "\"os_arch\":"},
			expectJSON:   true,
		},
		{
			name:         "version_yaml_format",
			args:         []string{"version", "--format", "yaml"},
			expectOutput: []string{"version:", "commit:", "built_at:", "built_by:", "go_version:", "os_arch:"},
			expectYAML:   true,
		},
		{
			name:         "version_yaml_short_flag",
			args:         []string{"version", "-f", "yaml"},
			expectOutput: []string{"version:", "commit:", "built_at:", "built_by:", "go_version:", "os_arch:"},
			expectYAML:   true,
		},
		{
			name:         "version_short_format",
			args:         []string{"version", "--format", "short"},
			expectOutput: []string{},
		},
		{
			name:         "version_short_format_short_flag",
			args:         []string{"version", "-f", "short"},
			expectOutput: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.CommandContext(ctx, binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "Command failed: %s", string(output))

			outputStr := string(output)

			// Check expected output strings
			for _, expected := range tt.expectOutput {
				assert.Contains(t, outputStr, expected, "Expected output to contain: %s", expected)
			}

			// Validate JSON format if expected
			if tt.expectJSON {
				var jsonOutput map[string]interface{}
				err := json.Unmarshal(output, &jsonOutput)
				assert.NoError(t, err, "Output should be valid JSON")

				// Check that all expected fields are present
				expectedFields := []string{"version", "commit", "built_at", "built_by", "go_version", "os_arch"}
				for _, field := range expectedFields {
					assert.Contains(t, jsonOutput, field, "JSON output should contain field: %s", field)
				}
			}

			// Validate YAML format if expected
			if tt.expectYAML {
				var yamlOutput map[string]interface{}
				err := yaml.Unmarshal(output, &yamlOutput)
				assert.NoError(t, err, "Output should be valid YAML")

				// Check that all expected fields are present
				expectedFields := []string{"version", "commit", "built_at", "built_by", "go_version", "os_arch"}
				for _, field := range expectedFields {
					assert.Contains(t, yamlOutput, field, "YAML output should contain field: %s", field)
				}
			}

			// For short format, just verify it's a single line with version info
			if strings.Contains(strings.Join(tt.args, " "), "short") {
				lines := strings.Split(strings.TrimSpace(outputStr), "\n")
				assert.Len(t, lines, 1, "Short format should output exactly one line")
				assert.NotEmpty(t, lines[0], "Short format should not be empty")
			}
		})
	}
}

// TestE2E_VersionCommand_ErrorCases tests error scenarios for the version command
func TestE2E_VersionCommand_ErrorCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	binaryPath := getSvzBinaryPath()

	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "version_invalid_format",
			args:        []string{"version", "--format", "invalid"},
			expectError: true,
		},
		{
			name:        "version_invalid_short_format",
			args:        []string{"version", "-f", "invalid"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.CommandContext(ctx, binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()

			if tt.expectError {
				assert.Error(t, err, "Command should have failed")
				outputStr := string(output)
				assert.Contains(t, outputStr, "Error:", "Error output should contain 'Error:'")
			} else {
				assert.NoError(t, err, "Command should have succeeded")
			}
		})
	}
}

// TestE2E_VersionCommand_OutputValidation tests detailed output validation
func TestE2E_VersionCommand_OutputValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	binaryPath := getSvzBinaryPath()

	t.Run("version_default_output_structure", func(t *testing.T) {
		cmd := exec.CommandContext(ctx, binaryPath, "version")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Command failed")

		outputStr := string(output)
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")

		// Verify structure
		assert.GreaterOrEqual(t, len(lines), 5, "Should have at least 5 lines of output")

		// Verify each line contains expected format
		for _, line := range lines {
			assert.Contains(t, line, ":", "Each line should contain a colon separator")
		}
	})

	t.Run("version_json_structure", func(t *testing.T) {
		cmd := exec.CommandContext(ctx, binaryPath, "version", "--format", "json")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Command failed")

		var jsonOutput map[string]interface{}
		err = json.Unmarshal(output, &jsonOutput)
		require.NoError(t, err, "Output should be valid JSON")

		// Verify all expected fields exist and are strings
		expectedFields := []string{"version", "commit", "built_at", "built_by", "go_version", "os_arch"}
		for _, field := range expectedFields {
			value, exists := jsonOutput[field]
			assert.True(t, exists, "Field %s should exist", field)
			assert.IsType(t, "", value, "Field %s should be a string", field)
		}
	})

	t.Run("version_yaml_structure", func(t *testing.T) {
		cmd := exec.CommandContext(ctx, binaryPath, "version", "--format", "yaml")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Command failed")

		var yamlOutput map[string]interface{}
		err = yaml.Unmarshal(output, &yamlOutput)
		require.NoError(t, err, "Output should be valid YAML")

		// Verify all expected fields exist
		expectedFields := []string{"version", "commit", "built_at", "built_by", "go_version", "os_arch"}
		for _, field := range expectedFields {
			_, exists := yamlOutput[field]
			assert.True(t, exists, "Field %s should exist", field)
		}
	})
}

// TestE2E_VersionCommand_ConcurrentExecution tests that version command works with concurrent execution
func TestE2E_VersionCommand_ConcurrentExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	binaryPath := getSvzBinaryPath()

	t.Run("concurrent_version_calls", func(t *testing.T) {
		const numConcurrentCalls = 10
		results := make(chan error, numConcurrentCalls)

		for i := 0; i < numConcurrentCalls; i++ {
			go func() {
				cmd := exec.CommandContext(ctx, binaryPath, "version")
				output, err := cmd.CombinedOutput()
				if err != nil {
					results <- err
					return
				}

				outputStr := string(output)
				if !strings.Contains(outputStr, "Version:") {
					results <- assert.AnError
					return
				}

				results <- nil
			}()
		}

		// Collect all results
		for i := 0; i < numConcurrentCalls; i++ {
			select {
			case err := <-results:
				assert.NoError(t, err, "Concurrent execution should not fail")
			case <-ctx.Done():
				t.Fatal("Test timeout")
			}
		}
	})
}
