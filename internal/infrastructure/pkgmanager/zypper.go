package pkgmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/kodflow/superviz.io/internal/utils"
)

// ZYPPER implements the package manager for openSUSE and SLES.
type ZYPPER struct{}

// NewZYPPER creates a new instance of ZYPPER manager.
//
// Returns:
//   - Pointer to a ZYPPER structure
func NewZYPPER() *ZYPPER {
	return &ZYPPER{}
}

// Name returns the package manager name.
//
// Returns:
//   - name: string package manager name ("zypper")
func (m *ZYPPER) Name() string { return "zypper" }

// Update returns the shell command to refresh the package index.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//
// Returns:
//   - cmd: string shell command string
//   - error if any
func (m *ZYPPER) Update(ctx context.Context) (string, error) {
	return "sudo zypper refresh", nil
}

// Upgrade returns the shell command to update all installed packages.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//
// Returns:
//   - cmd: string shell command string
//   - error if any
func (m *ZYPPER) Upgrade(ctx context.Context) (string, error) {
	return "sudo zypper update -y", nil
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
func (m *ZYPPER) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo zypper install -y %s", strings.Join(pkgs, " ")), nil
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
func (m *ZYPPER) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo zypper remove -y %s", strings.Join(pkgs, " ")), nil
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
func (m *ZYPPER) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if err := utils.ValidatePackageNames(pkg); err != nil {
		return "", err
	}
	return fmt.Sprintf("zypper se --installed-only %s", pkg), nil
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
func (m *ZYPPER) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if err := utils.ValidatePackageNames(pkg); err != nil {
		return "", "", err
	}
	installed := fmt.Sprintf("zypper info %s | grep Version | head -1 | awk '{print $3}'", pkg)
	available := fmt.Sprintf("zypper info %s | grep Version | tail -1 | awk '{print $3}'", pkg)
	return installed, available, nil
}
