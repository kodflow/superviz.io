package pkgmanager

import (
	"context"
	"fmt"
	"strings"
)

type DNF struct{}

func NewDNF() *DNF {
	return &DNF{}
}

func (m *DNF) Name() string { return "dnf" }

func (m *DNF) Update(ctx context.Context) (string, error) {
	return "sudo dnf check-update", nil
}

func (m *DNF) Upgrade(ctx context.Context) (string, error) {
	return "sudo dnf upgrade -y", nil
}

func (m *DNF) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	return fmt.Sprintf("sudo dnf install -y %s", strings.Join(pkgs, " ")), nil
}

func (m *DNF) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	return fmt.Sprintf("sudo dnf remove -y %s", strings.Join(pkgs, " ")), nil
}

func (m *DNF) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", fmt.Errorf("package name required")
	}
	return fmt.Sprintf("dnf list installed %s", pkg), nil
}

func (m *DNF) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", "", fmt.Errorf("package name required")
	}
	installed := fmt.Sprintf("dnf info %s | grep Version", pkg)
	available := fmt.Sprintf("dnf --showduplicates list %s | grep -v Installed | awk '{print $2}'", pkg)

	return installed, available, nil
}
