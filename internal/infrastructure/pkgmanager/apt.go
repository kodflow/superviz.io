package pkgmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/kodflow/superviz.io/internal/utils"
)

// APT implémente le gestionnaire de paquets pour les distributions basées sur Debian.
type APT struct{}

// NewAPT crée une nouvelle instance de gestionnaire APT.
//
// Returns:
//   - Pointeur vers une structure APT
func NewAPT() *APT {
	return &APT{}
}

// Name retourne le nom du gestionnaire de paquets.
//
// Returns:
//   - Nom du gestionnaire ("apt")
func (m *APT) Name() string { return "apt" }

// Update retourne la commande shell pour mettre à jour l'index des paquets.
//
// Parameters:
//   - ctx: Context pour timeout et annulation
//
// Returns:
//   - Chaîne de commande shell
//   - Erreur éventuelle
func (m *APT) Update(ctx context.Context) (string, error) {
	return "sudo apt update", nil
}

// Upgrade retourne la commande shell pour mettre à jour tous les paquets installés.
//
// Parameters:
//   - ctx: Context pour timeout et annulation
//
// Returns:
//   - Chaîne de commande shell
//   - Erreur éventuelle
func (m *APT) Upgrade(ctx context.Context) (string, error) {
	return "sudo apt upgrade -y", nil
}

// Install retourne la commande shell pour installer un ou plusieurs paquets.
//
// Parameters:
//   - ctx: Context pour timeout et annulation
//   - pkgs: Liste des paquets à installer
//
// Returns:
//   - Chaîne de commande shell
//   - Erreur si aucun paquet n'est spécifié ou si un nom de paquet est invalide
func (m *APT) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo apt install -y %s", strings.Join(pkgs, " ")), nil
}

// Remove retourne la commande shell pour désinstaller un ou plusieurs paquets.
//
// Parameters:
//   - ctx: Context pour timeout et annulation
//   - pkgs: Liste des paquets à désinstaller
//
// Returns:
//   - Chaîne de commande shell
//   - Erreur si aucun paquet n'est spécifié ou si un nom de paquet est invalide
func (m *APT) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo apt remove -y %s", strings.Join(pkgs, " ")), nil
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
func (m *APT) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if err := utils.ValidatePackageNames(pkg); err != nil {
		return "", err
	}
	return fmt.Sprintf("dpkg -s %s | grep Version", pkg), nil
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
func (m *APT) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if err := utils.ValidatePackageNames(pkg); err != nil {
		return "", "", err
	}
	installed := fmt.Sprintf("dpkg-query -W -f='${Version}' %s", pkg)
	available := fmt.Sprintf("apt-cache policy %s | grep Candidate | awk '{print $2}'", pkg)
	return installed, available, nil
}
