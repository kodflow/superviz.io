package pkgmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/kodflow/superviz.io/internal/utils"
)

// YUM implements the package manager for CentOS, RHEL and derivatives.
type YUM struct{}

// NewYUM creates a new instance of YUM manager.
//
//	mgr := NewYUM()
//	fmt.Println(mgr.Name())
//
// Parameters:
//   - None
//
// Returns:
//   - mgr: *YUM pointer to a YUM structure
func NewYUM() *YUM {
	return &YUM{}
}

// Name returns the package manager name.
//
//	mgr := NewYUM()
//	name := mgr.Name()
//	fmt.Println(name) // "yum"
//
// Parameters:
//   - None
//
// Returns:
//   - name: string package manager name ("yum")
func (m *YUM) Name() string { return "yum" }

// Update returns the shell command to update the package index.
//
//	mgr := NewYUM()
//	cmd, err := mgr.Update(ctx)
//	fmt.Println(cmd) // "sudo yum check-update"
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//
// Returns:
//   - cmd: string shell command string
//   - err: error if any
func (m *YUM) Update(ctx context.Context) (string, error) {
	return "sudo yum check-update", nil
}

// Upgrade returns the shell command to update all installed packages.
//
//	mgr := NewYUM()
//	cmd, err := mgr.Upgrade(ctx)
//	fmt.Println(cmd) // "sudo yum upgrade -y"
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//
// Returns:
//   - cmd: string shell command string
//   - err: error if any
func (m *YUM) Upgrade(ctx context.Context) (string, error) {
	return "sudo yum upgrade -y", nil
}

// Install returns the shell command to install one or more packages.
//
//	mgr := NewYUM()
//	cmd, err := mgr.Install(ctx, "vim", "git")
//	fmt.Println(cmd) // "sudo yum install -y vim git"
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - pkgs: ...string list of packages to install
//
// Returns:
//   - cmd: string shell command string
//   - err: error if no package is specified
func (m *YUM) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo yum install -y %s", strings.Join(pkgs, " ")), nil
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
func (m *YUM) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo yum remove -y %s", strings.Join(pkgs, " ")), nil
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
func (m *YUM) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if err := utils.ValidatePackageNames(pkg); err != nil {
		return "", err
	}
	return fmt.Sprintf("yum list installed %s", pkg), nil
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
func (m *YUM) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if err := utils.ValidatePackageNames(pkg); err != nil {
		return "", "", err
	}
	installed := fmt.Sprintf("yum info %s | grep Version", pkg)
	available := fmt.Sprintf("yum --showduplicates list %s | grep -v Installed | awk '{print $2}'", pkg)
	return installed, available, nil
}
