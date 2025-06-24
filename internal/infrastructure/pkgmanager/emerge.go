package pkgmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/kodflow/superviz.io/internal/utils"
)

// EMERGE implements the package manager for Gentoo.
type EMERGE struct{}

// NewEMERGE creates a new instance of EMERGE manager.
//
//	mgr := NewEMERGE()
//	fmt.Println(mgr.Name())
//
// Parameters:
//   - None
//
// Returns:
//   - mgr: *EMERGE pointer to an EMERGE structure
func NewEMERGE() *EMERGE {
	return &EMERGE{}
}

// Name returns the package manager name.
//
//	mgr := NewEMERGE()
//	name := mgr.Name()
//	fmt.Println(name) // "emerge"
//
// Parameters:
//   - None
//
// Returns:
//   - name: string package manager name ("emerge")
func (m *EMERGE) Name() string { return "emerge" }

// Update returns the shell command to synchronize the package index.
//
//	mgr := NewEMERGE()
//	cmd, err := mgr.Update(ctx)
//	fmt.Println(cmd) // "sudo emerge --sync"
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//
// Returns:
//   - cmd: string shell command string
//   - err: error if any
func (m *EMERGE) Update(ctx context.Context) (string, error) {
	return "sudo emerge --sync", nil
}

// Upgrade returns the shell command to update all installed packages.
//
//	mgr := NewEMERGE()
//	cmd, err := mgr.Upgrade(ctx)
//	fmt.Println(cmd) // "sudo emerge -uDN @world"
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//
// Returns:
//   - cmd: string shell command string
//   - err: error if any
func (m *EMERGE) Upgrade(ctx context.Context) (string, error) {
	return "sudo emerge -uDN @world", nil
}

// Install returns the shell command to install one or more packages.
//
//	mgr := NewEMERGE()
//	cmd, err := mgr.Install(ctx, "vim", "git")
//	fmt.Println(cmd) // "sudo emerge vim git"
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - pkgs: ...string list of packages to install
//
// Returns:
//   - cmd: string shell command string
//   - err: error if no package is specified
func (m *EMERGE) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo emerge %s", strings.Join(pkgs, " ")), nil
}

// Remove returns the shell command to uninstall one or more packages.
//
//	mgr := NewEMERGE()
//	cmd, err := mgr.Remove(ctx, "vim", "git")
//	fmt.Println(cmd) // "sudo emerge -C vim git"
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - pkgs: ...string list of packages to uninstall
//
// Returns:
//   - cmd: string shell command string
//   - err: error if no package is specified
func (m *EMERGE) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo emerge -C %s", strings.Join(pkgs, " ")), nil
}

// IsInstalled returns the shell command to check if a package is installed.
//
//	mgr := NewEMERGE()
//	cmd, err := mgr.IsInstalled(ctx, "vim")
//	fmt.Println(cmd) // "equery list vim"
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - pkg: string package name to check
//
// Returns:
//   - cmd: string shell command string
//   - err: error if package name is empty
func (m *EMERGE) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if err := utils.ValidatePackageNames(pkg); err != nil {
		return "", err
	}
	return fmt.Sprintf("equery list %s", pkg), nil
}

// VersionCheck returns the shell commands to get installed and available version of a package.
//
//	mgr := NewEMERGE()
//	inst, avail, err := mgr.VersionCheck(ctx, "vim")
//	fmt.Println(inst)  // "equery list vim | awk '{print $2}'"
//	fmt.Println(avail) // "emerge -p vim | grep '\\[ebuild' | awk '{print $4}'"
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - pkg: string package name to check
//
// Returns:
//   - installed: string shell command for installed version
//   - available: string shell command for available version
//   - err: error if package name is empty
func (m *EMERGE) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if err := utils.ValidatePackageNames(pkg); err != nil {
		return "", "", err
	}
	installed := fmt.Sprintf("equery list %s | awk '{print $2}'", pkg)
	available := fmt.Sprintf("emerge -p %s | grep '\\[ebuild' | awk '{print $4}'", pkg)

	return installed, available, nil
}
