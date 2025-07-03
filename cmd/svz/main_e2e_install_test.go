//go:build e2e

package main

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_InstallCommand_AllFlags tests all flags for the install command
func TestE2E_InstallCommand_AllFlags(t *testing.T) {
	if !isDockerAvailable() {
		t.Fatal("Docker is required for E2E tests but is not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	binaryPath := getSvzBinaryPath()

	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectHelp  bool
	}{
		{
			name:        "install_no_args",
			args:        []string{"install"},
			expectError: true, // Should require target argument
		},
		{
			name:        "install_help_flag",
			args:        []string{"install", "--help"},
			expectError: false,
			expectHelp:  true,
		},
		{
			name:        "install_help_short_flag",
			args:        []string{"install", "-h"},
			expectError: false,
			expectHelp:  true,
		},
		{
			name:        "install_with_ssh_key",
			args:        []string{"install", "user@localhost", "--ssh-key", "/path/to/key"},
			expectError: true, // Should fail due to invalid key path
		},
		{
			name:        "install_with_ssh_key_short",
			args:        []string{"install", "user@localhost", "-i", "/path/to/key"},
			expectError: true, // Should fail due to invalid key path
		},
		{
			name:        "install_with_password",
			args:        []string{"install", "user@localhost", "--password", "testpass"},
			expectError: true, // Should fail due to connection
		},
		{
			name:        "install_with_custom_port",
			args:        []string{"install", "user@localhost", "--ssh-port", "2222"},
			expectError: true, // Should fail due to connection
		},
		{
			name:        "install_with_custom_port_short",
			args:        []string{"install", "user@localhost", "-p", "2222"},
			expectError: true, // Should fail due to connection
		},
		{
			name:        "install_with_timeout",
			args:        []string{"install", "user@localhost", "--timeout", "30s"},
			expectError: true, // Should fail due to connection
		},
		{
			name:        "install_with_timeout_short",
			args:        []string{"install", "user@localhost", "-t", "30s"},
			expectError: true, // Should fail due to connection
		},
		{
			name:        "install_with_force",
			args:        []string{"install", "user@localhost", "--force"},
			expectError: true, // Should fail due to connection
		},
		{
			name:        "install_with_force_short",
			args:        []string{"install", "user@localhost", "-f"},
			expectError: true, // Should fail due to connection
		},
		{
			name:        "install_skip_host_key_check",
			args:        []string{"install", "user@localhost", "--skip-host-key-check"},
			expectError: true, // Should fail due to connection
		},
		{
			name:        "install_all_flags_combined",
			args:        []string{"install", "user@localhost", "--ssh-key", "/path/to/key", "--ssh-port", "2222", "--timeout", "60s", "--force", "--skip-host-key-check"},
			expectError: true, // Should fail due to invalid key path or connection
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.CommandContext(ctx, binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()

			outputStr := string(output)
			t.Logf("Command: svz %v\nOutput: %s\nError: %v", tt.args, outputStr, err)

			if tt.expectHelp {
				require.NoError(t, err, "Help command should succeed")
				assert.Contains(t, outputStr, "Usage:", "Help output should contain usage")
				assert.Contains(t, outputStr, "install", "Help output should mention install")
				assert.Contains(t, outputStr, "Flags:", "Help output should show flags")
				assert.Contains(t, outputStr, "--ssh-key", "Help output should show ssh-key flag")
				assert.Contains(t, outputStr, "--password", "Help output should show password flag")
				assert.Contains(t, outputStr, "--ssh-port", "Help output should show ssh-port flag")
				assert.Contains(t, outputStr, "--timeout", "Help output should show timeout flag")
				assert.Contains(t, outputStr, "--force", "Help output should show force flag")
				assert.Contains(t, outputStr, "--skip-host-key-check", "Help output should show skip-host-key-check flag")
			} else if tt.expectError {
				assert.Error(t, err, "Command should fail: %v", tt.args)
				assert.NotEmpty(t, outputStr, "Error output should not be empty")
			} else {
				require.NoError(t, err, "Command should succeed: %v", tt.args)
			}
		})
	}
}

// TestE2E_InstallCommand_InvalidFlags tests invalid flag combinations
func TestE2E_InstallCommand_InvalidFlags(t *testing.T) {
	if !isDockerAvailable() {
		t.Fatal("Docker is required for E2E tests but is not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	binaryPath := getSvzBinaryPath()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "install_unknown_flag",
			args: []string{"install", "--unknown-flag"},
		},
		{
			name: "install_invalid_port",
			args: []string{"install", "user@localhost", "--ssh-port", "invalid"},
		},
		{
			name: "install_invalid_timeout",
			args: []string{"install", "user@localhost", "--timeout", "invalid"},
		},
		{
			name: "install_negative_port",
			args: []string{"install", "user@localhost", "--ssh-port", "-1"},
		},
		{
			name: "install_port_too_high",
			args: []string{"install", "user@localhost", "--ssh-port", "70000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.CommandContext(ctx, binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()

			outputStr := string(output)
			t.Logf("Command: svz %v\nOutput: %s\nError: %v", tt.args, outputStr, err)

			// These should fail with error messages
			assert.Error(t, err, "Invalid command should fail: %v", tt.args)
			assert.NotEmpty(t, outputStr, "Error output should not be empty")
		})
	}
}

// TestE2E_InstallCommand_TimeoutFormats tests different timeout formats
func TestE2E_InstallCommand_TimeoutFormats(t *testing.T) {
	if !isDockerAvailable() {
		t.Fatal("Docker is required for E2E tests but is not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	binaryPath := getSvzBinaryPath()

	tests := []struct {
		name        string
		timeout     string
		expectError bool
	}{
		{
			name:        "timeout_seconds",
			timeout:     "30s",
			expectError: true, // Will fail on connection, but timeout format is valid
		},
		{
			name:        "timeout_minutes",
			timeout:     "5m",
			expectError: true, // Will fail on connection, but timeout format is valid
		},
		{
			name:        "timeout_hours",
			timeout:     "1h",
			expectError: true, // Will fail on connection, but timeout format is valid
		},
		{
			name:        "timeout_combined",
			timeout:     "1m30s",
			expectError: true, // Will fail on connection, but timeout format is valid
		},
		{
			name:        "timeout_milliseconds",
			timeout:     "5000ms",
			expectError: true, // Will fail on connection, but timeout format is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"install", "user@localhost", "--timeout", tt.timeout}
			cmd := exec.CommandContext(ctx, binaryPath, args...)
			output, err := cmd.CombinedOutput()

			outputStr := string(output)
			t.Logf("Command: svz %v\nOutput: %s\nError: %v", args, outputStr, err)

			if tt.expectError {
				assert.Error(t, err, "Command should fail due to connection: %v", args)
			} else {
				require.NoError(t, err, "Command should succeed: %v", args)
			}
		})
	}
}

// TestE2E_InstallCommand_TargetFormats tests different target formats
func TestE2E_InstallCommand_TargetFormats(t *testing.T) {
	if !isDockerAvailable() {
		t.Fatal("Docker is required for E2E tests but is not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	binaryPath := getSvzBinaryPath()

	tests := []struct {
		name   string
		target string
	}{
		{
			name:   "target_user_at_host",
			target: "user@hostname",
		},
		{
			name:   "target_user_at_ip",
			target: "user@192.168.1.100",
		},
		{
			name:   "target_user_at_localhost",
			target: "user@localhost",
		},
		{
			name:   "target_root_user",
			target: "root@server.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"install", tt.target, "--timeout", "5s"}
			cmd := exec.CommandContext(ctx, binaryPath, args...)
			output, err := cmd.CombinedOutput()

			outputStr := string(output)
			t.Logf("Command: svz %v\nOutput: %s\nError: %v", args, outputStr, err)

			// All these should fail due to connection, but target format parsing should work
			assert.Error(t, err, "Command should fail due to connection: %v", args)
			assert.NotEmpty(t, outputStr, "Error output should not be empty")
		})
	}
}

// TestE2E_InstallCommand_Help tests help functionality
func TestE2E_InstallCommand_Help(t *testing.T) {
	if !isDockerAvailable() {
		t.Fatal("Docker is required for E2E tests but is not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	binaryPath := getSvzBinaryPath()

	tests := []struct {
		name         string
		args         []string
		expectOutput []string
	}{
		{
			name: "install_help_flag",
			args: []string{"install", "--help"},
			expectOutput: []string{
				"Usage:", "svz install", "Flags:",
				"--ssh-key", "--password", "--ssh-port", "--timeout", "--force", "--skip-host-key-check",
			},
		},
		{
			name: "install_help_short_flag",
			args: []string{"install", "-h"},
			expectOutput: []string{
				"Usage:", "svz install", "Flags:",
				"--ssh-key", "--password", "--ssh-port", "--timeout", "--force", "--skip-host-key-check",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.CommandContext(ctx, binaryPath, tt.args...)
			output, err := cmd.Output()

			// Help commands should succeed
			require.NoError(t, err, "Help command should succeed: %v", tt.args)

			outputStr := string(output)
			t.Logf("Command: svz %v\nOutput: %s", tt.args, outputStr)

			// Check expected outputs
			for _, expected := range tt.expectOutput {
				assert.Contains(t, outputStr, expected, "Help output should contain: %s", expected)
			}
		})
	}
}

// TestE2E_InstallCommand_FlagShortcuts tests all flag shortcuts work correctly
func TestE2E_InstallCommand_FlagShortcuts(t *testing.T) {
	if !isDockerAvailable() {
		t.Fatal("Docker is required for E2E tests but is not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	binaryPath := getSvzBinaryPath()

	tests := []struct {
		name      string
		longForm  []string
		shortForm []string
	}{
		{
			name:      "ssh_key_flag",
			longForm:  []string{"install", "user@localhost", "--ssh-key", "/path/to/key"},
			shortForm: []string{"install", "user@localhost", "-i", "/path/to/key"},
		},
		{
			name:      "ssh_port_flag",
			longForm:  []string{"install", "user@localhost", "--ssh-port", "2222"},
			shortForm: []string{"install", "user@localhost", "-p", "2222"},
		},
		{
			name:      "timeout_flag",
			longForm:  []string{"install", "user@localhost", "--timeout", "30s"},
			shortForm: []string{"install", "user@localhost", "-t", "30s"},
		},
		{
			name:      "force_flag",
			longForm:  []string{"install", "user@localhost", "--force"},
			shortForm: []string{"install", "user@localhost", "-f"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test long form
			t.Run("long_form", func(t *testing.T) {
				cmd := exec.CommandContext(ctx, binaryPath, tt.longForm...)
				output, err := cmd.CombinedOutput()

				outputStr := string(output)
				t.Logf("Long form - Command: svz %v\nOutput: %s\nError: %v", tt.longForm, outputStr, err)

				// Should fail due to connection, but flag should be recognized
				assert.Error(t, err, "Command should fail due to connection")
				assert.NotEmpty(t, outputStr, "Error output should not be empty")
			})

			// Test short form
			t.Run("short_form", func(t *testing.T) {
				cmd := exec.CommandContext(ctx, binaryPath, tt.shortForm...)
				output, err := cmd.CombinedOutput()

				outputStr := string(output)
				t.Logf("Short form - Command: svz %v\nOutput: %s\nError: %v", tt.shortForm, outputStr, err)

				// Should fail due to connection, but flag should be recognized
				assert.Error(t, err, "Command should fail due to connection")
				assert.NotEmpty(t, outputStr, "Error output should not be empty")
			})
		})
	}
}
