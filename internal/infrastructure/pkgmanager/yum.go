package pkgmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/kodflow/superviz.io/internal/utils"
)

// YUM implémente le gestionnaire de paquets pour CentOS, RHEL et dérivés.
type YUM struct{}

// NewYUM crée une nouvelle instance de gestionnaire YUM.
//
// Returns:
//   - Pointeur vers une structure YUM
func NewYUM() *YUM {
	return &YUM{}
}

// Name retourne le nom du gestionnaire de paquets.
//
// Returns:
//   - Nom du gestionnaire ("yum")
func (m *YUM) Name() string { return "yum" }

// Update retourne la commande shell pour mettre à jour l'index des paquets.
//
// Parameters:
//   - ctx: Context pour timeout et annulation
//
// Returns:
//   - Chaîne de commande shell
//   - Erreur éventuelle
func (m *YUM) Update(ctx context.Context) (string, error) {
	return "sudo yum check-update", nil
}

// Upgrade retourne la commande shell pour mettre à jour tous les paquets installés.
//
// Parameters:
//   - ctx: Context pour timeout et annulation
//
// Returns:
//   - Chaîne de commande shell
//   - Erreur éventuelle
func (m *YUM) Upgrade(ctx context.Context) (string, error) {
	return "sudo yum upgrade -y", nil
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
func (m *YUM) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo yum install -y %s", strings.Join(pkgs, " ")), nil
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
func (m *YUM) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo yum remove -y %s", strings.Join(pkgs, " ")), nil
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
func (m *YUM) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", fmt.Errorf("package name required")
	}
	return fmt.Sprintf("yum list installed %s", pkg), nil
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
func (m *YUM) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", "", fmt.Errorf("package name required")
	}
	installed := fmt.Sprintf("yum info %s | grep Version", pkg)
	available := fmt.Sprintf("yum --showduplicates list %s | grep -v Installed | awk '{print $2}'", pkg)
	return installed, available, nil
}
