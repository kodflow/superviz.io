package pkgmanager

import (
	"context"
	"fmt"
	"strings"
)

type PACMAN struct{}

func NewPACMAN() *PACMAN {
	return &PACMAN{}
}

func (m *PACMAN) Name() string { return "pacman" }

func (m *PACMAN) Update(ctx context.Context) (string, error) {
	return "sudo pacman -Sy", nil
}

func (m *PACMAN) Upgrade(ctx context.Context) (string, error) {
	return "sudo pacman -Su --noconfirm", nil
}

func (m *PACMAN) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	return fmt.Sprintf("sudo pacman -S --noconfirm %s", strings.Join(pkgs, " ")), nil
}

func (m *PACMAN) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	return fmt.Sprintf("sudo pacman -Rns --noconfirm %s", strings.Join(pkgs, " ")), nil
}

func (m *PACMAN) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", fmt.Errorf("package name required")
	}
	return fmt.Sprintf("pacman -Qi %s", pkg), nil
}

func (m *PACMAN) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", "", fmt.Errorf("package name required")
	}
	installed := fmt.Sprintf("pacman -Qi %s | grep Version | awk '{print $3}'", pkg)
	available := fmt.Sprintf("pacman -Si %s | grep Version | awk '{print $3}'", pkg)

	return installed, available, nil
}
