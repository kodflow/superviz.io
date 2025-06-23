package pkgmanager_test

import (
	"context"
	"testing"

	"github.com/kodflow/superviz.io/internal/infrastructure/pkgmanager"
	"github.com/stretchr/testify/assert"
)

func TestYUM_Name(t *testing.T) {
	m := pkgmanager.NewYUM()
	assert.Equal(t, "yum", m.Name())
}

func TestYUM_Update(t *testing.T) {
	cmd, err := pkgmanager.NewYUM().Update(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "sudo yum check-update", cmd)
}

func TestYUM_Upgrade(t *testing.T) {
	cmd, err := pkgmanager.NewYUM().Upgrade(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "sudo yum upgrade -y", cmd)
}

func TestYUM_Install(t *testing.T) {
	m := pkgmanager.NewYUM()

	cmd, err := m.Install(context.Background(), "htop", "curl")
	assert.NoError(t, err)
	assert.Equal(t, "sudo yum install -y htop curl", cmd)

	_, err = m.Install(context.Background())
	assert.Error(t, err)
}

func TestYUM_Remove(t *testing.T) {
	m := pkgmanager.NewYUM()

	cmd, err := m.Remove(context.Background(), "htop", "curl")
	assert.NoError(t, err)
	assert.Equal(t, "sudo yum remove -y htop curl", cmd)

	_, err = m.Remove(context.Background())
	assert.Error(t, err)
}

func TestYUM_IsInstalled(t *testing.T) {
	m := pkgmanager.NewYUM()

	cmd, err := m.IsInstalled(context.Background(), "htop")
	assert.NoError(t, err)
	assert.Equal(t, "yum list installed htop", cmd)

	_, err = m.IsInstalled(context.Background(), "")
	assert.Error(t, err)
}

func TestYUM_VersionCheck(t *testing.T) {
	m := pkgmanager.NewYUM()

	inst, avail, err := m.VersionCheck(context.Background(), "htop")
	assert.NoError(t, err)
	assert.Equal(t, "yum info htop | grep Version", inst)
	assert.Equal(t, "yum --showduplicates list htop | grep -v Installed | awk '{print $2}'", avail)

	_, _, err = m.VersionCheck(context.Background(), "")
	assert.Error(t, err)
}
