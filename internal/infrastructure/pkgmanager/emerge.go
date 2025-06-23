package pkgmanager

import (
	"context"
	"fmt"
	"strings"
)

type EMERGE struct{}

func NewEMERGE() *EMERGE {
	return &EMERGE{}
}

func (m *EMERGE) Name() string { return "emerge" }

func (m *EMERGE) Update(ctx context.Context) (string, error) {
	return "sudo emerge --sync", nil
}

func (m *EMERGE) Upgrade(ctx context.Context) (string, error) {
	return "sudo emerge -uDN @world", nil
}

func (m *EMERGE) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	return fmt.Sprintf("sudo emerge %s", strings.Join(pkgs, " ")), nil
}

func (m *EMERGE) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	return fmt.Sprintf("sudo emerge -C %s", strings.Join(pkgs, " ")), nil
}

func (m *EMERGE) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", fmt.Errorf("package name required")
	}
	return fmt.Sprintf("equery list %s", pkg), nil
}

func (m *EMERGE) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", "", fmt.Errorf("package name required")
	}
	installed := fmt.Sprintf("equery list %s | awk '{print $2}'", pkg)
	available := fmt.Sprintf("emerge -p %s | grep '\\[ebuild' | awk '{print $4}'", pkg)

	return installed, available, nil
}
