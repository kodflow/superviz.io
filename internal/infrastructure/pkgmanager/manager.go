package pkgmanager

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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

// Manager définit l'interface pour tous les gestionnaires de paquets détectés.
//
// Chaque implémentation doit fournir des méthodes pour les opérations courantes sur les paquets.
type Manager interface {
	// Name retourne le nom du gestionnaire (ex: apt, apk...)
	Name() string
	// Update retourne la commande à exécuter pour faire un update.
	Update(ctx context.Context) (string, error)
	// Install retourne la commande pour installer un ou plusieurs paquets.
	Install(ctx context.Context, pkgs ...string) (string, error)
	// Remove retourne la commande pour désinstaller un ou plusieurs paquets.
	Remove(ctx context.Context, pkgs ...string) (string, error)
	// Upgrade retourne la commande pour faire un upgrade global.
	Upgrade(ctx context.Context) (string, error)
	// IsInstalled retourne la commande pour vérifier si un paquet est installé et sa version actuelle.
	IsInstalled(ctx context.Context, pkg string) (string, error)
	// VersionCheck retourne la commande pour comparer la version installée et la version du dépôt.
	VersionCheck(ctx context.Context, pkg string) (installedVersion string, availableVersion string, err error)
}

// Detect retourne le gestionnaire de paquets approprié en se basant sur /etc/os-release,
// avec un fallback sur les binaires présents dans le PATH.
//
// Returns:
//   - Instance de Manager correspondant à la distribution
//   - Erreur si aucun gestionnaire n'est détecté
func Detect() (Manager, error) {
	const osRelease = "/etc/os-release"

	if file, err := os.Open(osRelease); err == nil {
		defer file.Close() // nolint: errcheck

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "ID=") {
				distro := strings.Trim(strings.TrimPrefix(line, "ID="), `"`)
				if bin, ok := distroToPkgManager[distro]; ok {
					return detectFromBin(bin)
				}
				break
			}
		}
	}

	// Fallback : inspecter les binaires dans le PATH
	for _, bin := range []string{"apt", "apk", "dnf", "yum", "pacman", "zypper", "emerge"} {
		if _, err := exec.LookPath(bin); err == nil {
			return detectFromBin(bin)
		}
	}

	return nil, fmt.Errorf("unable to detect package manager")
}

// detectFromBin retourne une instance de Manager selon le nom du binaire.
//
// Parameters:
//   - bin: Nom du binaire à détecter (ex: "apt", "yum")
//
// Returns:
//   - Instance de Manager
//   - Erreur si le binaire n'est pas supporté
func detectFromBin(bin string) (Manager, error) {
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
