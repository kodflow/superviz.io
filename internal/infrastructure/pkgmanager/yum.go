package pkgmanager

import (
	"context"
	"fmt"
	"strings"
)

type YUM struct{}

func NewYUM() *YUM {
	return &YUM{}
}

func (m *YUM) Name() string { return "yum" }

func (m *YUM) Update(ctx context.Context) (string, error) {
	return "sudo yum check-update", nil
}

func (m *YUM) Upgrade(ctx context.Context) (string, error) {
	return "sudo yum upgrade -y", nil
}

func (m *YUM) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	return fmt.Sprintf("sudo yum install -y %s", strings.Join(pkgs, " ")), nil
}

func (m *YUM) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	return fmt.Sprintf("sudo yum remove -y %s", strings.Join(pkgs, " ")), nil
}

func (m *YUM) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", fmt.Errorf("package name required")
	}
	return fmt.Sprintf("yum list installed %s", pkg), nil
}

func (m *YUM) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", "", fmt.Errorf("package name required")
	}
	installed := fmt.Sprintf("yum info %s | grep Version", pkg)
	available := fmt.Sprintf("yum --showduplicates list %s | grep -v Installed | awk '{print $2}'", pkg)
	return installed, available, nil
}
