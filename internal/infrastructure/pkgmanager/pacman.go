package pkgmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/kodflow/superviz.io/internal/utils"
)

// PACMAN implements the package manager for Arch Linux and derivatives.
type PACMAN struct{}

// NewPACMAN creates a new instance of PACMAN manager.
//
// Returns:
//   - Pointer to a PACMAN structure
func NewPACMAN() *PACMAN {
	return &PACMAN{}
}

// Name returns the package manager name.
//
// Returns:
//   - name: string package manager name ("pacman")
func (m *PACMAN) Name() string { return "pacman" }

// Update returns the shell command to update the package index.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//
// Returns:
//   - cmd: string shell command string
//   - error if any
func (m *PACMAN) Update(ctx context.Context) (string, error) {
	return "sudo pacman -Sy", nil
}

// Upgrade returns the shell command to update all installed packages.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//
// Returns:
//   - cmd: string shell command string
//   - error if any
func (m *PACMAN) Upgrade(ctx context.Context) (string, error) {
	return "sudo pacman -Su --noconfirm", nil
}

// Install returns the shell command to install one or more packages.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - pkgs: ...string list of packages to install
//
// Returns:
//   - cmd: string shell command string
//   - err: error if no package is specified
func (m *PACMAN) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo pacman -S --noconfirm %s", strings.Join(pkgs, " ")), nil
}

// Remove returns the shell command to uninstall one or more packages.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - pkgs: ...string list of packages to uninstall
//
// Returns:
//   - cmd: string shell command string
//   - err: error if no package is specified
func (m *PACMAN) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo pacman -Rns --noconfirm %s", strings.Join(pkgs, " ")), nil
}

// IsInstalled returns the shell command to check if a package is installed.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - pkg: string package name to check
//
// Returns:
//   - cmd: string shell command string
//   - err: error if package name is empty
func (m *PACMAN) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if err := utils.ValidatePackageNames(pkg); err != nil {
		return "", err
	}
	return fmt.Sprintf("pacman -Qi %s", pkg), nil
}

// VersionCheck returns the shell commands to get installed and available version of a package.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - pkg: string package name to check
//
// Returns:
//   - installed: string shell command for installed version
//   - available: string shell command for available version
//   - err: error if package name is empty
func (m *PACMAN) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if err := utils.ValidatePackageNames(pkg); err != nil {
		return "", "", err
	}
	installed := fmt.Sprintf("pacman -Qi %s | grep Version | awk '{print $3}'", pkg)
	available := fmt.Sprintf("pacman -Si %s | grep Version | awk '{print $3}'", pkg)

	return installed, available, nil
}
