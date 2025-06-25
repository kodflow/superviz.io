package pkgmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/kodflow/superviz.io/internal/utils"
)

// APT implements the package manager for Debian-based distributions.
type APT struct{}

// NewAPT creates a new instance of APT manager.
//
// Returns:
//   - Pointer to an APT structure
func NewAPT() *APT {
	return &APT{}
}

// Name returns the package manager name.
//
// Returns:
//   - name: string package manager name ("apt")
func (m *APT) Name() string { return "apt" }

// Update returns the shell command to update the package index.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//
// Returns:
//   - cmd: string shell command string
//   - error if any
func (m *APT) Update(ctx context.Context) (string, error) {
	return "sudo apt update", nil
}

// Upgrade returns the shell command to update all installed packages.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//
// Returns:
//   - cmd: string shell command string
//   - error if any
func (m *APT) Upgrade(ctx context.Context) (string, error) {
	return "sudo apt upgrade -y", nil
}

// Install returns the shell command to install one or more packages.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - pkgs: ...string list of packages to install
//
// Returns:
//   - cmd: string shell command string
//   - err: error if no package is specified or if a package name is invalid
func (m *APT) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo apt install -y %s", strings.Join(pkgs, " ")), nil
}

// Remove returns the shell command to uninstall one or more packages.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - pkgs: ...string list of packages to uninstall
//
// Returns:
//   - cmd: string shell command string
//   - err: error if no package is specified or if a package name is invalid
func (m *APT) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo apt remove -y %s", strings.Join(pkgs, " ")), nil
}

// IsInstalled returns the shell command to check if a package is installed.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - pkg: string package name to check
//
// Returns:
//   - cmd: string shell command string
//   - Error if the package name is empty
func (m *APT) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if err := utils.ValidatePackageNames(pkg); err != nil {
		return "", err
	}
	return fmt.Sprintf("dpkg -s %s | grep Version", pkg), nil
}

// VersionCheck returns the shell commands to get the installed and available version of a package.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - pkg: string package name to check
//
// Returns:
//   - installed: string shell command for installed version
//   - available: string shell command for available version
//   - Error if the package name is empty
func (m *APT) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if err := utils.ValidatePackageNames(pkg); err != nil {
		return "", "", err
	}
	installed := fmt.Sprintf("dpkg-query -W -f='${Version}' %s", pkg)
	available := fmt.Sprintf("apt-cache policy %s | grep Candidate | awk '{print $2}'", pkg)
	return installed, available, nil
}
