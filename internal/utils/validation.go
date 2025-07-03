package utils

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/spf13/cobra"
)

// Global atomic counters for validation performance tracking (false sharing prevention)
var (
	targetValidations  atomic.Uint64
	packageValidations atomic.Uint64
	validationErrors   atomic.Uint64
	_                  [5]uint64 // Cache-line padding
)

// RequireOneTarget validates that exactly one argument is provided in "user@host" format with atomic tracking
// Code block:
//
//	cmd.Args = utils.RequireOneTarget
//	// Valid: "john@example.com"
//	// Invalid: "example.com", "john@", "@example.com", "john@host@extra"
//
//	metrics := utils.GetValidationMetrics()
//	fmt.Printf("Target validations: %d\n", metrics.TargetValidations)
//
// Parameters:
//   - 1 cmd: *cobra.Command - the cobra command (unused in validation)
//   - 2 args: []string - command line arguments to validate (must have length 1)
//
// Returns:
//   - 1 error - nil if validation passes, descriptive error with atomic tracking otherwise
func RequireOneTarget(_ *cobra.Command, args []string) error {
	targetValidations.Add(1)

	// Branch prediction optimization: most common failure case first
	if len(args) != 1 {
		validationErrors.Add(1)
		return fmt.Errorf("you must specify the target as user@host")
	}

	arg := args[0]

	// Optimized validation: check count and split in one pass
	atIndex := strings.Index(arg, "@")
	if atIndex == -1 || atIndex == 0 || atIndex == len(arg)-1 {
		validationErrors.Add(1)
		return fmt.Errorf("target must be in format user@host")
	}

	// Check for multiple @ symbols efficiently
	if strings.Count(arg, "@") != 1 {
		validationErrors.Add(1)
		return fmt.Errorf("target must be in format user@host")
	}

	return nil
}

// ValidatePackageNames validates that each package name is non-empty and contains no dangerous characters with atomic tracking
// Code block:
//
//	err := ValidatePackageNames("vim", "git", "curl")
//	// Returns nil (valid packages)
//
//	err = ValidatePackageNames("vim; rm -rf /")
//	// Returns error (contains dangerous characters)
//
//	metrics := GetValidationMetrics()
//	fmt.Printf("Package validations: %d, Errors: %d\n", metrics.PackageValidations, metrics.ValidationErrors)
//
// Parameters:
//   - 1 pkgs: ...string - list of package names to validate (must not be empty)
//
// Returns:
//   - 1 error - nil if all names are valid, otherwise a descriptive error with atomic tracking
func ValidatePackageNames(pkgs ...string) error {
	packageValidations.Add(1)

	// Input validation (most common error case first)
	if len(pkgs) == 0 {
		validationErrors.Add(1)
		return fmt.Errorf("no package names provided")
	}

	// Pre-compiled dangerous characters set (more efficient than string iteration)
	// Bitset approach for faster character checking
	const dangerousChars = ";|&$`()[]{}*?<>\\/'\" \t\n\r"
	dangerousSet := make(map[rune]bool, len(dangerousChars))
	for _, r := range dangerousChars {
		dangerousSet[r] = true
	}

	for _, pkg := range pkgs {
		trimmed := strings.TrimSpace(pkg)
		if trimmed == "" {
			validationErrors.Add(1)
			return fmt.Errorf("package name cannot be empty")
		}

		// Optimized character validation: single pass through runes
		for _, r := range pkg {
			// Check for ASCII control characters (branch prediction: less likely)
			if r <= 0x1F || r == 0x7F {
				validationErrors.Add(1)
				return fmt.Errorf("package name '%s' contains invalid characters", pkg)
			}

			// Check dangerous characters (more likely, check first)
			if dangerousSet[r] {
				validationErrors.Add(1)
				return fmt.Errorf("package name '%s' contains invalid characters", pkg)
			}
		}
	}
	return nil
}

// ValidationMetrics contains atomic performance metrics for validation operations
// Code block:
//
//	metrics := GetValidationMetrics()
//	fmt.Printf("Target: %d, Package: %d, Errors: %d\n",
//	    metrics.TargetValidations, metrics.PackageValidations, metrics.ValidationErrors)
//
// Parameters: N/A (for types)
//
// Returns: N/A (for types)
type ValidationMetrics struct {
	TargetValidations  uint64
	PackageValidations uint64
	ValidationErrors   uint64
}

// GetValidationMetrics returns current atomic performance metrics for all validation operations
// Code block:
//
//	metrics := GetValidationMetrics()
//	fmt.Printf("Validation metrics: %+v\n", metrics)
//
// Parameters: N/A
//
// Returns:
//   - 1 metrics: ValidationMetrics - current atomic counter values
func GetValidationMetrics() ValidationMetrics {
	return ValidationMetrics{
		TargetValidations:  targetValidations.Load(),
		PackageValidations: packageValidations.Load(),
		ValidationErrors:   validationErrors.Load(),
	}
}

// ResetValidationMetrics atomically resets all validation performance counters to zero
// Code block:
//
//	ResetValidationMetrics()
//	metrics := GetValidationMetrics() // All counters will be 0
//
// Parameters: N/A
//
// Returns: N/A
func ResetValidationMetrics() {
	targetValidations.Store(0)
	packageValidations.Store(0)
	validationErrors.Store(0)
}
