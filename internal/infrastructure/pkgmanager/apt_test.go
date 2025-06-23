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

	cmd, err := m.Install(context.Background(), "htop", "curl")
	assert.NoError(t, err)
	assert.Equal(t, "sudo apt install -y htop curl", cmd)

	_, err = m.Install(context.Background())
	assert.Error(t, err)
}

func TestAPT_Remove(t *testing.T) {
	m := pkgmanager.NewAPT()

	cmd, err := m.Remove(context.Background(), "htop", "curl")
	assert.NoError(t, err)
	assert.Equal(t, "sudo apt remove -y htop curl", cmd)

	_, err = m.Remove(context.Background())
	assert.Error(t, err)
}

func TestAPT_IsInstalled(t *testing.T) {
	m := pkgmanager.NewAPT()

	cmd, err := m.IsInstalled(context.Background(), "htop")
	assert.NoError(t, err)
	assert.Equal(t, "dpkg -s htop | grep Version", cmd)

	_, err = m.IsInstalled(context.Background(), "")
	assert.Error(t, err)
}

func TestAPT_VersionCheck(t *testing.T) {
	m := pkgmanager.NewAPT()

	inst, avail, err := m.VersionCheck(context.Background(), "htop")
	assert.NoError(t, err)
	assert.Equal(t, "dpkg-query -W -f='${Version}' htop", inst)
	assert.Equal(t, "apt-cache policy htop | grep Candidate | awk '{print $2}'", avail)

	_, _, err = m.VersionCheck(context.Background(), "")
	assert.Error(t, err)
}
