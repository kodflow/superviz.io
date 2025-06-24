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
// Cette fonction assure la sécurité en empêchant l'injection de commandes via les noms de paquets.
// Elle vérifie les caractères suivants comme potentiellement dangereux :
// - Caractères de contrôle de commande : ; | & $ ` ( ) [ ] { } < > * ? \ / ' "
// - Caractères d'espacement : espaces, tabs, newlines, carriage returns
// - Caractères de contrôle ASCII (0x00-0x1F, 0x7F)
//
// Parameters:
//   - pkgs: Liste des noms de paquets à valider
//
// Returns:
//   - error: nil si tous les noms sont valides, sinon une erreur descriptive
func ValidatePackageNames(pkgs ...string) error {
	if len(pkgs) == 0 {
		return fmt.Errorf("no package names provided")
	}

	// Caractères dangereux pour l'injection de commandes
	dangerousChars := ";|&$`()[]{}*?<>\\/'\" \t\n\r"

	for _, pkg := range pkgs {
		trimmed := strings.TrimSpace(pkg)
		if trimmed == "" {
			return fmt.Errorf("package name cannot be empty")
		}

		// Vérifier les caractères dangereux
		if strings.ContainsAny(pkg, dangerousChars) {
			return fmt.Errorf("package name '%s' contains invalid characters", pkg)
		}

		// Vérifier les caractères de contrôle ASCII
		for _, r := range pkg {
			if r <= 0x1F || r == 0x7F {
				return fmt.Errorf("package name '%s' contains invalid characters", pkg)
			}
		}
	}
	return nil
}
