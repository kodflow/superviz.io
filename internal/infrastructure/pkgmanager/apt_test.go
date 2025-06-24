package pkgmanager_test

import (
	"context"
	"testing"

	"github.com/kodflow/superviz.io/internal/infrastructure/pkgmanager"
	"github.com/stretchr/testify/assert"
)

func TestAPT_Name(t *testing.T) {
	m := pkgmanager.NewAPT()
	assert.Equal(t, "apt", m.Name())
}

func TestAPT_Update(t *testing.T) {
	cmd, err := pkgmanager.NewAPT().Update(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "sudo apt update", cmd)
}

func TestAPT_Upgrade(t *testing.T) {
	cmd, err := pkgmanager.NewAPT().Upgrade(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "sudo apt upgrade -y", cmd)
}

func TestAPT_Install(t *testing.T) {
	m := pkgmanager.NewAPT()

	// Test successful install
	cmd, err := m.Install(context.Background(), "htop", "curl")
	assert.NoError(t, err)
	assert.Equal(t, "sudo apt install -y htop curl", cmd)

	// Test no packages error
	_, err = m.Install(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "no package specified for install", err.Error())

	// Test dangerous package name validation
	_, err = m.Install(context.Background(), "htop; rm -rf /")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")

	// Test package name with command injection
	_, err = m.Install(context.Background(), "htop && evil_command")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")

	// Test package name with pipe
	_, err = m.Install(context.Background(), "htop | malicious")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")
}

func TestAPT_Remove(t *testing.T) {
	m := pkgmanager.NewAPT()

	// Test successful remove
	cmd, err := m.Remove(context.Background(), "htop", "curl")
	assert.NoError(t, err)
	assert.Equal(t, "sudo apt remove -y htop curl", cmd)

	// Test no packages error
	_, err = m.Remove(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "no package specified for removal", err.Error())

	// Test dangerous package name validation
	_, err = m.Remove(context.Background(), "htop; rm -rf /")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")

	// Test package name with variable expansion
	_, err = m.Remove(context.Background(), "htop$HOME")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")
}

func TestAPT_IsInstalled(t *testing.T) {
	m := pkgmanager.NewAPT()

	// Test successful check
	cmd, err := m.IsInstalled(context.Background(), "htop")
	assert.NoError(t, err)
	assert.Equal(t, "dpkg -s htop | grep Version", cmd)

	// Test empty package name
	_, err = m.IsInstalled(context.Background(), "")
	assert.Error(t, err)
	assert.Equal(t, "package name cannot be empty", err.Error())

	// Test dangerous package name validation
	_, err = m.IsInstalled(context.Background(), "htop`whoami`")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")
}

func TestAPT_VersionCheck(t *testing.T) {
	m := pkgmanager.NewAPT()

	// Test successful version check
	inst, avail, err := m.VersionCheck(context.Background(), "htop")
	assert.NoError(t, err)
	assert.Equal(t, "dpkg-query -W -f='${Version}' htop", inst)
	assert.Equal(t, "apt-cache policy htop | grep Candidate | awk '{print $2}'", avail)

	// Test empty package name
	_, _, err = m.VersionCheck(context.Background(), "")
	assert.Error(t, err)
	assert.Equal(t, "package name cannot be empty", err.Error())

	// Test dangerous package name validation
	_, _, err = m.VersionCheck(context.Background(), "htop $(ls)")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contains invalid characters")
}
