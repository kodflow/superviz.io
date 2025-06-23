package pkgmanager_test

import (
	"context"
	"testing"

	"github.com/kodflow/superviz.io/internal/infrastructure/pkgmanager"
	"github.com/stretchr/testify/assert"
)

func TestDNF_Name(t *testing.T) {
	m := pkgmanager.NewDNF()
	assert.Equal(t, "dnf", m.Name())
}

func TestDNF_Update(t *testing.T) {
	cmd, err := pkgmanager.NewDNF().Update(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "sudo dnf check-update", cmd)
}

func TestDNF_Upgrade(t *testing.T) {
	cmd, err := pkgmanager.NewDNF().Upgrade(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "sudo dnf upgrade -y", cmd)
}

func TestDNF_Install(t *testing.T) {
	m := pkgmanager.NewDNF()

	cmd, err := m.Install(context.Background(), "htop", "curl")
	assert.NoError(t, err)
	assert.Equal(t, "sudo dnf install -y htop curl", cmd)

	_, err = m.Install(context.Background())
	assert.Error(t, err)
}

func TestDNF_Remove(t *testing.T) {
	m := pkgmanager.NewDNF()

	cmd, err := m.Remove(context.Background(), "htop", "curl")
	assert.NoError(t, err)
	assert.Equal(t, "sudo dnf remove -y htop curl", cmd)

	_, err = m.Remove(context.Background())
	assert.Error(t, err)
}

func TestDNF_IsInstalled(t *testing.T) {
	m := pkgmanager.NewDNF()

	cmd, err := m.IsInstalled(context.Background(), "htop")
	assert.NoError(t, err)
	assert.Equal(t, "dnf list installed htop", cmd)

	_, err = m.IsInstalled(context.Background(), "")
	assert.Error(t, err)
}

func TestDNF_VersionCheck(t *testing.T) {
	m := pkgmanager.NewDNF()

	inst, avail, err := m.VersionCheck(context.Background(), "htop")
	assert.NoError(t, err)
	assert.Equal(t, "dnf info htop | grep Version", inst)
	assert.Equal(t, "dnf --showduplicates list htop | grep -v Installed | awk '{print $2}'", avail)

	_, _, err = m.VersionCheck(context.Background(), "")
	assert.Error(t, err)
}
