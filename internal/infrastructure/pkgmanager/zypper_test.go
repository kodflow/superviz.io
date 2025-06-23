package pkgmanager_test

import (
	"context"
	"testing"

	"github.com/kodflow/superviz.io/internal/infrastructure/pkgmanager"
	"github.com/stretchr/testify/assert"
)

func TestZYPPER_Name(t *testing.T) {
	m := pkgmanager.NewZYPPER()
	assert.Equal(t, "zypper", m.Name())
}

func TestZYPPER_Update(t *testing.T) {
	cmd, err := pkgmanager.NewZYPPER().Update(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "sudo zypper refresh", cmd)
}

func TestZYPPER_Upgrade(t *testing.T) {
	cmd, err := pkgmanager.NewZYPPER().Upgrade(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "sudo zypper update -y", cmd)
}

func TestZYPPER_Install(t *testing.T) {
	m := pkgmanager.NewZYPPER()

	cmd, err := m.Install(context.Background(), "htop", "curl")
	assert.NoError(t, err)
	assert.Equal(t, "sudo zypper install -y htop curl", cmd)

	_, err = m.Install(context.Background())
	assert.Error(t, err)
}

func TestZYPPER_Remove(t *testing.T) {
	m := pkgmanager.NewZYPPER()

	cmd, err := m.Remove(context.Background(), "htop", "curl")
	assert.NoError(t, err)
	assert.Equal(t, "sudo zypper remove -y htop curl", cmd)

	_, err = m.Remove(context.Background())
	assert.Error(t, err)
}

func TestZYPPER_IsInstalled(t *testing.T) {
	m := pkgmanager.NewZYPPER()

	cmd, err := m.IsInstalled(context.Background(), "htop")
	assert.NoError(t, err)
	assert.Equal(t, "zypper se --installed-only htop", cmd)

	_, err = m.IsInstalled(context.Background(), "")
	assert.Error(t, err)
}

func TestZYPPER_VersionCheck(t *testing.T) {
	m := pkgmanager.NewZYPPER()

	inst, avail, err := m.VersionCheck(context.Background(), "htop")
	assert.NoError(t, err)
	assert.Equal(t, "zypper info htop | grep Version | head -1 | awk '{print $3}'", inst)
	assert.Equal(t, "zypper info htop | grep Version | tail -1 | awk '{print $3}'", avail)

	_, _, err = m.VersionCheck(context.Background(), "")
	assert.Error(t, err)
}
