package common

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
)

// Mock SSH client for testing
type mockSSHClient struct {
	mock.Mock
}

func (m *mockSSHClient) Connect(ctx context.Context, config *ssh.Config) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *mockSSHClient) Execute(ctx context.Context, command string) error {
	args := m.Called(ctx, command)
	return args.Error(0)
}

func (m *mockSSHClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Tests for SudoHelper

func TestNewSudoHelper(t *testing.T) {
	client := &mockSSHClient{}
	helper := NewSudoHelper(client)

	require.NotNil(t, helper)
	assert.Equal(t, client, helper.client)
}

func TestSudoHelper_IsNeeded_NoSudoNeeded(t *testing.T) {
	client := &mockSSHClient{}

	// Mock the first check to succeed (directory is writable)
	client.On("Execute", mock.Anything, "test -w /etc/apt/sources.list.d/").Return(nil).Once()

	helper := NewSudoHelper(client)
	needSudo, err := helper.IsNeeded(context.Background())

	assert.NoError(t, err)
	assert.False(t, needSudo)
	client.AssertExpectations(t)
}

func TestSudoHelper_IsNeeded_SudoNeeded(t *testing.T) {
	client := &mockSSHClient{}

	// Mock all directory checks to fail
	client.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("permission denied")).Times(4)

	// Mock sudo availability check to succeed
	client.On("Execute", mock.Anything, "command -v sudo >/dev/null 2>&1").Return(nil)

	helper := NewSudoHelper(client)
	needSudo, err := helper.IsNeeded(context.Background())

	assert.NoError(t, err)
	assert.True(t, needSudo)
	client.AssertExpectations(t)
}

func TestSudoHelper_IsNeeded_SudoNotAvailable(t *testing.T) {
	client := &mockSSHClient{}

	// Mock all directory checks to fail
	client.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("permission denied")).Times(4)

	// Mock sudo availability check to fail
	client.On("Execute", mock.Anything, "command -v sudo >/dev/null 2>&1").Return(errors.New("command not found"))

	helper := NewSudoHelper(client)
	needSudo, err := helper.IsNeeded(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "root privileges required but sudo is not available")
	assert.False(t, needSudo)
	client.AssertExpectations(t)
}

func TestSudoHelper_AddPrefix_NoSudo(t *testing.T) {
	client := &mockSSHClient{}
	helper := NewSudoHelper(client)

	commands := []string{
		"apt update",
		"curl -fsSL https://example.com",
		"echo hello",
	}

	result := helper.AddPrefix(commands, false)

	// Should return original commands unchanged
	assert.Equal(t, commands, result)
}

func TestSudoHelper_AddPrefix_WithSudo(t *testing.T) {
	client := &mockSSHClient{}
	helper := NewSudoHelper(client)

	commands := []string{
		"apt update",                     // Should get sudo prefix
		"curl -fsSL https://example.com", // Should not get sudo prefix
		"cp /tmp/file /etc/",             // Should get sudo prefix (writes to /etc/)
		"echo hello",                     // Should not get sudo prefix
	}

	result := helper.AddPrefix(commands, true)

	expected := []string{
		"sudo apt update",
		"curl -fsSL https://example.com",
		"sudo cp /tmp/file /etc/",
		"echo hello",
	}

	assert.Equal(t, expected, result)
}

func TestSudoHelper_CommandNeedsSudo(t *testing.T) {
	client := &mockSSHClient{}
	helper := NewSudoHelper(client)

	tests := []struct {
		command  string
		expected bool
	}{
		// APT commands
		{"apt update", true},
		{"apt install package", true},
		{"apt-get update", true},

		// APK commands
		{"apk update", true},
		{"apk add package", true},

		// YUM/DNF commands
		{"yum install package", true},
		{"dnf install package", true},

		// Pacman commands
		{"pacman -S package", true},
		{"pacman-key --add", true},

		// System path operations
		{"cp /tmp/file /etc/config", true},
		{"mv /tmp/file /usr/share/file", true},

		// Non-root commands
		{"curl -fsSL https://example.com", false},
		{"echo hello", false},
		{"cat /tmp/file", false},
		{"ls -la", false},
		{"echo 'content' > /usr/share/file", true}, // Redirection detected due to /usr/ path
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := helper.commandNeedsSudo(tt.command)
			assert.Equal(t, tt.expected, result, "Command: %s", tt.command)
		})
	}
}

// Tests for CommandExecutor

func TestNewCommandExecutor(t *testing.T) {
	client := &mockSSHClient{}
	executor := NewCommandExecutor(client)

	require.NotNil(t, executor)
	assert.Equal(t, client, executor.client)
}

func TestCommandExecutor_Execute_Success(t *testing.T) {
	client := &mockSSHClient{}

	commands := []string{
		"echo hello",
		"ls -la",
		"pwd",
	}

	// Mock all commands to succeed
	for _, cmd := range commands {
		client.On("Execute", mock.Anything, cmd).Return(nil).Once()
	}

	executor := NewCommandExecutor(client)
	var output MockWriter

	err := executor.Execute(context.Background(), commands, &output)

	assert.NoError(t, err)

	// Verify output contains progress information
	outputStr := output.String()
	assert.Contains(t, outputStr, "[1/3] echo hello")
	assert.Contains(t, outputStr, "[2/3] ls -la")
	assert.Contains(t, outputStr, "[3/3] pwd")

	client.AssertExpectations(t)
}

func TestCommandExecutor_Execute_CommandFails(t *testing.T) {
	client := &mockSSHClient{}

	commands := []string{
		"echo hello",
		"failing-command",
		"pwd",
	}

	// Mock first command to succeed, second to fail
	client.On("Execute", mock.Anything, "echo hello").Return(nil).Once()
	client.On("Execute", mock.Anything, "failing-command").Return(errors.New("command failed")).Once()

	executor := NewCommandExecutor(client)
	var output MockWriter

	err := executor.Execute(context.Background(), commands, &output)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command failed: failing-command")
	assert.Contains(t, err.Error(), "command failed")

	client.AssertExpectations(t)
}

func TestCommandExecutor_Execute_WriteError(t *testing.T) {
	client := &mockSSHClient{}

	commands := []string{"echo hello"}

	executor := NewCommandExecutor(client)
	writer := &FailingWriter{}

	err := executor.Execute(context.Background(), commands, writer)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write to output")
}

// Tests for BaseHandler

func TestNewBaseHandler(t *testing.T) {
	client := &mockSSHClient{}
	handler := NewBaseHandler(client)

	require.NotNil(t, handler)
	assert.Equal(t, client, handler.client)
	assert.NotNil(t, handler.sudo)
}

func TestBaseHandler_ExecuteSetup_Success(t *testing.T) {
	client := &mockSSHClient{}

	// Mock sudo detection
	client.On("Execute", mock.Anything, "test -w /etc/apt/sources.list.d/").Return(nil).Once()

	// Mock command execution
	client.On("Execute", mock.Anything, "command1").Return(nil).Once()
	client.On("Execute", mock.Anything, "command2").Return(nil).Once()

	handler := NewBaseHandler(client)

	// Use a mock writer to capture output
	var output strings.Builder
	commands := []string{"command1", "command2"}

	err := handler.ExecuteSetup(context.Background(), &output, "Setting up test...", commands)

	assert.NoError(t, err)
	assert.Contains(t, output.String(), "Setting up test...")
	client.AssertExpectations(t)
}

func TestBaseHandler_ExecuteSetup_Basic(t *testing.T) {
	client := &mockSSHClient{}

	// Simple success case
	client.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(nil).Times(3) // sudo check + 2 commands

	handler := NewBaseHandler(client)

	var output strings.Builder
	commands := []string{"echo hello", "echo world"}

	err := handler.ExecuteSetup(context.Background(), &output, "Setting up...", commands)

	assert.NoError(t, err)
	assert.Contains(t, output.String(), "Setting up...")
	client.AssertExpectations(t)
}

func TestBaseHandler_ExecuteSetup_CommandExecutionError(t *testing.T) {
	client := &mockSSHClient{}

	// Mock sudo detection success then command failure
	client.On("Execute", mock.Anything, "test -w /etc/apt/sources.list.d/").Return(nil).Once()
	client.On("Execute", mock.Anything, "failing-command").Return(errors.New("command failed")).Once()

	handler := NewBaseHandler(client)

	var output strings.Builder
	commands := []string{"failing-command"}

	err := handler.ExecuteSetup(context.Background(), &output, "Test setup", commands)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command failed")
	client.AssertExpectations(t)
}

// Test helpers

// MockWriter implements io.Writer for testing
type MockWriter struct {
	data []byte
}

func (m *MockWriter) Write(p []byte) (n int, err error) {
	m.data = append(m.data, p...)
	return len(p), nil
}

func (m *MockWriter) String() string {
	return string(m.data)
}

// FailingWriter always returns an error when writing
type FailingWriter struct{}

func (f *FailingWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write failed")
}
