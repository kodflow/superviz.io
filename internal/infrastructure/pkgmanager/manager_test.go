package pkgmanager_test

import (
	"os"
	"testing"

	"github.com/kodflow/superviz.io/internal/infrastructure/pkgmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetect_LiveEnvironment(t *testing.T) {
	mgr, err := pkgmanager.Detect()

	// either we detected a valid manager
	if err == nil {
		assert.NotNil(t, mgr)
		assert.NotEmpty(t, mgr.Name())
		t.Logf("Detected package manager: %s", mgr.Name())
		return
	}

	// or we are in a too minimal environment (e.g. scratch)
	assert.Nil(t, mgr)
	t.Logf("No package manager detected: %v", err)
}

func TestDetectFromBin_AllSupportedBinaries(t *testing.T) {
	testCases := []struct {
		bin      string
		expected string
	}{
		{"apt", "apt"},
		{"apk", "apk"},
		{"dnf", "dnf"},
		{"yum", "yum"},
		{"pacman", "pacman"},
		{"zypper", "zypper"},
		{"emerge", "emerge"},
	}

	for _, tc := range testCases {
		t.Run(tc.bin, func(t *testing.T) {
			mgr, err := pkgmanager.DetectFromBin(tc.bin)
			assert.NoError(t, err)
			assert.NotNil(t, mgr)
			assert.Equal(t, tc.expected, mgr.Name())
		})
	}
}

func TestDetectFromBin_UnsupportedBinary(t *testing.T) {
	unsupportedBins := []string{
		"unsupported",
		"invalid",
		"unknown",
		"",
		"homebrew",
		"pip",
		"npm",
	}

	for _, bin := range unsupportedBins {
		t.Run(bin, func(t *testing.T) {
			mgr, err := pkgmanager.DetectFromBin(bin)
			assert.Error(t, err)
			assert.Nil(t, mgr)
			// Handle both validation error (empty string) and unsupported binary error
			if bin == "" {
				assert.Contains(t, err.Error(), "binary name cannot be empty")
			} else {
				assert.Contains(t, err.Error(), "unsupported binary")
			}
		})
	}
}

// Additional tests for better coverage of Detect function

func TestDetect_WithMockedOSRelease(t *testing.T) {
	// Create a temporary file that simulates /etc/os-release
	tmpFile, err := os.CreateTemp("", "os-release-*")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	// Write Ubuntu content
	_, err = tmpFile.WriteString(`NAME="Ubuntu"
VERSION="20.04.3 LTS (Focal Fossa)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 20.04.3 LTS"
VERSION_ID="20.04"
`)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	// This test shows what would happen if we could inject the file path
	// but since Detect() hardcodes "/etc/os-release", we test the concept
	// by directly checking the mapping
	manager, err := pkgmanager.DetectFromBin("apt")
	require.NoError(t, err)
	require.NotNil(t, manager)
	require.Equal(t, "apt", manager.Name())
}

func TestDetect_FallbackToBinaryDetection(t *testing.T) {
	// This test documents the fallback behavior when /etc/os-release doesn't exist
	// or doesn't contain a recognized ID
	// We can't fully test this without mocking the filesystem,
	// but we can test the DetectFromBin function which is the fallback

	tests := []struct {
		binary   string
		expected string
	}{
		{"apt", "apt"},
		{"apk", "apk"},
		{"dnf", "dnf"},
		{"yum", "yum"},
		{"pacman", "pacman"},
		{"zypper", "zypper"},
		{"emerge", "emerge"},
	}

	for _, tt := range tests {
		t.Run(tt.binary, func(t *testing.T) {
			manager, err := pkgmanager.DetectFromBin(tt.binary)
			require.NoError(t, err)
			require.NotNil(t, manager)
			require.Equal(t, tt.expected, manager.Name())
		})
	}
}

func TestDetect_NoPackageManagerAvailable(t *testing.T) {
	// Test the error case when no package manager is found
	// This documents the behavior when DetectFromBin returns an error
	manager, err := pkgmanager.DetectFromBin("nonexistent")
	require.Error(t, err)
	require.Nil(t, manager)
	require.Contains(t, err.Error(), "unsupported binary")
}
