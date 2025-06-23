package utils

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// RequireOneTarget validates that exactly one argument is provided in "user@host" format.
// This function is designed for use as a Cobra command argument validator.
// Parameters:
//   - cmd: The cobra command (unused)
//   - args: Command line arguments to validate
//
// Returns:
//   - error: nil if validation passes, descriptive error otherwise
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

// ValidatePackageNames vérifie que chaque nom de paquet est non vide et ne contient aucun caractère dangereux.
//
// Parameters:
//   - pkgs: Liste des noms de paquets à valider
//
// Returns:
//   - error: nil si tous les noms sont valides, sinon une erreur descriptive
func ValidatePackageNames(pkgs ...string) error {
	for _, pkg := range pkgs {
		if strings.TrimSpace(pkg) == "" || strings.ContainsAny(pkg, ";|&$`") {
			return fmt.Errorf("invalid package name: %s", pkg)
		}
	}
	return nil
}
