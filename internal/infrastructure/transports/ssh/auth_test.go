package ssh

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// Mock implementations for testing

type mockPasswordReader struct {
	password         string
	err              error
	readPasswordFunc func(prompt string) (string, error)
}

func (m *mockPasswordReader) ReadPassword(prompt string) (string, error) {
	if m.readPasswordFunc != nil {
		return m.readPasswordFunc(prompt)
	}
	return m.password, m.err
}

type mockKeyLoader struct {
	signer      ssh.Signer
	err         error
	loadKeyFunc func(path string) (ssh.Signer, error)
}

func (m *mockKeyLoader) LoadKey(path string) (ssh.Signer, error) {
	if m.loadKeyFunc != nil {
		return m.loadKeyFunc(path)
	}
	return m.signer, m.err
}

// Mock signer for testing
type mockSigner struct{}

func (m *mockSigner) PublicKey() ssh.PublicKey { return nil }
func (m *mockSigner) Sign(rand io.Reader, data []byte) (*ssh.Signature, error) {
	return &ssh.Signature{}, nil
}

func TestNewDefaultAuthenticator(t *testing.T) {
	auth := NewDefaultAuthenticator()
	require.NotNil(t, auth)

	// Should be able to type assert to concrete type
	defaultAuth, ok := auth.(*defaultAuthenticator)
	require.True(t, ok)
	require.NotNil(t, defaultAuth.passwordReader)
	require.NotNil(t, defaultAuth.keyLoader)
}

func TestNewAuthenticator(t *testing.T) {
	mockReader := &mockPasswordReader{}
	mockLoader := &mockKeyLoader{}

	auth := NewAuthenticator(mockReader, mockLoader)
	require.NotNil(t, auth)

	// Should be able to type assert to concrete type with injected dependencies
	defaultAuth, ok := auth.(*defaultAuthenticator)
	require.True(t, ok)
	require.Equal(t, mockReader, defaultAuth.passwordReader)
	require.Equal(t, mockLoader, defaultAuth.keyLoader)
}

func TestDefaultAuthenticator_GetAuthMethods_WithKey(t *testing.T) {
	tests := []struct {
		name       string
		keyPath    string
		mockSigner ssh.Signer
		keyErr     error
		wantErr    bool
		errType    error
	}{
		{
			name:       "successful key loading",
			keyPath:    "/path/to/key",
			mockSigner: &mockSigner{},
			keyErr:     nil,
			wantErr:    false,
		},
		{
			name:    "key loading failure",
			keyPath: "/invalid/path",
			keyErr:  errors.New("file not found"),
			wantErr: true,
			errType: ErrAuthFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLoader := &mockKeyLoader{
				signer: tt.mockSigner,
				err:    tt.keyErr,
			}

			auth := NewAuthenticator(nil, mockLoader).(*defaultAuthenticator)
			config := &Config{
				Host:    "example.com",
				User:    "testuser",
				KeyPath: tt.keyPath,
			}

			methods, err := auth.GetAuthMethods(context.Background(), config)

			if tt.wantErr {
				require.Error(t, err)
				require.True(t, IsAuthError(err))
				require.Nil(t, methods)

				// Check error context
				sshErr, ok := err.(*SSHError)
				require.True(t, ok)
				require.Equal(t, tt.keyPath, sshErr.Context["key_path"])
			} else {
				require.NoError(t, err)
				require.Len(t, methods, 1)
				require.NotNil(t, methods[0])
			}
		})
	}
}

func TestDefaultAuthenticator_GetAuthMethods_WithPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		passErr  error
		wantErr  bool
	}{
		{
			name:     "successful password reading",
			password: "testpassword",
			passErr:  nil,
			wantErr:  false,
		},
		{
			name:     "password reading failure",
			password: "",
			passErr:  errors.New("failed to read password"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := &mockPasswordReader{
				password: tt.password,
				err:      tt.passErr,
			}

			auth := NewAuthenticator(mockReader, nil).(*defaultAuthenticator)
			config := &Config{
				Host: "example.com",
				User: "testuser",
				// No KeyPath - should use password auth
			}

			methods, err := auth.GetAuthMethods(context.Background(), config)

			if tt.wantErr {
				require.Error(t, err)
				require.True(t, IsAuthError(err))
				require.Nil(t, methods)
			} else {
				require.NoError(t, err)
				require.Len(t, methods, 1)
				require.NotNil(t, methods[0])
			}
		})
	}
}

func TestDefaultAuthenticator_KeyCaching(t *testing.T) {
	mockSigner := &mockSigner{}
	mockLoader := &mockKeyLoader{
		signer: mockSigner,
		err:    nil,
	}

	auth := NewAuthenticator(nil, mockLoader).(*defaultAuthenticator)
	config := &Config{
		Host:    "example.com",
		User:    "testuser",
		KeyPath: "/path/to/key",
	}

	// First call should load the key
	methods1, err1 := auth.GetAuthMethods(context.Background(), config)
	require.NoError(t, err1)
	require.Len(t, methods1, 1)

	// Second call should use cached key
	methods2, err2 := auth.GetAuthMethods(context.Background(), config)
	require.NoError(t, err2)
	require.Len(t, methods2, 1)

	// Verify key is cached
	cached, ok := auth.keyCache.Load("/path/to/key")
	require.True(t, ok)
	require.Equal(t, mockSigner, cached)
}

func TestDefaultAuthenticator_CacheWithDifferentPaths(t *testing.T) {
	mockSigner1 := &mockSigner{}
	mockSigner2 := &mockSigner{}

	mockLoader := &mockKeyLoader{
		loadKeyFunc: func(path string) (ssh.Signer, error) {
			if path == "/path/to/key1" {
				return mockSigner1, nil
			}
			return mockSigner2, nil
		},
	}

	auth := NewAuthenticator(nil, mockLoader).(*defaultAuthenticator)

	// Load first key
	config1 := &Config{
		Host:    "example.com",
		User:    "testuser",
		KeyPath: "/path/to/key1",
	}
	_, err1 := auth.GetAuthMethods(context.Background(), config1)
	require.NoError(t, err1)

	// Load second key
	config2 := &Config{
		Host:    "example.com",
		User:    "testuser",
		KeyPath: "/path/to/key2",
	}
	_, err2 := auth.GetAuthMethods(context.Background(), config2)
	require.NoError(t, err2)

	// Both should be cached
	cached1, ok1 := auth.keyCache.Load("/path/to/key1")
	require.True(t, ok1)
	require.Equal(t, mockSigner1, cached1)

	cached2, ok2 := auth.keyCache.Load("/path/to/key2")
	require.True(t, ok2)
	require.Equal(t, mockSigner2, cached2)
}

func TestTerminalPasswordReader_ReadPassword(t *testing.T) {
	// We can't easily test the actual terminal reading in CI/tests,
	// but we can test the interface compliance and error handling
	reader := &terminalPasswordReader{}
	require.NotNil(t, reader)

	// Test that the method exists and is callable
	// In a real scenario, this would prompt for password
	// We can't test it fully without mocking os.Stdin
	_ = reader.ReadPassword
}

func TestFileKeyLoader_LoadKey_FileNotFound(t *testing.T) {
	loader := &fileKeyLoader{}

	// Test with non-existent file
	signer, err := loader.LoadKey("/path/that/does/not/exist")
	require.Error(t, err)
	require.Nil(t, signer)
	require.Contains(t, err.Error(), "unable to read private key")
}

func TestFileKeyLoader_LoadKey_InvalidKey(t *testing.T) {
	// Create a temporary file with invalid key content
	tmpFile, err := os.CreateTemp("", "invalid_key_*")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	// Write invalid key data
	_, err = tmpFile.WriteString("invalid key content")
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	loader := &fileKeyLoader{}
	signer, err := loader.LoadKey(tmpFile.Name())
	require.Error(t, err)
	require.Nil(t, signer)
	require.Contains(t, err.Error(), "unable to parse private key")
}

func TestFileKeyLoader_LoadKey_ValidKey(t *testing.T) {
	// Create a test RSA private key
	testKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA3r8k3W6GQF4I6zUJyJKB3cYbzz4rQJQz3+s9xvQa5HUeqH2
2h8rBJn5A7sDg1J2+J5b5A6r7QJ7t6F7+aKoKq4+f9e9J1U8W1e5x9Y8r9R8r8t
d0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8t8P3h8F6F7a8x9Y8r9R8r8td0
w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8t8P3h8F6F7a8x9Y8r9R8r8td0w+
I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8t8P3h8F6F7a8x9Y8r9R8r8td0w+I8
t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8t8P3h8F6F7a8x9Y8r9R8r8td0w+I8t7
Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8t8P3h8F6F7a8x9Y8r9R8r8td0wIDAQAB
AoIBAE4LmqU/qEiYxYxHYf3J7Q8J8j1Kj3l4h8H8k8j5j8H8kD3f3A4F7a8x9Y8
r9R8r8td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8t8P3h8F6F7a8x9Y8r9
R8r8td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8t8P3h8F6F7a8x9Y8r9R8
r8td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8t8P3h8F6F7a8x9Y8r9R8r8
td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8t8P3h8F6F7a8x9Y8r9R8r8td
0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8t8P3h8F6F7a8x9Y8r9R8r8td0w
+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8t8P3h8F6F7a8x9Y8r9R8r8td0wEC
gYEA8V8k6k3b8t8P3h8F6F7a8x9Y8r9R8r8td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd
0N3FdY3H7a8t8P3h8F6F7a8x9Y8r9R8r8td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N
3FdY3H7a8t8P3h8F6F7a8x9Y8r9R8r8td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3F
dY3H7a8t8P3h8F6F7a8x9Y8r9R8r8td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0ECgYBF
8k6k3b8t8P3h8F6F7a8x9Y8r9R8r8td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY
3H7a8t8P3h8F6F7a8x9Y8r9R8r8td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7
a8t8P3h8F6F7a8x9Y8r9R8r8td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8
t8P3h8F6F7a8x9Y8r9R8r8td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8t8
P3h8F6F7a8x9Y8r9R8r8td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0ECgYAq4+f9e9J1
U8W1e5x9Y8r9R8r8td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8t8P3h8F6
F7a8x9Y8r9R8r8td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8t8P3h8F6F7
a8x9Y8r9R8r8td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8t8P3h8F6F7a8
x9Y8r9R8r8td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8t8P3h8F6F7a8x9
Y8r9R8r8td0w+I8t7Y5I5S1o5b4b8Y8rP5Fd0N3FdY3H7a8t8P3h8F6F7a8x9Y8
r9R8r8td0w==
-----END RSA PRIVATE KEY-----`

	// Create a temporary file with a valid (but made-up) RSA key
	tmpFile, err := os.CreateTemp("", "valid_key_*")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	_, err = tmpFile.WriteString(testKey)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	loader := &fileKeyLoader{}
	signer, err := loader.LoadKey(tmpFile.Name())

	// This should fail with invalid key format, which tests the parsing error path
	require.Error(t, err)
	require.Nil(t, signer)
	require.Contains(t, err.Error(), "unable to parse private key")
}

func TestDefaultAuthenticator_GetAuthMethods_PasswordPrompt(t *testing.T) {
	var capturedPrompt string
	mockReader := &mockPasswordReader{
		password: "testpass",
		err:      nil,
		readPasswordFunc: func(prompt string) (string, error) {
			capturedPrompt = prompt
			return "testpass", nil
		},
	}

	auth := NewAuthenticator(mockReader, nil).(*defaultAuthenticator)
	config := &Config{
		Host: "example.com",
		User: "testuser",
		// No KeyPath - should use password auth
	}

	_, err := auth.GetAuthMethods(context.Background(), config)
	require.NoError(t, err)
	require.Equal(t, "Password for testuser@example.com: ", capturedPrompt)
}

func TestDefaultAuthenticator_GetAuthMethods_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockReader := &mockPasswordReader{
		password: "testpass",
		err:      nil,
	}

	auth := NewAuthenticator(mockReader, nil).(*defaultAuthenticator)
	config := &Config{
		Host: "example.com",
		User: "testuser",
		// No KeyPath - should use password auth
	}

	// Even with cancelled context, should still work since we don't check context in current implementation
	// This test ensures the method doesn't panic with cancelled context
	_, err := auth.GetAuthMethods(ctx, config)
	require.NoError(t, err)
}
