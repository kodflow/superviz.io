package pkgmanager

import (
	"context"
	"fmt"
	"strings"
)

type ZYPPER struct{}

func NewZYPPER() *ZYPPER {
	return &ZYPPER{}
}

func (m *ZYPPER) Name() string { return "zypper" }

func (m *ZYPPER) Update(ctx context.Context) (string, error) {
	return "sudo zypper refresh", nil
}

func (m *ZYPPER) Upgrade(ctx context.Context) (string, error) {
	return "sudo zypper update -y", nil
}

func (m *ZYPPER) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	return fmt.Sprintf("sudo zypper install -y %s", strings.Join(pkgs, " ")), nil
}

func (m *ZYPPER) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	return fmt.Sprintf("sudo zypper remove -y %s", strings.Join(pkgs, " ")), nil
}

func (m *ZYPPER) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", fmt.Errorf("package name required")
	}
	return fmt.Sprintf("zypper se --installed-only %s", pkg), nil
}

func (m *ZYPPER) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", "", fmt.Errorf("package name required")
	}
	installed := fmt.Sprintf("zypper info %s | grep Version | head -1 | awk '{print $3}'", pkg)
	available := fmt.Sprintf("zypper info %s | grep Version | tail -1 | awk '{print $3}'", pkg)
	return installed, available, nil
}
