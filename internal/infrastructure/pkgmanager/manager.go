package pkgmanager

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var distroToPkgManager = map[string]string{
	"ubuntu":   "apt",
	"debian":   "apt",
	"alpine":   "apk",
	"centos":   "yum",
	"rhel":     "yum",
	"fedora":   "dnf",
	"arch":     "pacman",
	"sles":     "zypper",
	"opensuse": "zypper",
	"gentoo":   "emerge",
}

// Manager defines the interface for all detected package managers.
//
// Each implementation must provide methods for common package operations.
type Manager interface {
	// Name returns the name of the package manager (e.g. apt, apk...)
	Name() string
	// Update returns the command to execute for updating the package index.
	Update(ctx context.Context) (string, error)
	// Install returns the command to install one or more packages.
	Install(ctx context.Context, pkgs ...string) (string, error)
	// Remove returns the command to uninstall one or more packages.
	Remove(ctx context.Context, pkgs ...string) (string, error)
	// Upgrade returns the command to perform a global upgrade.
	Upgrade(ctx context.Context) (string, error)
	// IsInstalled returns the command to check if a package is installed and its current version.
	IsInstalled(ctx context.Context, pkg string) (string, error)
	// VersionCheck returns the command to compare installed version and repository version.
	VersionCheck(ctx context.Context, pkg string) (installedVersion string, availableVersion string, err error)
}

// Detect returns the appropriate package manager based on /etc/os-release,
// with a fallback to binaries present in PATH.
//
// Returns:
//   - Manager instance corresponding to the distribution
//   - Error if no manager is detected
func Detect() (Manager, error) {
	const osRelease = "/etc/os-release"

	if file, err := os.Open(osRelease); err == nil {
		defer file.Close() // nolint: errcheck

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "ID=") {
				distro := strings.Trim(strings.TrimPrefix(line, "ID="), `"`)
				if bin, ok := distroToPkgManager[distro]; ok {
					return DetectFromBin(bin)
				}
				break
			}
		}
	}

	// Fallback : inspecter les binaires dans le PATH
	for _, bin := range []string{"apt", "apk", "dnf", "yum", "pacman", "zypper", "emerge"} {
		if _, err := exec.LookPath(bin); err == nil {
			return DetectFromBin(bin)
		}
	}

	return nil, fmt.Errorf("unable to detect package manager")
}

// DetectFromBin returns a Manager instance based on the binary name.
//
// Parameters:
//   - bin: Binary name to detect (e.g. "apt", "yum")
//
// Returns:
//   - Manager instance
//   - Error if the binary is not supported
func DetectFromBin(bin string) (Manager, error) {
	switch bin {
	case "apt":
		return NewAPT(), nil
	case "apk":
		return NewAPK(), nil
	case "dnf":
		return NewDNF(), nil
	case "yum":
		return NewYUM(), nil
	case "pacman":
		return NewPACMAN(), nil
	case "zypper":
		return NewZYPPER(), nil
	case "emerge":
		return NewEMERGE(), nil
	default:
		return nil, fmt.Errorf("unsupported binary: %s", bin)
	}
}
