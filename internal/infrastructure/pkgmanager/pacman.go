package pkgmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/kodflow/superviz.io/internal/utils"
)

// PACMAN implémente le gestionnaire de paquets pour Arch Linux et dérivés.
type PACMAN struct{}

// NewPACMAN crée une nouvelle instance de gestionnaire PACMAN.
//
// Returns:
//   - Pointeur vers une structure PACMAN
func NewPACMAN() *PACMAN {
	return &PACMAN{}
}

// Name retourne le nom du gestionnaire de paquets.
//
// Returns:
//   - Nom du gestionnaire ("pacman")
func (m *PACMAN) Name() string { return "pacman" }

// Update retourne la commande shell pour mettre à jour l'index des paquets.
//
// Parameters:
//   - ctx: Context pour timeout et annulation
//
// Returns:
//   - Chaîne de commande shell
//   - Erreur éventuelle
func (m *PACMAN) Update(ctx context.Context) (string, error) {
	return "sudo pacman -Sy", nil
}

// Upgrade retourne la commande shell pour mettre à jour tous les paquets installés.
//
// Parameters:
//   - ctx: Context pour timeout et annulation
//
// Returns:
//   - Chaîne de commande shell
//   - Erreur éventuelle
func (m *PACMAN) Upgrade(ctx context.Context) (string, error) {
	return "sudo pacman -Su --noconfirm", nil
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
func (m *PACMAN) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo pacman -S --noconfirm %s", strings.Join(pkgs, " ")), nil
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
func (m *PACMAN) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo pacman -Rns --noconfirm %s", strings.Join(pkgs, " ")), nil
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
func (m *PACMAN) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", fmt.Errorf("package name required")
	}
	return fmt.Sprintf("pacman -Qi %s", pkg), nil
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
func (m *PACMAN) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", "", fmt.Errorf("package name required")
	}
	installed := fmt.Sprintf("pacman -Qi %s | grep Version | awk '{print $3}'", pkg)
	available := fmt.Sprintf("pacman -Si %s | grep Version | awk '{print $3}'", pkg)

	return installed, available, nil
}
