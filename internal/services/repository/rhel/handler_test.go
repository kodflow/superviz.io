// internal/services/repository/rhel/handler_test.go
package rhel

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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
	assert.NotNil(t, handler.Base)
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
	client.On("Execute", mock.Anything, "cat > /tmp/superviz.repo << 'EOF'\n"+repoContent+"\nEOF").Return(errors.New("command failed"))

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

// Helper types for testing writers that fail
type failingWriter struct{}

func (w *failingWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("failed to write to output")
}

type conditionalFailingWriter struct {
	failOnSecond bool
	writeCount   int
}

func (w *conditionalFailingWriter) Write(p []byte) (n int, err error) {
	w.writeCount++
	if w.failOnSecond && w.writeCount >= 2 {
		return 0, errors.New("failed to write to output")
	}
	return len(p), nil
}

// Tests for validation functions
func TestValidateURL(t *testing.T) {
	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid HTTPS URL",
			url:       "https://repo.superviz.io/rpm/",
			expectErr: false,
		},
		{
			name:      "empty URL",
			url:       "",
			expectErr: true,
			errMsg:    "URL cannot be empty",
		},
		{
			name:      "whitespace only URL",
			url:       "   ",
			expectErr: true,
			errMsg:    "URL cannot be empty",
		},
		{
			name:      "HTTP URL (insecure)",
			url:       "http://repo.superviz.io/rpm/",
			expectErr: true,
			errMsg:    "URL must use HTTPS scheme",
		},
		{
			name:      "invalid URL format",
			url:       "not-a-url",
			expectErr: true,
			errMsg:    "URL must use HTTPS scheme",
		},
		{
			name:      "URL without host",
			url:       "https://",
			expectErr: true,
			errMsg:    "URL must have a valid host",
		},
		{
			name:      "FTP URL",
			url:       "ftp://repo.superviz.io/rpm/",
			expectErr: true,
			errMsg:    "URL must use HTTPS scheme",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateURL(tc.url)

			if tc.expectErr {
				assert.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRepoConfig(t *testing.T) {
	testCases := []struct {
		name      string
		config    *RepoConfig
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid config",
			config: &RepoConfig{
				Name:      "Test Repo",
				BaseURL:   "https://repo.example.com/rpm/",
				GPGKeyURL: "https://repo.example.com/gpg-key",
				Enabled:   true,
				GPGCheck:  true,
			},
			expectErr: false,
		},
		{
			name:      "nil config",
			config:    nil,
			expectErr: true,
			errMsg:    "repository configuration cannot be nil",
		},
		{
			name: "empty name",
			config: &RepoConfig{
				Name:      "",
				BaseURL:   "https://repo.example.com/rpm/",
				GPGKeyURL: "https://repo.example.com/gpg-key",
			},
			expectErr: true,
			errMsg:    "repository name cannot be empty",
		},
		{
			name: "whitespace only name",
			config: &RepoConfig{
				Name:      "   ",
				BaseURL:   "https://repo.example.com/rpm/",
				GPGKeyURL: "https://repo.example.com/gpg-key",
			},
			expectErr: true,
			errMsg:    "repository name cannot be empty",
		},
		{
			name: "invalid base URL",
			config: &RepoConfig{
				Name:      "Test Repo",
				BaseURL:   "http://insecure.example.com/",
				GPGKeyURL: "https://repo.example.com/gpg-key",
			},
			expectErr: true,
			errMsg:    "invalid base URL",
		},
		{
			name: "invalid GPG key URL",
			config: &RepoConfig{
				Name:      "Test Repo",
				BaseURL:   "https://repo.example.com/rpm/",
				GPGKeyURL: "ftp://insecure.example.com/gpg-key",
			},
			expectErr: true,
			errMsg:    "invalid GPG key URL",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRepoConfig(tc.config)

			if tc.expectErr {
				assert.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateRepoContent(t *testing.T) {
	config := &RepoConfig{
		Name:      "Test Repository",
		BaseURL:   "https://repo.example.com/rpm/",
		GPGKeyURL: "https://repo.example.com/gpg-key",
		Enabled:   true,
		GPGCheck:  true,
	}

	content, err := generateRepoContent(config)
	assert.NoError(t, err)
	assert.NotEmpty(t, content)

	// Check that all expected values are present
	assert.Contains(t, content, "[superviz]")
	assert.Contains(t, content, "name=Test Repository")
	assert.Contains(t, content, "baseurl=https://repo.example.com/rpm/")
	assert.Contains(t, content, "enabled=1")
	assert.Contains(t, content, "gpgcheck=1")
	assert.Contains(t, content, "gpgkey=https://repo.example.com/gpg-key")
}

func TestGenerateRepoContent_DisabledConfig(t *testing.T) {
	config := &RepoConfig{
		Name:      "Disabled Repository",
		BaseURL:   "https://repo.example.com/rpm/",
		GPGKeyURL: "https://repo.example.com/gpg-key",
		Enabled:   false,
		GPGCheck:  false,
	}

	content, err := generateRepoContent(config)
	assert.NoError(t, err)
	assert.NotEmpty(t, content)

	// Check that disabled flags are set correctly
	assert.Contains(t, content, "enabled=0")
	assert.Contains(t, content, "gpgcheck=0")
}

// Tests for provider functionality
func TestDefaultRepoProvider(t *testing.T) {
	provider := NewDefaultRepoProvider()
	config := provider.GetRepoConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "Superviz.io Repository", config.Name)
	assert.Equal(t, "https://repo.superviz.io/rpm/", config.BaseURL)
	assert.Equal(t, "https://repo.superviz.io/rpm/RPM-GPG-KEY-superviz", config.GPGKeyURL)
	assert.True(t, config.Enabled)
	assert.True(t, config.GPGCheck)

	// Validate that default config is valid
	err := validateRepoConfig(config)
	assert.NoError(t, err)
}

func TestCustomRepoProvider(t *testing.T) {
	provider := NewCustomRepoProvider(
		"Custom Test Repo",
		"https://custom.example.com/rpm/",
		"https://custom.example.com/gpg-key",
		false,
		false,
	)

	config := provider.GetRepoConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "Custom Test Repo", config.Name)
	assert.Equal(t, "https://custom.example.com/rpm/", config.BaseURL)
	assert.Equal(t, "https://custom.example.com/gpg-key", config.GPGKeyURL)
	assert.False(t, config.Enabled)
	assert.False(t, config.GPGCheck)
}

func TestNewHandlerWithProvider(t *testing.T) {
	client := &MockSSHClient{}
	provider := NewCustomRepoProvider(
		"Test Repo",
		"https://test.example.com/rpm/",
		"https://test.example.com/gpg-key",
		true,
		true,
	)

	handler := NewHandlerWithProvider(client, provider)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.Base)
	assert.Equal(t, provider, handler.provider)
}

// Test Setup with custom provider
func TestHandler_Setup_WithCustomProvider(t *testing.T) {
	client := &MockSSHClient{}
	provider := NewCustomRepoProvider(
		"Custom Test Repository",
		"https://custom.example.com/rpm/",
		"https://custom.example.com/gpg-key",
		true,
		true,
	)

	// Mock system directory write test - first one succeeds (no sudo needed)
	client.On("Execute", mock.Anything, "test -w /etc/apt/sources.list.d/").Return(nil)

	// Mock repository setup commands without sudo
	expectedRepoContent := `[superviz]
name=Custom Test Repository
baseurl=https://custom.example.com/rpm/
enabled=1
gpgcheck=1
gpgkey=https://custom.example.com/gpg-key`

	expectedCommands := []string{
		fmt.Sprintf("cat > /tmp/superviz.repo << 'EOF'\n%s\nEOF", expectedRepoContent),
		"cp /tmp/superviz.repo /etc/yum.repos.d/superviz.repo",
		"rm /tmp/superviz.repo",
		"rpm --import https://custom.example.com/gpg-key",
		"if command -v dnf >/dev/null 2>&1; then dnf clean all; elif command -v yum >/dev/null 2>&1; then yum clean all; fi",
	}

	for _, cmd := range expectedCommands {
		client.On("Execute", mock.Anything, cmd).Return(nil)
	}

	handler := NewHandlerWithProvider(client, provider)
	var output bytes.Buffer

	err := handler.Setup(context.Background(), &output)

	assert.NoError(t, err)
	assert.Contains(t, output.String(), "Setting up YUM/DNF repository...")
	client.AssertExpectations(t)
}

// Test Setup with invalid provider configuration
func TestHandler_Setup_WithInvalidProvider(t *testing.T) {
	client := &MockSSHClient{}
	provider := NewCustomRepoProvider(
		"",                                 // Invalid empty name
		"http://insecure.example.com/rpm/", // Invalid HTTP URL
		"https://example.com/gpg-key",
		true,
		true,
	)

	handler := NewHandlerWithProvider(client, provider)
	var output bytes.Buffer

	err := handler.Setup(context.Background(), &output)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid repository configuration")
	// No SSH commands should be executed due to validation failure
	client.AssertNotCalled(t, "Execute")
}
