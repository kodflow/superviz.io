package ssh

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// Mock implementations for testing

type mockDialer struct {
	connection Connection
	err        error
}

func (m *mockDialer) DialContext(ctx context.Context, network, addr string, config *ssh.ClientConfig) (Connection, error) {
	return m.connection, m.err
}

type mockConnection struct {
	session Session
	err     error
	closed  bool
}

func (m *mockConnection) NewSession() (Session, error) {
	return m.session, m.err
}

func (m *mockConnection) Close() error {
	m.closed = true
	return m.err
}

type mockSession struct {
	runErr   error
	runFunc  func(cmd string) error
	closeErr error
	closed   bool
}

func (m *mockSession) Run(cmd string) error {
	if m.runFunc != nil {
		return m.runFunc(cmd)
	}
	return m.runErr
}

func (m *mockSession) Close() error {
	m.closed = true
	return m.closeErr
}

type mockAuthenticator struct {
	authMethods []ssh.AuthMethod
	err         error
}

func (m *mockAuthenticator) GetAuthMethods(ctx context.Context, config *Config) ([]ssh.AuthMethod, error) {
	return m.authMethods, m.err
}

type mockHostKeyManager struct {
	callback ssh.HostKeyCallback
	err      error
}

func (m *mockHostKeyManager) GetHostKeyCallback(ctx context.Context, config *Config) (ssh.HostKeyCallback, error) {
	return m.callback, m.err
}

// Test NewClient

func TestNewClient_WithNilOptions(t *testing.T) {
	sshClient := NewClient(nil)
	require.NotNil(t, sshClient)

	// Should have default implementations
	concreteClient, ok := sshClient.(*client)
	require.True(t, ok)
	require.NotNil(t, concreteClient.authenticator)
	require.NotNil(t, concreteClient.hostKeyManager)
	require.NotNil(t, concreteClient.dialer)
}

func TestNewClient_WithOptions(t *testing.T) {
	mockAuth := &mockAuthenticator{}
	mockHKM := &mockHostKeyManager{}
	mockDialer := &mockDialer{}

	opts := &ClientOptions{
		Authenticator:  mockAuth,
		HostKeyManager: mockHKM,
		Dialer:         mockDialer,
	}

	sshClient := NewClient(opts)
	require.NotNil(t, sshClient)

	concreteClient, ok := sshClient.(*client)
	require.True(t, ok)
	require.Equal(t, mockAuth, concreteClient.authenticator)
	require.Equal(t, mockHKM, concreteClient.hostKeyManager)
	require.Equal(t, mockDialer, concreteClient.dialer)
}

func TestNewClient_WithPartialOptions(t *testing.T) {
	mockAuth := &mockAuthenticator{}

	opts := &ClientOptions{
		Authenticator: mockAuth,
		// HostKeyManager and Dialer are nil
	}

	sshClient := NewClient(opts)
	require.NotNil(t, sshClient)

	concreteClient, ok := sshClient.(*client)
	require.True(t, ok)
	require.Equal(t, mockAuth, concreteClient.authenticator)
	require.NotNil(t, concreteClient.hostKeyManager) // Should use default
	require.NotNil(t, concreteClient.dialer)         // Should use default
}

// Test sshClient.Connect

func TestClient_Connect_Success(t *testing.T) {
	mockAuth := &mockAuthenticator{
		authMethods: []ssh.AuthMethod{ssh.Password("test")},
		err:         nil,
	}
	mockHKM := &mockHostKeyManager{
		callback: ssh.InsecureIgnoreHostKey(),
		err:      nil,
	}
	mockConn := &mockConnection{}
	mockDialer := &mockDialer{
		connection: mockConn,
		err:        nil,
	}

	opts := &ClientOptions{
		Authenticator:  mockAuth,
		HostKeyManager: mockHKM,
		Dialer:         mockDialer,
	}

	sshClient := NewClient(opts).(*client)
	config := &Config{
		Host:    "example.com",
		User:    "testuser",
		Port:    22,
		Timeout: 30 * time.Second,
	}

	err := sshClient.Connect(context.Background(), config)
	require.NoError(t, err)
	require.Equal(t, config, sshClient.config)
	require.Equal(t, mockConn, sshClient.conn)
}

func TestClient_Connect_InvalidConfig(t *testing.T) {
	sshClient := NewClient(nil).(*client)
	config := &Config{
		// Invalid config - missing required fields
		Port: 22,
	}

	err := sshClient.Connect(context.Background(), config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "host cannot be empty")
}

func TestClient_Connect_HostKeyManagerError(t *testing.T) {
	mockAuth := &mockAuthenticator{}
	mockHKM := &mockHostKeyManager{
		err: errors.New("host key error"),
	}

	opts := &ClientOptions{
		Authenticator:  mockAuth,
		HostKeyManager: mockHKM,
	}

	sshClient := NewClient(opts).(*client)
	config := &Config{
		Host:    "example.com",
		User:    "testuser",
		Port:    22,
		Timeout: 30 * time.Second,
	}

	err := sshClient.Connect(context.Background(), config)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrHostKeyRejected))
}

func TestClient_Connect_AuthenticatorError(t *testing.T) {
	mockAuth := &mockAuthenticator{
		err: errors.New("auth error"),
	}
	mockHKM := &mockHostKeyManager{
		callback: ssh.InsecureIgnoreHostKey(),
	}

	opts := &ClientOptions{
		Authenticator:  mockAuth,
		HostKeyManager: mockHKM,
	}

	sshClient := NewClient(opts).(*client)
	config := &Config{
		Host:    "example.com",
		User:    "testuser",
		Port:    22,
		Timeout: 30 * time.Second,
	}

	err := sshClient.Connect(context.Background(), config)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrAuthFailed))
}

func TestClient_Connect_DialerError(t *testing.T) {
	mockAuth := &mockAuthenticator{
		authMethods: []ssh.AuthMethod{ssh.Password("test")},
	}
	mockHKM := &mockHostKeyManager{
		callback: ssh.InsecureIgnoreHostKey(),
	}
	mockDialer := &mockDialer{
		err: errors.New("connection failed"),
	}

	opts := &ClientOptions{
		Authenticator:  mockAuth,
		HostKeyManager: mockHKM,
		Dialer:         mockDialer,
	}

	sshClient := NewClient(opts).(*client)
	config := &Config{
		Host:    "example.com",
		User:    "testuser",
		Port:    22,
		Timeout: 30 * time.Second,
	}

	err := sshClient.Connect(context.Background(), config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "connection failed")
}

// Test sshClient.Execute

func TestClient_Execute_Success(t *testing.T) {
	mockSession := &mockSession{runErr: nil}
	mockConn := &mockConnection{session: mockSession}

	sshClient := &client{conn: mockConn}

	err := sshClient.Execute(context.Background(), "ls -la")
	require.NoError(t, err)
	require.True(t, mockSession.closed) // Session should be closed
}

func TestClient_Execute_NotConnected(t *testing.T) {
	sshClient := &client{} // No connection

	err := sshClient.Execute(context.Background(), "ls -la")
	require.Equal(t, ErrNotConnected, err)
}

func TestClient_Execute_SessionCreationError(t *testing.T) {
	mockConn := &mockConnection{
		err: errors.New("session creation failed"),
	}
	sshClient := &client{conn: mockConn}

	err := sshClient.Execute(context.Background(), "ls -la")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrSessionCreation))
}

func TestClient_Execute_CommandError(t *testing.T) {
	mockSession := &mockSession{
		runErr: errors.New("command failed"),
	}
	mockConn := &mockConnection{session: mockSession}
	sshClient := &client{conn: mockConn}

	err := sshClient.Execute(context.Background(), "invalid-command")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrCommandFailed))
	require.True(t, mockSession.closed) // Session should still be closed
}

func TestClient_Execute_ContextTimeout(t *testing.T) {
	// Create a session that blocks
	mockSession := &mockSession{
		runFunc: func(cmd string) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		},
	}

	mockConn := &mockConnection{session: mockSession}
	sshClient := &client{conn: mockConn}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := sshClient.Execute(ctx, "sleep 1")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrCommandTimeout))
}

// Test sshClient.Close

func TestClient_Close_Success(t *testing.T) {
	mockConn := &mockConnection{}
	sshClient := &client{conn: mockConn}

	err := sshClient.Close()
	require.NoError(t, err)
	require.True(t, mockConn.closed)
}

func TestClient_Close_NoConnection(t *testing.T) {
	sshClient := &client{} // No connection

	err := sshClient.Close()
	require.NoError(t, err) // Should not error
}

func TestClient_Close_ConnectionError(t *testing.T) {
	mockConn := &mockConnection{
		err: errors.New("close error"),
	}
	sshClient := &client{conn: mockConn}

	err := sshClient.Close()
	require.Error(t, err)
	require.Contains(t, err.Error(), "close error")
	require.True(t, mockConn.closed)
}

// Test session cleanup on error

func TestClient_Execute_SessionCleanupOnCloseError(t *testing.T) {
	mockSession := &mockSession{
		runErr:   nil,
		closeErr: errors.New("close error"),
	}
	mockConn := &mockConnection{session: mockSession}
	sshClient := &client{conn: mockConn}

	// Should not fail even if session close fails (warning to stderr)
	err := sshClient.Execute(context.Background(), "ls")
	require.NoError(t, err)
	require.True(t, mockSession.closed)
}
