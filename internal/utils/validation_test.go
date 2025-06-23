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
