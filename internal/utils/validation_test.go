package utils_test

import (
	"testing"

	"github.com/kodflow/superviz.io/internal/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestRequireOneTarget(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid single target",
			args:    []string{"user@host"},
			wantErr: false,
		},
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
			errMsg:  "you must specify the target as user@host",
		},
		{
			name:    "multiple arguments (two)",
			args:    []string{"user@host1", "user@host2"},
			wantErr: true,
			errMsg:  "you must specify the target as user@host",
		},
		{
			name:    "multiple arguments (three)",
			args:    []string{"user@host1", "user@host2", "user@host3"},
			wantErr: true,
			errMsg:  "you must specify the target as user@host",
		},
		{
			name:    "empty string argument",
			args:    []string{""},
			wantErr: true,
			errMsg:  "target must be in format user@host",
		},
		{
			name:    "malformed target format (no @)",
			args:    []string{"not-a-target"},
			wantErr: true,
			errMsg:  "target must be in format user@host",
		},
		{
			name:    "multiple @ symbols",
			args:    []string{"user@host@extra"},
			wantErr: true,
			errMsg:  "target must be in format user@host",
		},
		{
			name:    "missing user",
			args:    []string{"@host"},
			wantErr: true,
			errMsg:  "target must be in format user@host",
		},
		{
			name:    "missing host",
			args:    []string{"user@"},
			wantErr: true,
			errMsg:  "target must be in format user@host",
		},
	}

	for _, tc := range cases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			err := utils.RequireOneTarget(cmd, tc.args)

			if tc.wantErr {
				require.Error(t, err)
				require.Equal(t, tc.errMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestValidatePackageNames tests the ValidatePackageNames function with comprehensive security cases.
func TestValidatePackageNames(t *testing.T) {
	cases := []struct {
		name     string
		packages []string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid single package",
			packages: []string{"nginx"},
			wantErr:  false,
		},
		{
			name:     "valid multiple packages",
			packages: []string{"nginx", "curl", "git"},
			wantErr:  false,
		},
		{
			name:     "valid package with version",
			packages: []string{"nginx=1.20.1"},
			wantErr:  false,
		},
		{
			name:     "valid package with arch",
			packages: []string{"nginx-amd64"},
			wantErr:  false,
		},
		{
			name:     "valid package with dots and hyphens",
			packages: []string{"lib.package-name"},
			wantErr:  false,
		},
		{
			name:     "valid package with numbers",
			packages: []string{"package123"},
			wantErr:  false,
		},
		{
			name:     "valid package with underscore",
			packages: []string{"package_name"},
			wantErr:  false,
		},
		{
			name:     "valid package with plus",
			packages: []string{"g++"},
			wantErr:  false,
		},
		{
			name:     "empty packages list",
			packages: []string{},
			wantErr:  true,
			errMsg:   "no package names provided",
		},
		{
			name:     "empty package name",
			packages: []string{""},
			wantErr:  true,
			errMsg:   "package name cannot be empty",
		},
		{
			name:     "package with semicolon (command injection)",
			packages: []string{"nginx; rm -rf /"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with pipe (command injection)",
			packages: []string{"nginx | cat /etc/passwd"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with ampersand (command injection)",
			packages: []string{"nginx && rm -rf /"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with backtick (command injection)",
			packages: []string{"nginx`whoami`"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with dollar (variable expansion)",
			packages: []string{"nginx$HOME"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with parentheses (subshell)",
			packages: []string{"nginx$(ls)"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with newline (command injection)",
			packages: []string{"nginx\nrm -rf /"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with tab (command injection)",
			packages: []string{"nginx\trm -rf /"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with carriage return",
			packages: []string{"nginx\rrm -rf /"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with less than (redirection)",
			packages: []string{"nginx < /etc/passwd"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with greater than (redirection)",
			packages: []string{"nginx > /dev/null"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with asterisk (globbing)",
			packages: []string{"nginx*"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with question mark (globbing)",
			packages: []string{"nginx?"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with square brackets (globbing)",
			packages: []string{"nginx[a-z]"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with curly braces (expansion)",
			packages: []string{"nginx{1,2}"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with forward slash (path)",
			packages: []string{"../nginx"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with backslash (escape)",
			packages: []string{"nginx\\test"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with single quote (injection)",
			packages: []string{"nginx'test'"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with double quote (injection)",
			packages: []string{"nginx\"test\""},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "mixed valid and invalid packages",
			packages: []string{"nginx", "curl; rm -rf /", "git"},
			wantErr:  true,
			errMsg:   "package name 'curl; rm -rf /' contains invalid characters",
		},
		{
			name:     "package with spaces",
			packages: []string{"nginx test"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "package with null byte",
			packages: []string{"nginx\x00"},
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
	}

	for _, tc := range cases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			err := utils.ValidatePackageNames(tc.packages...)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
