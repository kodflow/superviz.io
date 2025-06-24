package ssh

import (
	"bufio"
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// mockHostKeyStore implements HostKeyStore interface for testing
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

// mockUserPrompter implements UserPrompter interface for testing
type mockUserPrompter struct {
	mock.Mock
}

func (m *mockUserPrompter) PromptYesNo(message string) (bool, error) {
	args := m.Called(message)
	return args.Bool(0), args.Error(1)
}

// Test helper to create a test SSH key
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

	// Should be a defaultHostKeyManager
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

	// Test that it allows any host key (insecure)
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

	// Test that the callback works as expected
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

	// Should contain .ssh/known_hosts
	assert.Contains(t, path, ".ssh")
	assert.Contains(t, path, "known_hosts")

	// Should be cached
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

func TestTerminalPrompter_PromptYesNo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"yes", "yes\n", true},
		{"y", "y\n", true},
		{"YES", "YES\n", true},
		{"Y", "Y\n", true},
		{"no", "no\n", false},
		{"n", "n\n", false},
		{"NO", "NO\n", false},
		{"N", "N\n", false},
		{"maybe", "maybe\n", false},
		{"empty", "\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a reader with test input
			reader := bufio.NewReader(strings.NewReader(tt.input))

			prompter := &terminalPrompter{
				reader: reader,
			}

			result, err := prompter.PromptYesNo("Test prompt: ")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
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
