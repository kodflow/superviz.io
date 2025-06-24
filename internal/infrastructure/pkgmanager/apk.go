package pkgmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/kodflow/superviz.io/internal/utils"
)

// APK implements the package manager for Alpine Linux.
type APK struct{}

// NewAPK creates a new instance of APK manager.
//
// Returns:
//   - Pointer to an APK structure
func NewAPK() *APK {
	return &APK{}
}

// Name returns the package manager name.
//
// Returns:
//   - name: string package manager name ("apk")
func (m *APK) Name() string { return "apk" }

// Update returns the shell command to update the package index.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//
// Returns:
//   - cmd: string shell command string
//   - error if any
func (m *APK) Update(ctx context.Context) (string, error) {
	return "sudo apk update", nil
}

// Upgrade returns the shell command to update all installed packages.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//
// Returns:
//   - cmd: string shell command string
//   - error if any
func (m *APK) Upgrade(ctx context.Context) (string, error) {
	return "sudo apk upgrade", nil
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
func (m *APK) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo apk add %s", strings.Join(pkgs, " ")), nil
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
func (m *APK) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo apk del %s", strings.Join(pkgs, " ")), nil
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
func (m *APK) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if err := utils.ValidatePackageNames(pkg); err != nil {
		return "", err
	}
	return fmt.Sprintf("apk info -e %s", pkg), nil
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
func (m *APK) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if err := utils.ValidatePackageNames(pkg); err != nil {
		return "", "", err
	}
	installed := fmt.Sprintf("apk info -v %s | cut -d'-' -f2-", pkg)
	available := fmt.Sprintf("apk search -v %s | grep %s | cut -d'-' -f2-", pkg, pkg)

	return installed, available, nil
}
