package services

import (
	"context"
	"errors"
	"testing"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSSHClient for testing detector
type MockSSHClientDetector struct {
	mock.Mock
}

func (m *MockSSHClientDetector) Connect(ctx context.Context, config *ssh.Config) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockSSHClientDetector) Execute(ctx context.Context, command string) error {
	args := m.Called(ctx, command)
	return args.Error(0)
}

func (m *MockSSHClientDetector) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewDetector(t *testing.T) {
	client := &MockSSHClientDetector{}
	detectorInstance := NewDetector(client)

	assert.NotNil(t, detectorInstance)
	assert.Equal(t, client, detectorInstance.(*detector).client)
}

func TestDetector_Detect_Ubuntu(t *testing.T) {
	client := &MockSSHClientDetector{}

	// Mock os-release file exists
	client.On("Execute", mock.Anything, "test -f /etc/os-release").Return(nil)

	// Mock distribution checks - Ubuntu succeeds, others fail (but may not be called due to map iteration order)
	client.On("Execute", mock.Anything, "grep -q 'ID=ubuntu' /etc/os-release").Return(nil).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=debian' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=alpine' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q '^ID=\"\\?centos' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q '^ID=\"\\?rhel' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=fedora' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=arch' /etc/os-release").Return(errors.New("not found")).Maybe()

	detectorInstance := NewDetector(client)
	result, err := detectorInstance.Detect(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "ubuntu", result)
	client.AssertExpectations(t)
}

func TestDetector_Detect_Debian(t *testing.T) {
	client := &MockSSHClientDetector{}

	// Mock os-release file exists
	client.On("Execute", mock.Anything, "test -f /etc/os-release").Return(nil)

	// Mock distribution checks - Debian succeeds, others fail (maybe called in any order)
	client.On("Execute", mock.Anything, "grep -q 'ID=ubuntu' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=debian' /etc/os-release").Return(nil).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=alpine' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q '^ID=\"\\?centos' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q '^ID=\"\\?rhel' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=fedora' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=arch' /etc/os-release").Return(errors.New("not found")).Maybe()

	detectorInstance := NewDetector(client)
	result, err := detectorInstance.Detect(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "debian", result)
	client.AssertExpectations(t)
}

func TestDetector_Detect_Alpine(t *testing.T) {
	client := &MockSSHClientDetector{}

	// Mock os-release file exists
	client.On("Execute", mock.Anything, "test -f /etc/os-release").Return(nil)

	// Mock distribution checks - Alpine succeeds, others fail (maybe called in any order)
	client.On("Execute", mock.Anything, "grep -q 'ID=ubuntu' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=debian' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=alpine' /etc/os-release").Return(nil).Maybe()
	client.On("Execute", mock.Anything, "grep -q '^ID=\"\\?centos' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q '^ID=\"\\?rhel' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=fedora' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=arch' /etc/os-release").Return(errors.New("not found")).Maybe()

	detectorInstance := NewDetector(client)
	result, err := detectorInstance.Detect(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "alpine", result)
	client.AssertExpectations(t)
}

func TestDetector_Detect_CentOS(t *testing.T) {
	client := &MockSSHClientDetector{}

	// Mock os-release file exists
	client.On("Execute", mock.Anything, "test -f /etc/os-release").Return(nil)

	// Mock distribution checks - CentOS succeeds, others fail (maybe called in any order)
	client.On("Execute", mock.Anything, "grep -q 'ID=ubuntu' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=debian' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=alpine' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q '^ID=\"\\?centos' /etc/os-release").Return(nil).Maybe()
	client.On("Execute", mock.Anything, "grep -q '^ID=\"\\?rhel' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=fedora' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=arch' /etc/os-release").Return(errors.New("not found")).Maybe()

	detectorInstance := NewDetector(client)
	result, err := detectorInstance.Detect(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "centos", result)
	client.AssertExpectations(t)
}

func TestDetector_Detect_RHEL(t *testing.T) {
	client := &MockSSHClientDetector{}

	// Mock os-release file exists
	client.On("Execute", mock.Anything, "test -f /etc/os-release").Return(nil)

	// Mock distribution checks - RHEL succeeds, others fail (maybe called in any order)
	client.On("Execute", mock.Anything, "grep -q 'ID=ubuntu' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=debian' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=alpine' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q '^ID=\"\\?centos' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q '^ID=\"\\?rhel' /etc/os-release").Return(nil).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=fedora' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=arch' /etc/os-release").Return(errors.New("not found")).Maybe()

	detectorInstance := NewDetector(client)
	result, err := detectorInstance.Detect(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "rhel", result)
	client.AssertExpectations(t)
}

func TestDetector_Detect_Fedora(t *testing.T) {
	client := &MockSSHClientDetector{}

	// Mock os-release file exists
	client.On("Execute", mock.Anything, "test -f /etc/os-release").Return(nil)

	// Mock distribution checks - Fedora succeeds, others fail (maybe called in any order)
	client.On("Execute", mock.Anything, "grep -q 'ID=ubuntu' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=debian' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=alpine' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q '^ID=\"\\?centos' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q '^ID=\"\\?rhel' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=fedora' /etc/os-release").Return(nil).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=arch' /etc/os-release").Return(errors.New("not found")).Maybe()

	detectorInstance := NewDetector(client)
	result, err := detectorInstance.Detect(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "fedora", result)
	client.AssertExpectations(t)
}

func TestDetector_Detect_Arch(t *testing.T) {
	client := &MockSSHClientDetector{}

	// Mock os-release file exists
	client.On("Execute", mock.Anything, "test -f /etc/os-release").Return(nil)

	// Mock distribution checks - Arch succeeds, others fail (maybe called in any order)
	client.On("Execute", mock.Anything, "grep -q 'ID=ubuntu' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=debian' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=alpine' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q '^ID=\"\\?centos' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q '^ID=\"\\?rhel' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=fedora' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=arch' /etc/os-release").Return(nil).Maybe()

	detectorInstance := NewDetector(client)
	result, err := detectorInstance.Detect(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "arch", result)
	client.AssertExpectations(t)
}

func TestDetector_Detect_FallbackDebian(t *testing.T) {
	client := &MockSSHClientDetector{}

	// Mock os-release file doesn't exist
	client.On("Execute", mock.Anything, "test -f /etc/os-release").Return(errors.New("not found"))
	// Mock apt command exists (fallback to Debian)
	client.On("Execute", mock.Anything, "command -v apt >/dev/null 2>&1").Return(nil)

	detectorInstance := NewDetector(client)
	result, err := detectorInstance.Detect(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "debian", result)
	client.AssertExpectations(t)
}

func TestDetector_Detect_FallbackAlpine(t *testing.T) {
	client := &MockSSHClientDetector{}

	// Mock os-release file doesn't exist
	client.On("Execute", mock.Anything, "test -f /etc/os-release").Return(errors.New("not found"))
	// Mock apt not found, apk found
	client.On("Execute", mock.Anything, "command -v apt >/dev/null 2>&1").Return(errors.New("not found"))
	client.On("Execute", mock.Anything, "command -v apk >/dev/null 2>&1").Return(nil)

	detectorInstance := NewDetector(client)
	result, err := detectorInstance.Detect(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "alpine", result)
	client.AssertExpectations(t)
}

func TestDetector_Detect_FallbackCentOS(t *testing.T) {
	client := &MockSSHClientDetector{}

	// Mock os-release file doesn't exist
	client.On("Execute", mock.Anything, "test -f /etc/os-release").Return(errors.New("not found"))
	// Mock apt and apk not found, yum found
	client.On("Execute", mock.Anything, "command -v apt >/dev/null 2>&1").Return(errors.New("not found"))
	client.On("Execute", mock.Anything, "command -v apk >/dev/null 2>&1").Return(errors.New("not found"))
	client.On("Execute", mock.Anything, "command -v yum >/dev/null 2>&1").Return(nil)

	detectorInstance := NewDetector(client)
	result, err := detectorInstance.Detect(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "centos", result)
	client.AssertExpectations(t)
}

func TestDetector_Detect_FallbackArch(t *testing.T) {
	client := &MockSSHClientDetector{}

	// Mock os-release file doesn't exist
	client.On("Execute", mock.Anything, "test -f /etc/os-release").Return(errors.New("not found"))
	// Mock previous package managers not found, pacman found
	client.On("Execute", mock.Anything, "command -v apt >/dev/null 2>&1").Return(errors.New("not found"))
	client.On("Execute", mock.Anything, "command -v apk >/dev/null 2>&1").Return(errors.New("not found"))
	client.On("Execute", mock.Anything, "command -v yum >/dev/null 2>&1").Return(errors.New("not found"))
	client.On("Execute", mock.Anything, "command -v pacman >/dev/null 2>&1").Return(nil)

	detectorInstance := NewDetector(client)
	result, err := detectorInstance.Detect(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "arch", result)
	client.AssertExpectations(t)
}

func TestDetector_Detect_Unknown(t *testing.T) {
	client := &MockSSHClientDetector{}

	// Mock os-release file doesn't exist
	client.On("Execute", mock.Anything, "test -f /etc/os-release").Return(errors.New("not found"))
	// Mock all package managers not found
	client.On("Execute", mock.Anything, "command -v apt >/dev/null 2>&1").Return(errors.New("not found"))
	client.On("Execute", mock.Anything, "command -v apk >/dev/null 2>&1").Return(errors.New("not found"))
	client.On("Execute", mock.Anything, "command -v yum >/dev/null 2>&1").Return(errors.New("not found"))
	client.On("Execute", mock.Anything, "command -v pacman >/dev/null 2>&1").Return(errors.New("not found"))

	detectorInstance := NewDetector(client)
	result, err := detectorInstance.Detect(context.Background())

	assert.Error(t, err)
	assert.Equal(t, "unknown", result)
	assert.Contains(t, err.Error(), "unable to detect distribution")
	client.AssertExpectations(t)
}

func TestDetector_Detect_NoOSReleaseAllDistrosFail(t *testing.T) {
	client := &MockSSHClientDetector{}

	// Mock os-release file exists but all distro checks fail
	client.On("Execute", mock.Anything, "test -f /etc/os-release").Return(nil)

	// Mock all distribution checks fail (maybe called in any order)
	client.On("Execute", mock.Anything, "grep -q 'ID=ubuntu' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=debian' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=alpine' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q '^ID=\"\\?centos' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q '^ID=\"\\?rhel' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=fedora' /etc/os-release").Return(errors.New("not found")).Maybe()
	client.On("Execute", mock.Anything, "grep -q 'ID=arch' /etc/os-release").Return(errors.New("not found")).Maybe()

	// Fallback to package manager detection - apt found
	client.On("Execute", mock.Anything, "command -v apt >/dev/null 2>&1").Return(nil)

	detectorInstance := NewDetector(client)
	result, err := detectorInstance.Detect(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "debian", result)
	client.AssertExpectations(t)
}
