// internal/services/repository/rhel/handler_test.go
package rhel

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
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

func TestNewHandler(t *testing.T) {
	client := &MockSSHClient{}
	handler := NewHandler(client)

	assert.NotNil(t, handler)
	assert.Equal(t, client, handler.client)
	assert.NotNil(t, handler.sudo)
}

func TestHandler_Setup_Success_NoSudoNeeded(t *testing.T) {
	client := &MockSSHClient{}
	
	// Mock system directory write test - first one succeeds (no sudo needed)
	client.On("Execute", mock.Anything, "test -w /etc/apt/sources.list.d/").Return(nil)
	
	repoContent := `[superviz]
name=Superviz.io Repository
baseurl=https://repo.superviz.io/rpm/
enabled=1
gpgcheck=1
gpgkey=https://repo.superviz.io/rpm/RPM-GPG-KEY-superviz`

	// Mock repository setup commands without sudo
	expectedCommands := []string{
		"cat > /tmp/superviz.repo << 'EOF'\n" + repoContent + "\nEOF",
		"cp /tmp/superviz.repo /etc/yum.repos.d/superviz.repo",
		"rm /tmp/superviz.repo",
		"rpm --import https://repo.superviz.io/rpm/RPM-GPG-KEY-superviz",
		"if command -v dnf >/dev/null 2>&1; then dnf clean all; elif command -v yum >/dev/null 2>&1; then yum clean all; fi",
	}
	
	for _, cmd := range expectedCommands {
		client.On("Execute", mock.Anything, cmd).Return(nil)
	}

	handler := NewHandler(client)
	var output bytes.Buffer

	err := handler.Setup(context.Background(), &output)

	assert.NoError(t, err)
	assert.Contains(t, output.String(), "Setting up YUM/DNF repository...")
	assert.NotContains(t, output.String(), "Using sudo for system operations...")
	client.AssertExpectations(t)
}

func TestHandler_Setup_Success_WithSudo(t *testing.T) {
	client := &MockSSHClient{}
	
	// Mock system directory write tests - all fail (need sudo)
	client.On("Execute", mock.Anything, "test -w /etc/apt/sources.list.d/").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/apk/repositories").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/yum.repos.d/").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/pacman.conf").Return(errors.New("not writable"))
	
	// Mock sudo check - sudo available
	client.On("Execute", mock.Anything, "command -v sudo >/dev/null 2>&1").Return(nil)
	
	repoContent := `[superviz]
name=Superviz.io Repository
baseurl=https://repo.superviz.io/rpm/
enabled=1
gpgcheck=1
gpgkey=https://repo.superviz.io/rpm/RPM-GPG-KEY-superviz`

	// Mock repository setup commands with sudo prefix
	expectedCommands := []string{
		"cat > /tmp/superviz.repo << 'EOF'\n" + repoContent + "\nEOF",
		"sudo cp /tmp/superviz.repo /etc/yum.repos.d/superviz.repo",
		"rm /tmp/superviz.repo",
		"sudo rpm --import https://repo.superviz.io/rpm/RPM-GPG-KEY-superviz",
		"if command -v dnf >/dev/null 2>&1; then dnf clean all; elif command -v yum >/dev/null 2>&1; then yum clean all; fi",
	}
	
	for _, cmd := range expectedCommands {
		client.On("Execute", mock.Anything, cmd).Return(nil)
	}

	handler := NewHandler(client)
	var output bytes.Buffer

	err := handler.Setup(context.Background(), &output)

	assert.NoError(t, err)
	assert.Contains(t, output.String(), "Setting up YUM/DNF repository...")
	assert.Contains(t, output.String(), "Using sudo for system operations...")
	client.AssertExpectations(t)
}

func TestHandler_Setup_Success_SudoNotAvailable(t *testing.T) {
	client := &MockSSHClient{}
	
	// Mock system directory write tests - all fail (need sudo)
	client.On("Execute", mock.Anything, "test -w /etc/apt/sources.list.d/").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/apk/repositories").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/yum.repos.d/").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/pacman.conf").Return(errors.New("not writable"))
	
	// Mock sudo check - sudo not found
	client.On("Execute", mock.Anything, "command -v sudo >/dev/null 2>&1").Return(errors.New("sudo not found"))
	
	handler := NewHandler(client)
	var output bytes.Buffer

	err := handler.Setup(context.Background(), &output)

	// This should fail because we need sudo but it's not available
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "root privileges required but sudo is not available")
	assert.Contains(t, output.String(), "Setting up YUM/DNF repository...")
}

func TestHandler_Setup_SudoDetectionError(t *testing.T) {
	client := &MockSSHClient{}
	
	// Mock all Execute calls to return connection error
	client.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("connection failed"))

	handler := NewHandler(client)
	var output bytes.Buffer

	err := handler.Setup(context.Background(), &output)

	// Should get connection error during the write test or sudo check
	assert.Error(t, err)
	assert.Contains(t, output.String(), "Setting up YUM/DNF repository...")
}

func TestHandler_Setup_WriteError(t *testing.T) {
	client := &MockSSHClient{}
	handler := NewHandler(client)

	// Use a writer that will fail
	writer := &failingWriter{}

	err := handler.Setup(context.Background(), writer)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write to output")
}

func TestHandler_Setup_CommandExecutionError(t *testing.T) {
	client := &MockSSHClient{}
	
	// Mock system directory write test - first one succeeds (no sudo needed)
	client.On("Execute", mock.Anything, "test -w /etc/apt/sources.list.d/").Return(nil)
	
	repoContent := `[superviz]
name=Superviz.io Repository
baseurl=https://repo.superviz.io/rpm/
enabled=1
gpgcheck=1
gpgkey=https://repo.superviz.io/rpm/RPM-GPG-KEY-superviz`

	// Mock first command to fail
	client.On("Execute", mock.Anything, "cat > /tmp/superviz.repo << 'EOF'\n" + repoContent + "\nEOF").Return(errors.New("command failed"))

	handler := NewHandler(client)
	var output bytes.Buffer

	err := handler.Setup(context.Background(), &output)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command failed")
	client.AssertExpectations(t)
}

func TestHandler_Setup_SudoWriteError(t *testing.T) {
	client := &MockSSHClient{}
	
	// Mock system directory write tests - all fail (need sudo)
	client.On("Execute", mock.Anything, "test -w /etc/apt/sources.list.d/").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/apk/repositories").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/yum.repos.d/").Return(errors.New("not writable"))
	client.On("Execute", mock.Anything, "test -w /etc/pacman.conf").Return(errors.New("not writable"))
	
	// Mock sudo check - sudo available
	client.On("Execute", mock.Anything, "command -v sudo >/dev/null 2>&1").Return(nil)

	handler := NewHandler(client)

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
