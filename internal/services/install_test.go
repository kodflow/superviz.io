package services

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
	"github.com/kodflow/superviz.io/internal/providers"
)

// Mock implementations

type mockInstallProvider struct {
	mock.Mock
}

func (m *mockInstallProvider) GetInstallInfo() providers.InstallInfo {
	args := m.Called()
	return args.Get(0).(providers.InstallInfo)
}

func (m *mockInstallProvider) GetRepositoryURL() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockInstallProvider) GetPackageName() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockInstallProvider) GetGPGKeyID() string {
	args := m.Called()
	return args.String(0)
}

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

type mockDistroDetector struct {
	mock.Mock
}

func (m *mockDistroDetector) Detect(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

type mockRepoSetup struct {
	mock.Mock
}

func (m *mockRepoSetup) Setup(ctx context.Context, distro string, w io.Writer) error {
	args := m.Called(ctx, distro, w)
	return args.Error(0)
}

// Tests for bufferedWriter

func TestBufferedWriter_Write(t *testing.T) {
	var buf bytes.Buffer
	bw := &bufferedWriter{Writer: bufio.NewWriter(&buf)}

	data := []byte("test data")
	n, err := bw.Write(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.NoError(t, bw.Flush())
	assert.Equal(t, "test data", buf.String())
}

func TestBufferedWriter_Write_WithPreviousError(t *testing.T) {
	var buf bytes.Buffer
	bw := &bufferedWriter{
		Writer: bufio.NewWriter(&buf),
	}
	// Set atomic error
	bw.err.Store(errors.New("previous error"))

	data := []byte("test data")
	n, err := bw.Write(data)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
	assert.Contains(t, err.Error(), "previous error")
}

func TestBufferedWriter_Printf(t *testing.T) {
	var buf bytes.Buffer
	bw := &bufferedWriter{Writer: bufio.NewWriter(&buf)}

	bw.Printf("Hello %s, count: %d", "world", 42)

	assert.NoError(t, bw.Error())
	assert.Equal(t, "Hello world, count: 42", buf.String())
}

func TestBufferedWriter_Printf_WithPreviousError(t *testing.T) {
	var buf bytes.Buffer
	bw := &bufferedWriter{
		Writer: bufio.NewWriter(&buf),
	}
	// Set atomic error
	bw.err.Store(errors.New("previous error"))

	bw.Printf("Hello %s", "world")

	// Should not write anything due to previous error
	assert.Error(t, bw.Error())
	assert.Equal(t, "", buf.String())
}

func TestBufferedWriter_Error(t *testing.T) {
	var buf bytes.Buffer
	bw := &bufferedWriter{Writer: bufio.NewWriter(&buf)}

	// No error initially
	assert.NoError(t, bw.Error())

	// Set atomic error
	bw.err.Store(errors.New("test error"))
	assert.Error(t, bw.Error())
	assert.Contains(t, bw.Error().Error(), "test error")
}

// Test bufferedWriter edge cases for better coverage
func TestBufferedWriter_WriteError(t *testing.T) {
	// Create a buffered writer with an existing error
	var buf bytes.Buffer
	bw := &bufferedWriter{Writer: bufio.NewWriter(&buf)}

	// Simulate a write error by setting an atomic error
	bw.err.Store(errors.New("write failed"))

	// Write should return the existing error without writing
	n, err := bw.Write([]byte("test"))
	assert.Equal(t, 0, n)
	assert.Error(t, err)
	assert.Equal(t, "write failed", err.Error())

	// Buffer should remain empty
	assert.Equal(t, "", buf.String())
}

func TestBufferedWriter_FlushError(t *testing.T) {
	var buf bytes.Buffer
	bw := &bufferedWriter{Writer: bufio.NewWriter(&buf)}

	// Set an atomic error state
	bw.err.Store(errors.New("flush error"))

	// Flush should return the existing error
	err := bw.Flush()
	assert.Error(t, err)
	assert.Equal(t, "flush error", err.Error())
}

// Mock writer that fails on write
type failingWriter struct {
	shouldFail bool
}

func (fw *failingWriter) Write(p []byte) (n int, err error) {
	if fw.shouldFail {
		return 0, errors.New("write operation failed")
	}
	return len(p), nil
}

func TestBufferedWriter_WriteFailure(t *testing.T) {
	failingWriter := &failingWriter{shouldFail: true}
	bw := &bufferedWriter{Writer: bufio.NewWriter(failingWriter)}

	// Write a large amount of data to trigger flush
	largeData := make([]byte, 5000) // Larger than default buffer size
	for i := range largeData {
		largeData[i] = 'A'
	}

	// Write should fail during flush
	_, err := bw.Write(largeData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "write operation failed")

	// Subsequent writes should return the stored error
	n2, err2 := bw.Write([]byte("test2"))
	assert.Equal(t, 0, n2)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "write operation failed")
}

// Tests for InstallService

func TestNewInstallService_WithNilOptions(t *testing.T) {
	service := NewInstallService(nil)

	require.NotNil(t, service)
	assert.NotNil(t, service.provider)
	assert.NotNil(t, service.client)
	assert.NotNil(t, service.detector)
	assert.NotNil(t, service.repoSetup)
}

func TestNewInstallService_WithOptions(t *testing.T) {
	provider := &mockInstallProvider{}
	client := &mockSSHClient{}
	detector := &mockDistroDetector{}
	repoSetup := &mockRepoSetup{}

	opts := &InstallServiceOptions{
		Provider:       provider,
		SSHClient:      client,
		DistroDetector: detector,
		RepoSetup:      repoSetup,
	}

	service := NewInstallService(opts)

	require.NotNil(t, service)
	assert.Equal(t, provider, service.provider)
	assert.Equal(t, client, service.client)
	assert.Equal(t, detector, service.detector)
	assert.Equal(t, repoSetup, service.repoSetup)
}

func TestNewInstallService_WithPartialOptions(t *testing.T) {
	provider := &mockInstallProvider{}

	opts := &InstallServiceOptions{
		Provider: provider,
		// Other fields are nil
	}

	service := NewInstallService(opts)

	require.NotNil(t, service)
	assert.Equal(t, provider, service.provider)
	assert.NotNil(t, service.client)    // Should get default
	assert.NotNil(t, service.detector)  // Should get default
	assert.NotNil(t, service.repoSetup) // Should get default
}

func TestInstallService_ValidateAndPrepareConfig_Valid(t *testing.T) {
	service := NewInstallService(nil)
	config := &providers.InstallConfig{}
	args := []string{"user@host.example.com"}

	err := service.ValidateAndPrepareConfig(config, args)

	assert.NoError(t, err)
	assert.Equal(t, "user", config.User)
	assert.Equal(t, "host.example.com", config.Host)
	assert.Equal(t, "user@host.example.com", config.Target)
}

func TestInstallService_ValidateAndPrepareConfig_NilConfig(t *testing.T) {
	service := NewInstallService(nil)

	err := service.ValidateAndPrepareConfig(nil, []string{"user@host"})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNilConfig)
}

func TestInstallService_ValidateAndPrepareConfig_NoArgs(t *testing.T) {
	service := NewInstallService(nil)
	config := &providers.InstallConfig{}

	err := service.ValidateAndPrepareConfig(config, []string{})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidTarget)
}

func TestInstallService_ValidateAndPrepareConfig_InvalidFormat(t *testing.T) {
	service := NewInstallService(nil)
	config := &providers.InstallConfig{}

	tests := []struct {
		name   string
		target string
	}{
		{"no_at_symbol", "userhost"},
		{"starts_with_at", "@host"},
		{"ends_with_at", "user@"},
		{"only_at", "@"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateAndPrepareConfig(config, []string{tt.target})
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidTarget)
		})
	}
}

func TestInstallService_Install_Success(t *testing.T) {
	// Setup mocks
	client := &mockSSHClient{}
	detector := &mockDistroDetector{}
	repoSetup := &mockRepoSetup{}
	provider := &mockInstallProvider{}

	client.On("Connect", mock.Anything, mock.Anything).Return(nil)
	client.On("Close").Return(nil)
	detector.On("Detect", mock.Anything).Return("ubuntu", nil)
	repoSetup.On("Setup", mock.Anything, "ubuntu", mock.Anything).Return(nil)

	opts := &InstallServiceOptions{
		Provider:       provider,
		SSHClient:      client,
		DistroDetector: detector,
		RepoSetup:      repoSetup,
	}

	service := NewInstallService(opts)

	config := &providers.InstallConfig{
		Host:   "test.example.com",
		User:   "testuser",
		Target: "testuser@test.example.com",
	}

	var output bytes.Buffer
	err := service.Install(context.Background(), &output, config)

	assert.NoError(t, err)

	outputStr := output.String()
	assert.Contains(t, outputStr, "Starting repository setup")
	assert.Contains(t, outputStr, "Connected to testuser@test.example.com")
	assert.Contains(t, outputStr, "Detected distribution: ubuntu")
	assert.Contains(t, outputStr, "Repository setup completed successfully")
	assert.Contains(t, outputStr, "sudo apt update && sudo apt install superviz")

	client.AssertExpectations(t)
	detector.AssertExpectations(t)
	repoSetup.AssertExpectations(t)
}

func TestInstallService_Install_NilWriter(t *testing.T) {
	service := NewInstallService(nil)
	config := &providers.InstallConfig{}

	err := service.Install(context.Background(), nil, config)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNilWriter)
}

func TestInstallService_Install_NilConfig(t *testing.T) {
	service := NewInstallService(nil)
	var output bytes.Buffer

	err := service.Install(context.Background(), &output, nil)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNilConfig)
}

func TestInstallService_Install_ConnectionError(t *testing.T) {
	client := &mockSSHClient{}
	client.On("Connect", mock.Anything, mock.Anything).Return(errors.New("connection failed"))

	opts := &InstallServiceOptions{
		SSHClient: client,
	}

	service := NewInstallService(opts)
	config := &providers.InstallConfig{
		Host:   "test.example.com",
		User:   "testuser",
		Target: "testuser@test.example.com",
	}

	var output bytes.Buffer
	err := service.Install(context.Background(), &output, config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to testuser@test.example.com")

	client.AssertExpectations(t)
}

func TestInstallService_Install_DetectionError(t *testing.T) {
	client := &mockSSHClient{}
	detector := &mockDistroDetector{}

	client.On("Connect", mock.Anything, mock.Anything).Return(nil)
	client.On("Close").Return(nil)
	detector.On("Detect", mock.Anything).Return("", errors.New("detection failed"))

	opts := &InstallServiceOptions{
		SSHClient:      client,
		DistroDetector: detector,
	}

	service := NewInstallService(opts)
	config := &providers.InstallConfig{
		Host:   "test.example.com",
		User:   "testuser",
		Target: "testuser@test.example.com",
	}

	var output bytes.Buffer
	err := service.Install(context.Background(), &output, config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to detect distribution")

	client.AssertExpectations(t)
	detector.AssertExpectations(t)
}

func TestInstallService_Install_RepoSetupError(t *testing.T) {
	client := &mockSSHClient{}
	detector := &mockDistroDetector{}
	repoSetup := &mockRepoSetup{}

	client.On("Connect", mock.Anything, mock.Anything).Return(nil)
	client.On("Close").Return(nil)
	detector.On("Detect", mock.Anything).Return("ubuntu", nil)
	repoSetup.On("Setup", mock.Anything, "ubuntu", mock.Anything).Return(errors.New("setup failed"))

	opts := &InstallServiceOptions{
		SSHClient:      client,
		DistroDetector: detector,
		RepoSetup:      repoSetup,
	}

	service := NewInstallService(opts)
	config := &providers.InstallConfig{
		Host:   "test.example.com",
		User:   "testuser",
		Target: "testuser@test.example.com",
	}

	var output bytes.Buffer
	err := service.Install(context.Background(), &output, config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to setup repository")

	client.AssertExpectations(t)
	detector.AssertExpectations(t)
	repoSetup.AssertExpectations(t)
}

func TestInstallService_Install_CloseError(t *testing.T) {
	client := &mockSSHClient{}
	detector := &mockDistroDetector{}
	repoSetup := &mockRepoSetup{}

	client.On("Connect", mock.Anything, mock.Anything).Return(nil)
	client.On("Close").Return(errors.New("close failed"))
	detector.On("Detect", mock.Anything).Return("ubuntu", nil)
	repoSetup.On("Setup", mock.Anything, "ubuntu", mock.Anything).Return(nil)

	opts := &InstallServiceOptions{
		SSHClient:      client,
		DistroDetector: detector,
		RepoSetup:      repoSetup,
	}

	service := NewInstallService(opts)
	config := &providers.InstallConfig{
		Host:   "test.example.com",
		User:   "testuser",
		Target: "testuser@test.example.com",
	}

	var output bytes.Buffer
	err := service.Install(context.Background(), &output, config)

	// Should succeed despite close error (best effort - error is logged but ignored)
	assert.NoError(t, err)

	// The main installation messages should be present
	outputStr := output.String()
	assert.Contains(t, outputStr, "Repository setup completed successfully")

	client.AssertExpectations(t)
	detector.AssertExpectations(t)
	repoSetup.AssertExpectations(t)
}

func TestInstallService_CreateSSHConfig(t *testing.T) {
	service := NewInstallService(nil)

	config := &providers.InstallConfig{
		Host:             "test.example.com",
		User:             "testuser",
		Port:             2222,
		KeyPath:          "/path/to/key",
		Timeout:          30 * time.Second,
		SkipHostKeyCheck: true,
	}

	sshConfig := service.createSSHConfig(config)

	assert.Equal(t, "test.example.com", sshConfig.Host)
	assert.Equal(t, "testuser", sshConfig.User)
	assert.Equal(t, 2222, sshConfig.Port)
	assert.Equal(t, "/path/to/key", sshConfig.KeyPath)
	assert.Equal(t, 30*time.Second, sshConfig.Timeout)
	assert.True(t, sshConfig.SkipHostKeyCheck)
	assert.True(t, sshConfig.AcceptNewHostKey)
}

func TestInstallService_WrapConnectionError(t *testing.T) {
	service := NewInstallService(nil)
	target := "user@host.example.com"

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "auth_error",
			err:      ssh.NewError(ssh.ErrAuthFailed, "permission denied"),
			expected: "authentication failed for user@host.example.com",
		},
		{
			name:     "connection_error",
			err:      ssh.NewError(ssh.ErrConnectionFailed, "connection refused"),
			expected: "failed to connect to user@host.example.com",
		},
		{
			name:     "generic_error",
			err:      errors.New("generic error"),
			expected: "failed to connect to user@host.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := service.wrapConnectionError(tt.err, target)
			assert.Contains(t, wrapped.Error(), tt.expected)
		})
	}
}

func TestInstallService_GetInstallCommand(t *testing.T) {
	service := NewInstallService(nil)

	tests := []struct {
		distro   string
		expected string
	}{
		{"ubuntu", "  sudo apt update && sudo apt install superviz\n"},
		{"debian", "  sudo apt update && sudo apt install superviz\n"},
		{"alpine", "  sudo apk update && sudo apk add superviz\n"},
		{"centos", "  sudo yum install superviz  # or dnf install superviz\n"},
		{"rhel", "  sudo yum install superviz  # or dnf install superviz\n"},
		{"fedora", "  sudo dnf install superviz\n"},
		{"arch", "  sudo pacman -S superviz\n"},
		{"suse", "  sudo zypper install superviz\n"},
		{"gentoo", "  sudo emerge superviz\n"},
		{"unknown", "  Please check your package manager documentation\n"},
		{"UBUNTU", "  sudo apt update && sudo apt install superviz\n"}, // Test case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.distro, func(t *testing.T) {
			result := service.getInstallCommand(tt.distro)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInstallService_GetInstallInfo(t *testing.T) {
	provider := &mockInstallProvider{}
	expectedInfo := providers.InstallInfo{
		RepositoryURL: "https://repo.example.com",
		PackageName:   "superviz",
		Version:       "1.0.0",
		GPGKeyID:      "test-key-id",
		Target:        "test-target",
	}

	provider.On("GetInstallInfo").Return(expectedInfo)

	opts := &InstallServiceOptions{
		Provider: provider,
	}

	service := NewInstallService(opts)
	info := service.GetInstallInfo()

	assert.Equal(t, expectedInfo, info)
	provider.AssertExpectations(t)
}

// Test install commands map
func TestInstallCommands(t *testing.T) {
	// Verify all expected distributions have commands
	expectedDistros := []string{
		"ubuntu", "debian", "alpine", "centos", "rhel",
		"fedora", "arch", "suse", "gentoo",
	}

	for _, distro := range expectedDistros {
		t.Run(distro, func(t *testing.T) {
			cmd, ok := installCommands[distro]
			assert.True(t, ok, "Missing command for distribution: %s", distro)
			assert.NotEmpty(t, cmd, "Empty command for distribution: %s", distro)
			assert.Contains(t, cmd, "superviz", "Command should contain 'superviz': %s", cmd)
		})
	}
}

func TestInstallService_Install_CloseErrorLogging(t *testing.T) {
	// Setup mocks
	sshClient := &mockSSHClient{}
	detector := &mockDistroDetector{}
	repoSetup := &mockRepoSetup{}
	provider := &mockInstallProvider{}

	opts := &InstallServiceOptions{
		Provider:       provider,
		SSHClient:      sshClient,
		DistroDetector: detector,
		RepoSetup:      repoSetup,
	}

	service := NewInstallService(opts)

	config := &providers.InstallConfig{
		Target: "user@host.com",
		User:   "user",
		Host:   "host.com",
	}

	var output bytes.Buffer

	// Mock operations - Close() will return an error
	sshClient.On("Connect", mock.Anything, mock.Anything).Return(nil)
	sshClient.On("Close").Return(errors.New("close failed"))
	detector.On("Detect", mock.Anything).Return("ubuntu", nil)
	repoSetup.On("Setup", mock.Anything, "ubuntu", mock.Anything).Return(nil)

	err := service.Install(context.Background(), &output, config)

	// Should succeed even with close error (best effort)
	assert.NoError(t, err)

	// Should contain successful installation output
	assert.Contains(t, output.String(), "Repository setup completed successfully")

	// Verify mocks - most importantly that Close was called
	sshClient.AssertExpectations(t)
	detector.AssertExpectations(t)
	repoSetup.AssertExpectations(t)
}
