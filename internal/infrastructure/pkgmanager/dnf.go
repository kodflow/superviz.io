package pkgmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/kodflow/superviz.io/internal/utils"
)

// DNF implements the package manager for Fedora, RHEL and derivatives.
type DNF struct{}

// NewDNF creates a new instance of DNF manager.
//
// Returns:
//   - Pointer to a DNF structure
func NewDNF() *DNF {
	return &DNF{}
}

// Name returns the package manager name.
//
// Returns:
//   - name: string package manager name ("dnf")
func (m *DNF) Name() string { return "dnf" }

// Update returns the shell command to update the package index.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//
// Returns:
//   - cmd: string shell command string
//   - error if any
func (m *DNF) Update(ctx context.Context) (string, error) {
	return "sudo dnf check-update", nil
}

// Upgrade returns the shell command to update all installed packages.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//
// Returns:
//   - cmd: string shell command string
//   - error if any
func (m *DNF) Upgrade(ctx context.Context) (string, error) {
	return "sudo dnf upgrade -y", nil
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
func (m *DNF) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo dnf install -y %s", strings.Join(pkgs, " ")), nil
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
func (m *DNF) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo dnf remove -y %s", strings.Join(pkgs, " ")), nil
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
func (m *DNF) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if err := utils.ValidatePackageNames(pkg); err != nil {
		return "", err
	}
	return fmt.Sprintf("dnf list installed %s", pkg), nil
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
func (m *DNF) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if err := utils.ValidatePackageNames(pkg); err != nil {
		return "", "", err
	}
	installed := fmt.Sprintf("dnf info %s | grep Version", pkg)
	available := fmt.Sprintf("dnf --showduplicates list %s | grep -v Installed | awk '{print $2}'", pkg)

	return installed, available, nil
}
