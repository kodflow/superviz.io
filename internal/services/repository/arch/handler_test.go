// internal/services/repository/arch/handler_test.go
package arch

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
	"github.com/kodflow/superviz.io/internal/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSSHClient mocks the SSH client interface
type MockSSHClient struct {
	mock.Mock
}

func (m *MockSSHClient) Connect(ctx context.Context, config *ssh.Config) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockSSHClient) Execute(ctx context.Context, command string) error {
	args := m.Called(ctx, command)
	return args.Error(0)
}

func (m *MockSSHClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockInstallProvider mocks the install provider interface
type MockInstallProvider struct {
	mock.Mock
}

func (m *MockInstallProvider) Install(ctx context.Context, packageName string) error {
	args := m.Called(ctx, packageName)
	return args.Error(0)
}

func (m *MockInstallProvider) GetInstallInfo() providers.InstallInfo {
	args := m.Called()
	return args.Get(0).(providers.InstallInfo)
}

func (m *MockInstallProvider) GetRepositoryURL() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockInstallProvider) GetPackageName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockInstallProvider) GetGPGKeyID() string {
	args := m.Called()
	return args.String(0)
}

func TestNewHandler(t *testing.T) {
	client := &MockSSHClient{}
	provider := &MockInstallProvider{}
	handler := NewHandler(client, provider)

	assert.NotNil(t, handler)
	assert.Equal(t, client, handler.client)
	assert.Equal(t, provider, handler.provider)
	assert.NotNil(t, handler.sudo)
}

func TestHandler_Setup_Success_NoSudoNeeded(t *testing.T) {
	client := &MockSSHClient{}
	provider := &MockInstallProvider{}
	
	// Mock system directory write test - first one succeeds (no sudo needed)
	client.On("Execute", mock.Anything, "test -w /etc/apt/sources.list.d/").Return(nil)
	
	// Mock provider to return a GPG key ID
	provider.On("GetGPGKeyID").Return("ABC123")
	
	// Mock repository setup commands without sudo
	expectedCommands := []string{
		"cat >> /tmp/superviz-pacman.conf << 'EOF'\n\n[superviz]\nServer = https://repo.superviz.io/arch/$arch\nEOF",
		"cat /tmp/superviz-pacman.conf >> /etc/pacman.conf",
		"rm /tmp/superviz-pacman.conf",
		"pacman-key --recv-keys ABC123",
		"pacman-key --lsign-key ABC123",
		"pacman -Sy",
	}
	
	for _, cmd := range expectedCommands {
		client.On("Execute", mock.Anything, cmd).Return(nil)
	}

	handler := NewHandler(client, provider)
	var output bytes.Buffer

	err := handler.Setup(context.Background(), &output)

	assert.NoError(t, err)
	assert.Contains(t, output.String(), "Setting up Pacman repository...")
	assert.NotContains(t, output.String(), "Using sudo for system operations...")
	client.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestHandler_Setup_Success_WithSudo(t *testing.T) {
	client := &MockSSHClient{}
	provider := &MockInstallProvider{}
	
	// Mock system directory write tests - all fail (need sudo)
	client.On("Execute", mock.Anything, "test -w /etc/apt/sources.list.d/").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/apk/repositories").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/yum.repos.d/").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/pacman.conf").Return(errors.New("not writable"))
	
	// Mock sudo check - sudo available
	client.On("Execute", mock.Anything, "command -v sudo >/dev/null 2>&1").Return(nil)
	
	// Mock provider to return a GPG key ID
	provider.On("GetGPGKeyID").Return("DEF456")
	
	// Mock repository setup commands with sudo prefix
	expectedCommands := []string{
		"cat >> /tmp/superviz-pacman.conf << 'EOF'\n\n[superviz]\nServer = https://repo.superviz.io/arch/$arch\nEOF",
		"sudo cat /tmp/superviz-pacman.conf >> /etc/pacman.conf",
		"rm /tmp/superviz-pacman.conf",
		"sudo pacman-key --recv-keys DEF456",
		"sudo pacman-key --lsign-key DEF456",
		"sudo pacman -Sy",
	}
	
	for _, cmd := range expectedCommands {
		client.On("Execute", mock.Anything, cmd).Return(nil)
	}

	handler := NewHandler(client, provider)
	var output bytes.Buffer

	err := handler.Setup(context.Background(), &output)

	assert.NoError(t, err)
	assert.Contains(t, output.String(), "Setting up Pacman repository...")
	assert.Contains(t, output.String(), "Using sudo for system operations...")
	client.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestHandler_Setup_Success_SudoNotAvailable(t *testing.T) {
	client := &MockSSHClient{}
	provider := &MockInstallProvider{}
	
	// Mock system directory write tests - all fail (need sudo)
	client.On("Execute", mock.Anything, "test -w /etc/apt/sources.list.d/").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/apk/repositories").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/yum.repos.d/").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/pacman.conf").Return(errors.New("not writable"))
	
	// Mock sudo check - sudo not found
	client.On("Execute", mock.Anything, "command -v sudo >/dev/null 2>&1").Return(errors.New("sudo not found"))
	
	handler := NewHandler(client, provider)
	var output bytes.Buffer

	err := handler.Setup(context.Background(), &output)

	// This should fail because we need sudo but it's not available
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "root privileges required but sudo is not available")
	assert.Contains(t, output.String(), "Setting up Pacman repository...")
}

func TestHandler_Setup_SudoDetectionError(t *testing.T) {
	client := &MockSSHClient{}
	provider := &MockInstallProvider{}
	
	// Mock all Execute calls to return connection error
	client.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("connection failed"))

	handler := NewHandler(client, provider)
	var output bytes.Buffer

	err := handler.Setup(context.Background(), &output)

	// Should get connection error during the write test or sudo check
	assert.Error(t, err)
	assert.Contains(t, output.String(), "Setting up Pacman repository...")
}

func TestHandler_Setup_WriteError(t *testing.T) {
	client := &MockSSHClient{}
	provider := &MockInstallProvider{}
	handler := NewHandler(client, provider)

	// Use a writer that will fail
	writer := &failingWriter{}

	err := handler.Setup(context.Background(), writer)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write to output")
}

func TestHandler_Setup_CommandExecutionError(t *testing.T) {
	client := &MockSSHClient{}
	provider := &MockInstallProvider{}
	
	// Mock system directory write test - first one succeeds (no sudo needed)
	client.On("Execute", mock.Anything, "test -w /etc/apt/sources.list.d/").Return(nil)
	
	// Mock provider to return a GPG key ID
	provider.On("GetGPGKeyID").Return("GHI789")
	
	// Mock first command to fail
	client.On("Execute", mock.Anything, "cat >> /tmp/superviz-pacman.conf << 'EOF'\n\n[superviz]\nServer = https://repo.superviz.io/arch/$arch\nEOF").Return(errors.New("command failed"))

	handler := NewHandler(client, provider)
	var output bytes.Buffer

	err := handler.Setup(context.Background(), &output)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command failed")
	client.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestHandler_Setup_SudoWriteError(t *testing.T) {
	client := &MockSSHClient{}
	provider := &MockInstallProvider{}
	
	// Mock system directory write tests - all fail (need sudo)
	client.On("Execute", mock.Anything, "test -w /etc/apt/sources.list.d/").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/apk/repositories").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/yum.repos.d/").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/pacman.conf").Return(errors.New("not writable"))
	
	// Mock sudo check - sudo available
	client.On("Execute", mock.Anything, "command -v sudo >/dev/null 2>&1").Return(nil)

	handler := NewHandler(client, provider)

	// Use a writer that fails on the second write (sudo message)
	writer := &conditionalFailingWriter{failOnSecond: true}

	err := handler.Setup(context.Background(), writer)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write to output")
}

// Helper types for testing

type failingWriter struct{}

func (w *failingWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write failed")
}

type conditionalFailingWriter struct {
	writeCount   int
	failOnSecond bool
}

func (w *conditionalFailingWriter) Write(p []byte) (n int, err error) {
	w.writeCount++
	if w.failOnSecond && w.writeCount == 2 {
		return 0, errors.New("write failed")
	}
	return len(p), nil
}
