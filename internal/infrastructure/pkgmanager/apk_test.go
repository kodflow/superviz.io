package pkgmanager_test

import (
	"context"
	"testing"

	"github.com/kodflow/superviz.io/internal/infrastructure/pkgmanager"
	"github.com/stretchr/testify/assert"
)

func TestAPK_Name(t *testing.T) {
	m := pkgmanager.NewAPK()
	assert.Equal(t, "apk", m.Name())
}

func TestAPK_Update(t *testing.T) {
	cmd, err := pkgmanager.NewAPK().Update(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "sudo apk update", cmd)
}

func TestAPK_Upgrade(t *testing.T) {
	cmd, err := pkgmanager.NewAPK().Upgrade(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "sudo apk upgrade", cmd)
}

func TestAPK_Install(t *testing.T) {
	m := pkgmanager.NewAPK()

	cmd, err := m.Install(context.Background(), "htop", "curl")
	assert.NoError(t, err)
	assert.Equal(t, "sudo apk add htop curl", cmd)

	_, err = m.Install(context.Background())
	assert.Error(t, err)

	// Test security validation - dangerous package names should be rejected
	_, err = m.Install(context.Background(), "package; rm -rf /")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")

	_, err = m.Install(context.Background(), "package && malicious_command")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")

	_, err = m.Install(context.Background(), "package`command`")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")

	_, err = m.Install(context.Background(), "package$(command)")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")
}

func TestAPK_Remove(t *testing.T) {
	m := pkgmanager.NewAPK()

	cmd, err := m.Remove(context.Background(), "htop", "curl")
	assert.NoError(t, err)
	assert.Equal(t, "sudo apk del htop curl", cmd)

	_, err = m.Remove(context.Background())
	assert.Error(t, err)

	// Test security validation - dangerous package names should be rejected
	_, err = m.Remove(context.Background(), "package; rm -rf /")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")

	_, err = m.Remove(context.Background(), "package && malicious_command")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")

	_, err = m.Remove(context.Background(), "package`command`")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")

	_, err = m.Remove(context.Background(), "package$(command)")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")
}

func TestAPK_IsInstalled(t *testing.T) {
	m := pkgmanager.NewAPK()

	cmd, err := m.IsInstalled(context.Background(), "htop")
	assert.NoError(t, err)
	assert.Equal(t, "apk info -e htop", cmd)

	_, err = m.IsInstalled(context.Background(), "")
	assert.Error(t, err)

	// Test security validation - dangerous package names should be rejected
	_, err = m.IsInstalled(context.Background(), "package; rm -rf /")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")

	_, err = m.IsInstalled(context.Background(), "package && malicious_command")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")

	_, err = m.IsInstalled(context.Background(), "package`command`")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")

	_, err = m.IsInstalled(context.Background(), "package$(command)")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")
}

func TestAPK_VersionCheck(t *testing.T) {
	m := pkgmanager.NewAPK()

	inst, avail, err := m.VersionCheck(context.Background(), "htop")
	assert.NoError(t, err)
	assert.Equal(t, "apk info -v htop | cut -d'-' -f2-", inst)
	assert.Equal(t, "apk search -v htop | grep htop | cut -d'-' -f2-", avail)

	_, _, err = m.VersionCheck(context.Background(), "")
	assert.Error(t, err)

	// Test security validation - dangerous package names should be rejected
	_, _, err = m.VersionCheck(context.Background(), "package; rm -rf /")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")

	_, _, err = m.VersionCheck(context.Background(), "package && malicious_command")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")

	_, _, err = m.VersionCheck(context.Background(), "package`command`")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")

	_, _, err = m.VersionCheck(context.Background(), "package$(command)")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")
}
