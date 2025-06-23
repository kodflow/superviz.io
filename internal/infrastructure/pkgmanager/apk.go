package pkgmanager

import (
	"context"
	"fmt"
	"strings"
)

type APK struct{}

func NewAPK() *APK {
	return &APK{}
}

func (m *APK) Name() string { return "apk" }

func (m *APK) Update(ctx context.Context) (string, error) {
	return "sudo apk update", nil
}

func (m *APK) Upgrade(ctx context.Context) (string, error) {
	return "sudo apk upgrade", nil
}

func (m *APK) Install(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for install")
	}
	return fmt.Sprintf("sudo apk add %s", strings.Join(pkgs, " ")), nil
}

func (m *APK) Remove(ctx context.Context, pkgs ...string) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package specified for removal")
	}
	return fmt.Sprintf("sudo apk del %s", strings.Join(pkgs, " ")), nil
}

func (m *APK) IsInstalled(ctx context.Context, pkg string) (string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", fmt.Errorf("package name required")
	}
	return fmt.Sprintf("apk info -e %s", pkg), nil
}

func (m *APK) VersionCheck(ctx context.Context, pkg string) (string, string, error) {
	if strings.TrimSpace(pkg) == "" {
		return "", "", fmt.Errorf("package name required")
	}
	installed := fmt.Sprintf("apk info -v %s | cut -d'-' -f2-", pkg)
	available := fmt.Sprintf("apk search -v %s | grep %s | cut -d'-' -f2-", pkg, pkg)

	return installed, available, nil
}
