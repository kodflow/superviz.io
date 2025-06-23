package pkgmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/kodflow/superviz.io/internal/utils"
)

// ZYPPER implémente le gestionnaire de paquets pour openSUSE et SLES.
type ZYPPER struct{}

// NewZYPPER crée une nouvelle instance de gestionnaire ZYPPER.
//
// Returns:
//   - Pointeur vers une structure ZYPPER
func NewZYPPER() *ZYPPER {
	return &ZYPPER{}
}

// Name retourne le nom du gestionnaire de paquets.
//
// Returns:
//   - Nom du gestionnaire ("zypper")
func (m *ZYPPER) Name() string { return "zypper" }

// Update retourne la commande shell pour rafraîchir l'index des paquets.
//
// Parameters:
//   - ctx: Context pour timeout et annulation
//
// Returns:
//   - Chaîne de commande shell
//   - Erreur éventuelle
func (m *ZYPPER) Update(ctx context.Context) (string, error) {
	return "sudo zypper refresh", nil
}

// Upgrade retourne la commande shell pour mettre à jour tous les paquets installés.
//
// Parameters:
//   - ctx: Context pour timeout et annulation
//
// Returns:
//   - Chaîne de commande shell
//   - Erreur éventuelle
func (m *ZYPPER) Upgrade(ctx context.Context) (string, error) {
	return "sudo zypper update -y", nil
}

// Install retourne la commande shell pour installer un ou plusieurs paquets.
//
// Parameters:
//   - ctx: Context pour timeout et annulation
//   - pkgs: Liste des paquets à installer
//
// Returns:
//   - Chaîne de commande shell
//   - Erreur si aucun paquet n'est spécifié
func (m *ZYPPER) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo zypper install -y %s", strings.Join(pkgs, " ")), nil
}

// Remove retourne la commande shell pour désinstaller un ou plusieurs paquets.
//
// Parameters:
//   - ctx: Context pour timeout et annulation
//   - pkgs: Liste des paquets à désinstaller
//
// Returns:
//   - Chaîne de commande shell
//   - Erreur si aucun paquet n'est spécifié
func (m *ZYPPER) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo zypper remove -y %s", strings.Join(pkgs, " ")), nil
}

// IsInstalled retourne la commande shell pour vérifier si un paquet est installé.
//
// Parameters:
//   - ctx: Context pour timeout et annulation
//   - pkg: Nom du paquet à vérifier
//
// Returns:
//   - Chaîne de commande shell
//   - Erreur si le nom du paquet est vide
func (m *ZYPPER) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", fmt.Errorf("package name required")
	}
	return fmt.Sprintf("zypper se --installed-only %s", pkg), nil
}

// VersionCheck retourne les commandes shell pour obtenir la version installée et disponible d'un paquet.
//
// Parameters:
//   - ctx: Context pour timeout et annulation
//   - pkg: Nom du paquet à vérifier
//
// Returns:
//   - Commande pour version installée
//   - Commande pour version disponible
//   - Erreur si le nom du paquet est vide
func (m *ZYPPER) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", "", fmt.Errorf("package name required")
	}
	installed := fmt.Sprintf("zypper info %s | grep Version | head -1 | awk '{print $3}'", pkg)
	available := fmt.Sprintf("zypper info %s | grep Version | tail -1 | awk '{print $3}'", pkg)
	return installed, available, nil
}
