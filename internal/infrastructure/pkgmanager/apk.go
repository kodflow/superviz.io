package pkgmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/kodflow/superviz.io/internal/utils"
)

// APK implémente le gestionnaire de paquets pour Alpine Linux.
type APK struct{}

// NewAPK crée une nouvelle instance de gestionnaire APK.
//
// Returns:
//   - Pointeur vers une structure APK
func NewAPK() *APK {
	return &APK{}
}

// Name retourne le nom du gestionnaire de paquets.
//
// Returns:
//   - Nom du gestionnaire ("apk")
func (m *APK) Name() string { return "apk" }

// Update retourne la commande shell pour mettre à jour l'index des paquets.
//
// Parameters:
//   - ctx: Context pour timeout et annulation
//
// Returns:
//   - Chaîne de commande shell
//   - Erreur éventuelle
func (m *APK) Update(ctx context.Context) (string, error) {
	return "sudo apk update", nil
}

// Upgrade retourne la commande shell pour mettre à jour tous les paquets installés.
//
// Parameters:
//   - ctx: Context pour timeout et annulation
//
// Returns:
//   - Chaîne de commande shell
//   - Erreur éventuelle
func (m *APK) Upgrade(ctx context.Context) (string, error) {
	return "sudo apk upgrade", nil
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
func (m *APK) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo apk add %s", strings.Join(pkgs, " ")), nil
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
func (m *APK) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	if err := utils.ValidatePackageNames(pkgs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("sudo apk del %s", strings.Join(pkgs, " ")), nil
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
func (m *APK) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", fmt.Errorf("package name required")
	}
	return fmt.Sprintf("apk info -e %s", pkg), nil
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
func (m *APK) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", "", fmt.Errorf("package name required")
	}
	installed := fmt.Sprintf("apk info -v %s | cut -d'-' -f2-", pkg)
	available := fmt.Sprintf("apk search -v %s | grep %s | cut -d'-' -f2-", pkg, pkg)

	return installed, available, nil
}
