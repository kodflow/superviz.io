package services_test

import (
	"bytes"
	"testing"

	"github.com/kodflow/superviz.io/internal/providers"
	"github.com/kodflow/superviz.io/internal/services"
	"github.com/stretchr/testify/require"
)

type mockVersionProvider struct {
	info providers.VersionInfo
}

func (m *mockVersionProvider) GetVersionInfo() providers.VersionInfo {
	return m.info
}

func newMockProvider() *mockVersionProvider {
	return &mockVersionProvider{
		info: providers.VersionInfo{
			Version:   "test-version",
			Commit:    "test-commit",
			BuiltAt:   "test-date",
			BuiltBy:   "test-user",
			GoVersion: "go1.21.0",
			OSArch:    "linux/amd64",
		},
	}
}

func TestNewVersionService_WithProvider(t *testing.T) {
	t.Parallel()

	mockProvider := newMockProvider()
	service := services.NewVersionService(mockProvider)

	require.NotNil(t, service)

	info := service.GetVersionInfo()
	require.Equal(t, "test-version", info.Version)
	require.Equal(t, "test-commit", info.Commit)
}

func TestNewVersionService_WithNilProvider(t *testing.T) {
	t.Parallel()

	service := services.NewVersionService(nil)

	require.NotNil(t, service)

	info := service.GetVersionInfo()
	require.Equal(t, "dev", info.Version) // Default values
	require.Equal(t, "none", info.Commit)
}

func TestVersionService_GetVersionInfo(t *testing.T) {
	t.Parallel()

	mockProvider := newMockProvider()
	service := services.NewVersionService(mockProvider)

	info := service.GetVersionInfo()

	require.Equal(t, "test-version", info.Version)
	require.Equal(t, "test-commit", info.Commit)
	require.Equal(t, "test-date", info.BuiltAt)
	require.Equal(t, "test-user", info.BuiltBy)
	require.Equal(t, "go1.21.0", info.GoVersion)
	require.Equal(t, "linux/amd64", info.OSArch)
}

func TestVersionService_DisplayVersion_Success(t *testing.T) {
	t.Parallel()

	mockProvider := newMockProvider()
	service := services.NewVersionService(mockProvider)

	var buf bytes.Buffer
	err := service.DisplayVersion(&buf)

	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "test-version")
	require.Contains(t, output, "test-commit")
	require.Contains(t, output, "Version:")
	require.Contains(t, output, "Commit:")
}

func TestVersionService_DisplayVersion_NilWriter(t *testing.T) {
	t.Parallel()

	service := services.NewVersionService(nil)

	err := service.DisplayVersion(nil)

	require.Error(t, err)
	require.Equal(t, services.ErrNilWriter, err)
}

func TestVersionService_DisplayVersionString(t *testing.T) {
	t.Parallel()

	mockProvider := newMockProvider()
	service := services.NewVersionService(mockProvider)

	output := service.DisplayVersionString()

	require.NotEmpty(t, output)
	require.Contains(t, output, "test-version")
	require.Contains(t, output, "test-commit")
	require.Contains(t, output, "Version:")
	require.Contains(t, output, "Commit:")
}

func TestVersionService_Performance(t *testing.T) {
	t.Parallel()

	service := services.NewVersionService(nil)

	// Multiple calls should be fast due to caching
	for i := 0; i < 100; i++ {
		info := service.GetVersionInfo()
		require.NotEmpty(t, info.Version)

		output := service.DisplayVersionString()
		require.NotEmpty(t, output)
	}
}
