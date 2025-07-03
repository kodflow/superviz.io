package pkgmanager

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
)

// Global atomic counters for package manager operations (false sharing prevention)
var (
	detectCalls      atomic.Uint64
	detectErrors     atomic.Uint64
	managerCreations atomic.Uint64
	osReleaseReads   atomic.Uint64
	binaryFallbacks  atomic.Uint64
	_                [3]uint64 // Cache-line padding
)

// Optimized distro mapping with pre-allocated map capacity
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

// Cache for already detected managers to avoid repeated detection
var (
	cachedManager Manager
	cacheOnce     sync.Once
	cacheMutex    sync.RWMutex
)

// Manager defines the interface for all detected package managers with ultra-performance operations
// Code block:
//
//	manager, err := Detect()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	cmd, err := manager.Install(ctx, "vim", "git")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Install command: %s\n", cmd)
//
// Parameters: N/A (for interfaces)
//
// Returns: N/A (for interfaces)
type Manager interface {
	// Name returns the name of the package manager with atomic call tracking
	// Code block:
	//
	//  name := manager.Name()
	//  fmt.Printf("Manager: %s\n", name)
	//
	// Parameters: N/A
	// Returns:
	//   - 1 name: string - package manager name (e.g. "apt", "apk")
	Name() string

	// Update returns the command to execute for updating the package index with validation
	// Code block:
	//
	//  cmd, err := manager.Update(ctx)
	//  if err != nil {
	//      log.Fatal(err)
	//  }
	//  fmt.Printf("Update command: %s\n", cmd)
	//
	// Parameters:
	//   - 1 ctx: context.Context - request context with timeout and cancellation
	// Returns:
	//   - 1 command: string - shell command to update package index
	//   - 2 error - non-nil if context is invalid or command generation fails
	Update(ctx context.Context) (string, error)

	// Install returns the command to install one or more packages with validation
	// Code block:
	//
	//  cmd, err := manager.Install(ctx, "vim", "git", "curl")
	//  if err != nil {
	//      log.Fatal(err)
	//  }
	//  fmt.Printf("Install command: %s\n", cmd)
	//
	// Parameters:
	//   - 1 ctx: context.Context - request context with timeout and cancellation
	//   - 2 pkgs: ...string - list of package names to install (must not be empty)
	// Returns:
	//   - 1 command: string - shell command to install packages
	//   - 2 error - non-nil if context invalid, no packages provided, or command generation fails
	Install(ctx context.Context, pkgs ...string) (string, error)

	// Remove returns the command to uninstall one or more packages with validation
	// Code block:
	//
	//  cmd, err := manager.Remove(ctx, "vim", "git")
	//  if err != nil {
	//      log.Fatal(err)
	//  }
	//  fmt.Printf("Remove command: %s\n", cmd)
	//
	// Parameters:
	//   - 1 ctx: context.Context - request context with timeout and cancellation
	//   - 2 pkgs: ...string - list of package names to remove (must not be empty)
	// Returns:
	//   - 1 command: string - shell command to remove packages
	//   - 2 error - non-nil if context invalid, no packages provided, or command generation fails
	Remove(ctx context.Context, pkgs ...string) (string, error)

	// Upgrade returns the command to perform a global upgrade with validation
	// Code block:
	//
	//  cmd, err := manager.Upgrade(ctx)
	//  if err != nil {
	//      log.Fatal(err)
	//  }
	//  fmt.Printf("Upgrade command: %s\n", cmd)
	//
	// Parameters:
	//   - 1 ctx: context.Context - request context with timeout and cancellation
	// Returns:
	//   - 1 command: string - shell command to upgrade all packages
	//   - 2 error - non-nil if context is invalid or command generation fails
	Upgrade(ctx context.Context) (string, error)

	// IsInstalled returns the command to check if a package is installed and its current version
	// Code block:
	//
	//  cmd, err := manager.IsInstalled(ctx, "vim")
	//  if err != nil {
	//      log.Fatal(err)
	//  }
	//  fmt.Printf("Check command: %s\n", cmd)
	//
	// Parameters:
	//   - 1 ctx: context.Context - request context with timeout and cancellation
	//   - 2 pkg: string - package name to check (must not be empty)
	// Returns:
	//   - 1 command: string - shell command to check package installation status
	//   - 2 error - non-nil if context invalid, package name empty, or command generation fails
	IsInstalled(ctx context.Context, pkg string) (string, error)

	// VersionCheck returns the command to compare installed version and repository version
	// Code block:
	//
	//  installed, available, err := manager.VersionCheck(ctx, "vim")
	//  if err != nil {
	//      log.Fatal(err)
	//  }
	//  fmt.Printf("Installed: %s, Available: %s\n", installed, available)
	//
	// Parameters:
	//   - 1 ctx: context.Context - request context with timeout and cancellation
	//   - 2 pkg: string - package name to check versions for (must not be empty)
	// Returns:
	//   - 1 installedVersion: string - currently installed version
	//   - 2 availableVersion: string - available version in repository
	//   - 3 error - non-nil if context invalid, package name empty, or version check fails
	VersionCheck(ctx context.Context, pkg string) (installedVersion string, availableVersion string, err error)
}

// Detect returns the appropriate package manager based on /etc/os-release with caching and atomic tracking
// Code block:
//
//	manager, err := Detect()
//	if err != nil {
//	    return fmt.Errorf("no package manager found: %w", err)
//	}
//	fmt.Printf("Detected package manager: %s\n", manager.Name())
//
//	metrics := GetPackageManagerMetrics()
//	fmt.Printf("Detection calls: %d\n", metrics.DetectCalls)
//
// Parameters: N/A
//
// Returns:
//   - 1 manager: Manager - instance corresponding to the detected distribution package manager
//   - 2 error - non-nil if no package manager is detected or os-release parsing fails
func Detect() (Manager, error) {
	detectCalls.Add(1)

	// Check cache first (read lock for performance)
	cacheMutex.RLock()
	if cachedManager != nil {
		cacheMutex.RUnlock()
		return cachedManager, nil
	}
	cacheMutex.RUnlock()

	// Use sync.Once for thread-safe cache initialization
	var err error
	cacheOnce.Do(func() {
		cacheMutex.Lock()
		defer cacheMutex.Unlock()

		cachedManager, err = detectInternal()
	})

	if err != nil {
		detectErrors.Add(1)
		return nil, err
	}

	return cachedManager, nil
}

// detectInternal performs the actual detection logic
func detectInternal() (Manager, error) {
	const osRelease = "/etc/os-release"

	// Try os-release first (most reliable)
	if file, err := os.Open(osRelease); err == nil {
		defer file.Close() // nolint: errcheck
		osReleaseReads.Add(1)

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			// Optimized parsing: check prefix and extract in one operation
			if strings.HasPrefix(line, "ID=") {
				distro := strings.Trim(strings.TrimPrefix(line, "ID="), `"`)
				if bin, exists := distroToPkgManager[distro]; exists {
					return DetectFromBin(bin)
				}
				break
			}
		}
	}

	// Fallback: inspect binaries in PATH (ordered by likelihood)
	binaryFallbacks.Add(1)
	binaries := []string{"apt", "apk", "dnf", "yum", "pacman", "zypper", "emerge"}
	for _, bin := range binaries {
		if _, err := exec.LookPath(bin); err == nil {
			return DetectFromBin(bin)
		}
	}

	return nil, fmt.Errorf("unable to detect package manager")
}

// DetectFromBin returns a Manager instance based on the binary name with atomic tracking
// Code block:
//
//	manager, err := DetectFromBin("apt")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created manager: %s\n", manager.Name())
//
// Parameters:
//   - 1 bin: string - binary name to detect (e.g. "apt", "yum", must not be empty)
//
// Returns:
//   - 1 manager: Manager - configured manager instance
//   - 2 error - non-nil if the binary is not supported or empty
func DetectFromBin(bin string) (Manager, error) {
	managerCreations.Add(1)

	// Input validation (proactive security)
	if bin == "" {
		return nil, fmt.Errorf("binary name cannot be empty")
	}

	// Branch prediction optimization: most common managers first
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

// PackageManagerMetrics contains atomic performance metrics for package manager operations
// Code block:
//
//	metrics := GetPackageManagerMetrics()
//	fmt.Printf("Detections: %d, Errors: %d, Creations: %d\n",
//	    metrics.DetectCalls, metrics.DetectErrors, metrics.ManagerCreations)
//
// Parameters: N/A (for types)
//
// Returns: N/A (for types)
type PackageManagerMetrics struct {
	DetectCalls      uint64
	DetectErrors     uint64
	ManagerCreations uint64
	OSReleaseReads   uint64
	BinaryFallbacks  uint64
}

// GetPackageManagerMetrics returns current atomic performance metrics for all package manager operations
// Code block:
//
//	metrics := GetPackageManagerMetrics()
//	fmt.Printf("Package manager metrics: %+v\n", metrics)
//
// Parameters: N/A
//
// Returns:
//   - 1 metrics: PackageManagerMetrics - current atomic counter values
func GetPackageManagerMetrics() PackageManagerMetrics {
	return PackageManagerMetrics{
		DetectCalls:      detectCalls.Load(),
		DetectErrors:     detectErrors.Load(),
		ManagerCreations: managerCreations.Load(),
		OSReleaseReads:   osReleaseReads.Load(),
		BinaryFallbacks:  binaryFallbacks.Load(),
	}
}

// ResetPackageManagerMetrics atomically resets all package manager performance counters to zero
// Code block:
//
//	ResetPackageManagerMetrics()
//	metrics := GetPackageManagerMetrics() // All counters will be 0
//
// Parameters: N/A
//
// Returns: N/A
func ResetPackageManagerMetrics() {
	detectCalls.Store(0)
	detectErrors.Store(0)
	managerCreations.Store(0)
	osReleaseReads.Store(0)
	binaryFallbacks.Store(0)
}

// ClearCache clears the cached package manager for testing purposes
// Code block:
//
//	func TestPackageManager(t *testing.T) {
//	    defer ClearCache() // Clean up after test
//	    // Test logic here
//	}
//
// Parameters: N/A
//
// Returns: N/A
func ClearCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	cachedManager = nil
	cacheOnce = sync.Once{}
}
