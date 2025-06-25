package repository

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
	"github.com/kodflow/superviz.io/internal/providers"
)

// Mock implementations

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

// Tests for Setup

func TestNewSetup(t *testing.T) {
	client := &mockSSHClient{}
	provider := &mockInstallProvider{}

	setup := NewSetup(client, provider)

	require.NotNil(t, setup)

	// Verify it implements the Setup interface
	var _ = Setup(setup)
}

func TestSetup_Setup_Ubuntu(t *testing.T) {
	client := &mockSSHClient{}
	provider := &mockInstallProvider{}

	// Mock all SSH commands to succeed - debian handler will call multiple commands
	client.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	setup := NewSetup(client, provider)
	var output bytes.Buffer

	err := setup.Setup(context.Background(), "ubuntu", &output)

	assert.NoError(t, err)
	assert.Contains(t, output.String(), "Setting up APT repository")
	client.AssertExpectations(t)
}

func TestSetup_Setup_Debian(t *testing.T) {
	client := &mockSSHClient{}
	provider := &mockInstallProvider{}

	// Mock all SSH commands to succeed
	client.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	setup := NewSetup(client, provider)
	var output bytes.Buffer

	err := setup.Setup(context.Background(), "debian", &output)

	assert.NoError(t, err)
	assert.Contains(t, output.String(), "Setting up APT repository")
	client.AssertExpectations(t)
}

func TestSetup_Setup_Alpine(t *testing.T) {
	client := &mockSSHClient{}
	provider := &mockInstallProvider{}

	// Mock all SSH commands to succeed
	client.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	setup := NewSetup(client, provider)
	var output bytes.Buffer

	err := setup.Setup(context.Background(), "alpine", &output)

	assert.NoError(t, err)
	client.AssertExpectations(t)
}

func TestSetup_Setup_CentOS(t *testing.T) {
	client := &mockSSHClient{}
	provider := &mockInstallProvider{}

	// Mock all SSH commands to succeed
	client.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	setup := NewSetup(client, provider)
	var output bytes.Buffer

	err := setup.Setup(context.Background(), "centos", &output)

	assert.NoError(t, err)
	client.AssertExpectations(t)
}

func TestSetup_Setup_RHEL(t *testing.T) {
	client := &mockSSHClient{}
	provider := &mockInstallProvider{}

	// Mock all SSH commands to succeed
	client.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	setup := NewSetup(client, provider)
	var output bytes.Buffer

	err := setup.Setup(context.Background(), "rhel", &output)

	assert.NoError(t, err)
	client.AssertExpectations(t)
}

func TestSetup_Setup_Fedora(t *testing.T) {
	client := &mockSSHClient{}
	provider := &mockInstallProvider{}

	// Mock all SSH commands to succeed
	client.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	setup := NewSetup(client, provider)
	var output bytes.Buffer

	err := setup.Setup(context.Background(), "fedora", &output)

	assert.NoError(t, err)
	client.AssertExpectations(t)
}

func TestSetup_Setup_Arch(t *testing.T) {
	client := &mockSSHClient{}
	provider := &mockInstallProvider{}

	// Mock provider method that arch handler will call
	provider.On("GetGPGKeyID").Return("test-gpg-key-id")

	// Mock all SSH commands to succeed
	client.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	setup := NewSetup(client, provider)
	var output bytes.Buffer

	err := setup.Setup(context.Background(), "arch", &output)

	assert.NoError(t, err)
	client.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestSetup_Setup_UnsupportedDistribution(t *testing.T) {
	client := &mockSSHClient{}
	provider := &mockInstallProvider{}

	setup := NewSetup(client, provider)
	var output bytes.Buffer

	err := setup.Setup(context.Background(), "unsupported", &output)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported distribution: unsupported")
}

func TestSetup_Setup_CommandError(t *testing.T) {
	client := &mockSSHClient{}
	provider := &mockInstallProvider{}

	// Mock all commands to fail - this will trigger the "sudo is not available" error
	client.On("Execute", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("command failed"))

	setup := NewSetup(client, provider)
	var output bytes.Buffer

	err := setup.Setup(context.Background(), "ubuntu", &output)

	assert.Error(t, err)
	// The actual error message depends on the sudo detection logic
	assert.True(t, err != nil, "Should return an error")
	client.AssertExpectations(t)
}

// Test the interface implementation
func TestSetup_ImplementsInterface(t *testing.T) {
	client := &mockSSHClient{}
	provider := &mockInstallProvider{}

	setup := NewSetup(client, provider)

	// Verify it implements the Setup interface
	var _ = Setup(setup)
}
