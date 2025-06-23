package pkgmanager

import (
	"context"
	"fmt"
	"strings"
)

type APT struct{}

func NewAPT() *APT {
	return &APT{}
}

func (m *APT) Name() string { return "apt" }

func (m *APT) Update(ctx context.Context) (string, error) {
	return "sudo apt update", nil
}

func (m *APT) Upgrade(ctx context.Context) (string, error) {
	return "sudo apt upgrade -y", nil
}

func (m *APT) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	return fmt.Sprintf("sudo apt install -y %s", strings.Join(pkgs, " ")), nil
}

func (m *APT) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	return fmt.Sprintf("sudo apt remove -y %s", strings.Join(pkgs, " ")), nil
}

func (m *APT) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", fmt.Errorf("package name required")
	}
	return fmt.Sprintf("dpkg -s %s | grep Version", pkg), nil
}

func (m *APT) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", "", fmt.Errorf("package name required")
	}
	installed := fmt.Sprintf("dpkg-query -W -f='${Version}' %s", pkg)
	available := fmt.Sprintf("apt-cache policy %s | grep Candidate | awk '{print $2}'", pkg)
	return installed, available, nil
}
