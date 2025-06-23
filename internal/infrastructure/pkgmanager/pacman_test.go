package pkgmanager_test

import (
	"context"
	"testing"

	"github.com/kodflow/superviz.io/internal/infrastructure/pkgmanager"
	"github.com/stretchr/testify/assert"
)

func TestPACMAN_Name(t *testing.T) {
	m := pkgmanager.NewPACMAN()
	assert.Equal(t, "pacman", m.Name())
}

func TestPACMAN_Update(t *testing.T) {
	cmd, err := pkgmanager.NewPACMAN().Update(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "sudo pacman -Sy", cmd)
}

func TestPACMAN_Upgrade(t *testing.T) {
	cmd, err := pkgmanager.NewPACMAN().Upgrade(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "sudo pacman -Su --noconfirm", cmd)
}

func TestPACMAN_Install(t *testing.T) {
	m := pkgmanager.NewPACMAN()

	cmd, err := m.Install(context.Background(), "htop", "curl")
	assert.NoError(t, err)
	assert.Equal(t, "sudo pacman -S --noconfirm htop curl", cmd)

	_, err = m.Install(context.Background())
	assert.Error(t, err)
}

func TestPACMAN_Remove(t *testing.T) {
	m := pkgmanager.NewPACMAN()

	cmd, err := m.Remove(context.Background(), "htop", "curl")
	assert.NoError(t, err)
	assert.Equal(t, "sudo pacman -Rns --noconfirm htop curl", cmd)

	_, err = m.Remove(context.Background())
	assert.Error(t, err)
}

func TestPACMAN_IsInstalled(t *testing.T) {
	m := pkgmanager.NewPACMAN()

	cmd, err := m.IsInstalled(context.Background(), "htop")
	assert.NoError(t, err)
	assert.Equal(t, "pacman -Qi htop", cmd)

	_, err = m.IsInstalled(context.Background(), "")
	assert.Error(t, err)
}

func TestPACMAN_VersionCheck(t *testing.T) {
	m := pkgmanager.NewPACMAN()

	installed, available, err := m.VersionCheck(context.Background(), "htop")
	assert.NoError(t, err)
	assert.Equal(t, "pacman -Qi htop | grep Version | awk '{print $3}'", installed)
	assert.Equal(t, "pacman -Si htop | grep Version | awk '{print $3}'", available)

	_, _, err = m.VersionCheck(context.Background(), "")
	assert.Error(t, err)
}
