package utils

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// RequireOneTarget validates that exactly one argument is provided in "user@host" format.
//
// RequireOneTarget ensures that command receives precisely one argument in the correct
// user@host format for SSH connections. This function is designed for use as a Cobra
// command argument validator.
//
// Example:
//
//	cmd.Args = utils.RequireOneTarget
//	// Valid: "john@example.com"
//	// Invalid: "example.com", "john@", "@example.com", "john@host@extra"
//
// Parameters:
//   - cmd: *cobra.Command the cobra command (unused in validation)
//   - args: []string command line arguments to validate
//
// Returns:
//   - err: error nil if validation passes, descriptive error otherwise
func RequireOneTarget(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("you must specify the target as user@host")
	}
	if !strings.Contains(args[0], "@") || strings.Count(args[0], "@") != 1 {
		return fmt.Errorf("target must be in format user@host")
	}
	parts := strings.Split(args[0], "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("target must be in format user@host")
	}
	return nil
}

// ValidatePackageNames validates that each package name is non-empty and contains no dangerous characters.
//
// ValidatePackageNames ensures security by preventing command injection via package names.
// It checks for the following potentially dangerous characters:
// - Command control characters: ; | & $ ` ( ) [ ] { } < > * ? \ / ' "
// - Whitespace characters: spaces, tabs, newlines, carriage returns
// - ASCII control characters (0x00-0x1F, 0x7F)
//
// Example:
//
//	err := ValidatePackageNames("vim", "git", "curl")
//	// Returns nil (valid packages)
//
//	err = ValidatePackageNames("vim; rm -rf /")
//	// Returns error (contains dangerous characters)
//
// Parameters:
//   - pkgs: ...string list of package names to validate
//
// Returns:
//   - err: error nil if all names are valid, otherwise a descriptive error
func ValidatePackageNames(pkgs ...string) error {
	if len(pkgs) == 0 {
		return fmt.Errorf("no package names provided")
	}

	// Dangerous characters for command injection
	dangerousChars := ";|&$`()[]{}*?<>\\/'\" \t\n\r"

	for _, pkg := range pkgs {
		trimmed := strings.TrimSpace(pkg)
		if trimmed == "" {
			return fmt.Errorf("package name cannot be empty")
		}

		// Check for dangerous characters
		if strings.ContainsAny(pkg, dangerousChars) {
			return fmt.Errorf("package name '%s' contains invalid characters", pkg)
		}

		// Check for ASCII control characters
		for _, r := range pkg {
			if r <= 0x1F || r == 0x7F {
				return fmt.Errorf("package name '%s' contains invalid characters", pkg)
			}
		}
	}
	return nil
}
