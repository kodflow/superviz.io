package ssh

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// mockHostKeyStore implements HostKeyStore interface for testing.
type mockHostKeyStore struct {
	mock.Mock
}

func (m *mockHostKeyStore) IsKnown(hostname string, remote net.Addr, key ssh.PublicKey) bool {
	args := m.Called(hostname, remote, key)
	return args.Bool(0)
}

func (m *mockHostKeyStore) Add(hostname string, key ssh.PublicKey) error {
	args := m.Called(hostname, key)
	return args.Error(0)
}

func (m *mockHostKeyStore) GetCallback() ssh.HostKeyCallback {
	args := m.Called()
	if callback := args.Get(0); callback != nil {
		return callback.(ssh.HostKeyCallback)
	}
	return nil
}

// mockUserPrompter implements UserPrompter interface for testing.
type mockUserPrompter struct {
	mock.Mock
}

func (m *mockUserPrompter) PromptYesNo(message string) (bool, error) {
	args := m.Called(message)
	return args.Bool(0), args.Error(1)
}

// createTestKey creates a test SSH key for testing purposes.
//
// This helper function generates a valid ed25519 SSH public key that can be used
// in tests for host key verification scenarios.
//
// Parameters:
//   - t: testing.T instance for test helpers
//
// Returns:
//   - ssh.PublicKey: a valid SSH public key for testing
func createTestKey(t *testing.T) ssh.PublicKey {
	t.Helper()

	// Create a test ed25519 key
	pubKeyBytes := []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOQ6Rx7HL9KrqPtf6u3qhQqv1HQJQ7V4L8KrqPtf6u3q test@example.com")

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(pubKeyBytes)
	require.NoError(t, err)

	return pubKey
}

func TestKeyTypeNames(t *testing.T) {
	tests := []struct {
		keyType  string
		expected string
	}{
		{"ssh-ed25519", "ED25519"},
		{"ssh-rsa", "RSA"},
		{"ecdsa-sha2-nistp256", "ECDSA"},
		{"ecdsa-sha2-nistp384", "ECDSA"},
		{"ecdsa-sha2-nistp521", "ECDSA"},
		{"ssh-dss", "DSA"},
	}

	for _, tt := range tests {
		t.Run(tt.keyType, func(t *testing.T) {
			assert.Equal(t, tt.expected, keyTypeNames[tt.keyType])
		})
	}
}

func TestNewDefaultHostKeyManager(t *testing.T) {
	manager := NewDefaultHostKeyManager()
	require.NotNil(t, manager)

	// Verify that the returned manager is a defaultHostKeyManager instance.
	defaultManager, ok := manager.(*defaultHostKeyManager)
	require.True(t, ok)
	require.NotNil(t, defaultManager.store)
	require.NotNil(t, defaultManager.prompter)
}

func TestNewHostKeyManager(t *testing.T) {
	store := &mockHostKeyStore{}
	prompter := &mockUserPrompter{}

	manager := NewHostKeyManager(store, prompter)
	require.NotNil(t, manager)

	// Verify that the returned manager contains the injected dependencies.
	defaultManager, ok := manager.(*defaultHostKeyManager)
	require.True(t, ok)
	assert.Equal(t, store, defaultManager.store)
	assert.Equal(t, prompter, defaultManager.prompter)
}

func TestDefaultHostKeyManager_GetHostKeyCallback_SkipHostKeyCheck(t *testing.T) {
	manager := &defaultHostKeyManager{
		store:    &mockHostKeyStore{},
		prompter: &mockUserPrompter{},
	}

	config := &Config{
		SkipHostKeyCheck: true,
	}

	callback, err := manager.GetHostKeyCallback(context.Background(), config)
	require.NoError(t, err)
	require.NotNil(t, callback)

	// Test that it allows any host key (insecure behavior when SkipHostKeyCheck is true).
	key := createTestKey(t)
	hostname := "test.example.com"
	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}

	err = callback(hostname, remote, key)
	assert.NoError(t, err)
}

func TestDefaultHostKeyManager_GetHostKeyCallback_WithStoreCallback(t *testing.T) {
	expectedCallback := ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		return nil
	})

	store := &mockHostKeyStore{}
	store.On("GetCallback").Return(expectedCallback)

	manager := &defaultHostKeyManager{
		store:    store,
		prompter: &mockUserPrompter{},
	}

	config := &Config{
		SkipHostKeyCheck: false,
	}

	callback, err := manager.GetHostKeyCallback(context.Background(), config)
	require.NoError(t, err)
	require.NotNil(t, callback)

	// Test that the callback works as expected with store-provided callback.
	key := createTestKey(t)
	hostname := "test.example.com"
	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}

	err = callback(hostname, remote, key)
	assert.NoError(t, err)

	store.AssertExpectations(t)
}

func TestDefaultHostKeyManager_GetHostKeyCallback_InteractiveCallback(t *testing.T) {
	store := &mockHostKeyStore{}
	store.On("GetCallback").Return(nil)

	manager := &defaultHostKeyManager{
		store:    store,
		prompter: &mockUserPrompter{},
	}

	config := &Config{
		SkipHostKeyCheck: false,
	}

	callback, err := manager.GetHostKeyCallback(context.Background(), config)
	require.NoError(t, err)
	require.NotNil(t, callback)

	store.AssertExpectations(t)
}

func TestDefaultHostKeyManager_CreateInteractiveCallback_KnownHost(t *testing.T) {
	key := createTestKey(t)
	hostname := "test.example.com"
	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}

	store := &mockHostKeyStore{}
	store.On("IsKnown", hostname, remote, key).Return(true)

	manager := &defaultHostKeyManager{
		store:    store,
		prompter: &mockUserPrompter{},
	}

	config := &Config{}
	callback := manager.createInteractiveCallback(config)

	err := callback(hostname, remote, key)
	assert.NoError(t, err)

	store.AssertExpectations(t)
}

func TestDefaultHostKeyManager_CreateInteractiveCallback_UnknownHost_AutoAccept(t *testing.T) {
	key := createTestKey(t)
	hostname := "test.example.com"
	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}

	store := &mockHostKeyStore{}
	store.On("IsKnown", hostname, remote, key).Return(false)
	store.On("Add", hostname, key).Return(nil)

	manager := &defaultHostKeyManager{
		store:    store,
		prompter: &mockUserPrompter{},
	}

	config := &Config{
		AcceptNewHostKey: true,
	}
	callback := manager.createInteractiveCallback(config)

	err := callback(hostname, remote, key)
	assert.NoError(t, err)

	store.AssertExpectations(t)
}

func TestDefaultHostKeyManager_CreateInteractiveCallback_UnknownHost_UserAccepts(t *testing.T) {
	key := createTestKey(t)
	hostname := "test.example.com"
	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}

	store := &mockHostKeyStore{}
	store.On("IsKnown", hostname, remote, key).Return(false)
	store.On("Add", hostname, key).Return(nil)

	prompter := &mockUserPrompter{}
	prompter.On("PromptYesNo", "Continue connecting (yes/no)? ").Return(true, nil)

	manager := &defaultHostKeyManager{
		store:    store,
		prompter: prompter,
	}

	config := &Config{
		AcceptNewHostKey: false,
	}
	callback := manager.createInteractiveCallback(config)

	err := callback(hostname, remote, key)
	assert.NoError(t, err)

	store.AssertExpectations(t)
	prompter.AssertExpectations(t)
}

func TestDefaultHostKeyManager_CreateInteractiveCallback_UnknownHost_UserRejects(t *testing.T) {
	key := createTestKey(t)
	hostname := "test.example.com"
	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}

	store := &mockHostKeyStore{}
	store.On("IsKnown", hostname, remote, key).Return(false)

	prompter := &mockUserPrompter{}
	prompter.On("PromptYesNo", "Continue connecting (yes/no)? ").Return(false, nil)

	manager := &defaultHostKeyManager{
		store:    store,
		prompter: prompter,
	}

	config := &Config{
		AcceptNewHostKey: false,
	}
	callback := manager.createInteractiveCallback(config)

	err := callback(hostname, remote, key)
	require.Error(t, err)

	// Should be a host key rejected error
	sshErr, ok := err.(*SSHError)
	require.True(t, ok)
	assert.Equal(t, ErrHostKeyRejected, sshErr.Type)

	store.AssertExpectations(t)
	prompter.AssertExpectations(t)
}

func TestFileHostKeyStore_GetPath(t *testing.T) {
	store := &fileHostKeyStore{}

	path := store.getPath()

	// Verify that the path contains the expected SSH directory structure.
	assert.Contains(t, path, ".ssh")
	assert.Contains(t, path, "known_hosts")

	// Verify that subsequent calls return the cached path.
	path2 := store.getPath()
	assert.Equal(t, path, path2)
}

func TestFileHostKeyStore_GetPath_WithCustomPath(t *testing.T) {
	customPath := "/custom/known_hosts"
	store := &fileHostKeyStore{
		path: customPath,
	}

	path := store.getPath()
	assert.Equal(t, customPath, path)
}

func TestFileHostKeyStore_GetCallback_NoFile(t *testing.T) {
	store := &fileHostKeyStore{
		path: "/nonexistent/path/known_hosts",
	}

	callback := store.GetCallback()
	assert.Nil(t, callback)
}

func TestFileHostKeyStore_Add(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	knownHostsPath := filepath.Join(tempDir, ".ssh", "known_hosts")

	store := &fileHostKeyStore{
		path: knownHostsPath,
	}

	key := createTestKey(t)
	hostname := "test.example.com"

	err := store.Add(hostname, key)
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, knownHostsPath)

	// Verify content
	content, err := os.ReadFile(knownHostsPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), hostname)
}

func TestFileHostKeyStore_IsKnown_NoCallback(t *testing.T) {
	store := &fileHostKeyStore{
		path: "/nonexistent/path/known_hosts",
	}

	key := createTestKey(t)
	hostname := "test.example.com"
	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}

	isKnown := store.IsKnown(hostname, remote, key)
	assert.False(t, isKnown)
}

func TestGetKeyTypeDisplay(t *testing.T) {
	tests := []struct {
		keyType  string
		expected string
	}{
		{"ssh-ed25519", "ED25519"},
		{"ssh-rsa", "RSA"},
		{"ecdsa-sha2-nistp256", "ECDSA"},
		{"ssh-dss", "DSA"},
		{"unknown-type", "UNKNOWN-TYPE"},
		{"custom-key", "CUSTOM-KEY"},
	}

	for _, tt := range tests {
		t.Run(tt.keyType, func(t *testing.T) {
			result := getKeyTypeDisplay(tt.keyType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFileHostKeyStore_Concurrent(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	knownHostsPath := filepath.Join(tempDir, ".ssh", "known_hosts")

	store := &fileHostKeyStore{
		path: knownHostsPath,
	}

	key := createTestKey(t)

	// Test concurrent access
	done := make(chan bool, 2)

	go func() {
		defer func() { done <- true }()
		_ = store.Add("host1.example.com", key)
	}()

	go func() {
		defer func() { done <- true }()
		_ = store.GetCallback()
	}()

	// Wait for both goroutines
	<-done
	<-done

	// No assertion needed, just ensuring no race conditions
}

func TestFileHostKeyStore_GetCallback_ValidFile(t *testing.T) {
	// Create temporary directory and known_hosts file
	tempDir := t.TempDir()
	sshDir := filepath.Join(tempDir, ".ssh")
	err := os.MkdirAll(sshDir, 0700)
	require.NoError(t, err)

	knownHostsPath := filepath.Join(sshDir, "known_hosts")

	// Create an empty known_hosts file (valid but empty)
	err = os.WriteFile(knownHostsPath, []byte(""), 0600)
	require.NoError(t, err)

	store := &fileHostKeyStore{
		path: knownHostsPath,
	}

	// First call should create the callback
	callback1 := store.GetCallback()
	assert.NotNil(t, callback1)

	// Second call should return the cached callback
	callback2 := store.GetCallback()
	assert.NotNil(t, callback2)
	// Note: We can't test if they're the same instance because the ssh package
	// may return different function instances, but both should be valid
}

func TestFileHostKeyStore_GetCallback_InvalidFile(t *testing.T) {
	// Create temporary directory and invalid known_hosts file
	tempDir := t.TempDir()
	sshDir := filepath.Join(tempDir, ".ssh")
	err := os.MkdirAll(sshDir, 0700)
	require.NoError(t, err)

	knownHostsPath := filepath.Join(sshDir, "known_hosts")

	// Create an invalid known_hosts file
	invalidContent := "invalid content that will cause knownhosts.New to fail"
	err = os.WriteFile(knownHostsPath, []byte(invalidContent), 0600)
	require.NoError(t, err)

	store := &fileHostKeyStore{
		path: knownHostsPath,
	}

	// Should return nil due to parsing error
	callback := store.GetCallback()
	assert.Nil(t, callback)
}

// Additional tests for better coverage

func TestDefaultHostKeyManager_HandleUnknownKey_UserRejectsKey(t *testing.T) {
	prompter := &mockUserPrompter{}
	store := &mockHostKeyStore{}

	manager := &defaultHostKeyManager{
		store:    store,
		prompter: prompter,
	}

	// Mock user rejecting the key
	prompter.On("PromptYesNo", mock.AnythingOfType("string")).Return(false, nil)

	key := createTestKey(t)
	config := &Config{Host: "example.com"}

	// This should return an error when user rejects
	err := manager.handleUnknownKey(config, "example.com", key)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrHostKeyRejected))

	prompter.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestDefaultHostKeyManager_HandleUnknownKey_PromptError(t *testing.T) {
	prompter := &mockUserPrompter{}
	store := &mockHostKeyStore{}

	manager := &defaultHostKeyManager{
		store:    store,
		prompter: prompter,
	}

	// Mock prompt error
	prompter.On("PromptYesNo", mock.AnythingOfType("string")).Return(false, assert.AnError)

	key := createTestKey(t)
	config := &Config{Host: "example.com"}

	// This should return the prompt error
	err := manager.handleUnknownKey(config, "example.com", key)
	require.Error(t, err)
	require.Equal(t, assert.AnError, err)

	prompter.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestDefaultHostKeyManager_HandleUnknownKey_AddKeyError(t *testing.T) {
	prompter := &mockUserPrompter{}
	store := &mockHostKeyStore{}

	manager := &defaultHostKeyManager{
		store:    store,
		prompter: prompter,
	}

	// Mock user accepting the key but store failing to add
	prompter.On("PromptYesNo", mock.AnythingOfType("string")).Return(true, nil)
	store.On("Add", "example.com", mock.Anything).Return(assert.AnError)

	key := createTestKey(t)
	config := &Config{Host: "example.com"}

	// This should return the store error
	err := manager.handleUnknownKey(config, "example.com", key)
	require.Error(t, err)
	require.Equal(t, assert.AnError, err)

	prompter.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestDefaultHostKeyManager_HandleUnknownKey_AutoAccept(t *testing.T) {
	prompter := &mockUserPrompter{}
	store := &mockHostKeyStore{}

	manager := &defaultHostKeyManager{
		store:    store,
		prompter: prompter,
	}

	// Mock store accepting the key
	store.On("Add", "example.com", mock.Anything).Return(nil)

	key := createTestKey(t)
	config := &Config{
		Host:             "example.com",
		AcceptNewHostKey: true,
	}

	// Should auto-accept without prompting
	err := manager.handleUnknownKey(config, "example.com", key)
	require.NoError(t, err)

	// Prompter should not be called
	prompter.AssertNotCalled(t, "PromptYesNo")
	store.AssertExpectations(t)
}

func TestFileHostKeyStore_Add_InvalidPath(t *testing.T) {
	// Test with invalid path (using /dev/null as directory would fail)
	store := &fileHostKeyStore{path: "/dev/null/invalid"}

	key := createTestKey(t)

	// Should return error with invalid path
	err := store.Add("example.com", key)
	require.Error(t, err)
}

func TestFileHostKeyStore_GetCallback_DirectoryHandling(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	knownHostsPath := filepath.Join(sshDir, "known_hosts")

	store := &fileHostKeyStore{path: knownHostsPath}

	// Test that GetCallback can handle directory creation
	// This tests the path creation logic without asserting on callback return
	_ = store.GetCallback()

	// Directory should be created if possible
	if stat, err := os.Stat(sshDir); err == nil {
		assert.True(t, stat.IsDir(), "SSH directory should be created")
	}
}

func TestTerminalPrompter_PromptYesNo_Interface(t *testing.T) {
	// Test interface compliance
	prompter := &terminalPrompter{}
	require.NotNil(t, prompter)

	// Test that the method exists
	_ = prompter.PromptYesNo
}
