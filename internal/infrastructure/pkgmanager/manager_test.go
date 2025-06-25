package pkgmanager_test

import (
	"testing"

	"github.com/kodflow/superviz.io/internal/infrastructure/pkgmanager"
	"github.com/stretchr/testify/assert"
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
			assert.Contains(t, err.Error(), "unsupported binary")
		})
	}
}
