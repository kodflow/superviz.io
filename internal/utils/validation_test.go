package utils_test

import (
	"testing"

	"github.com/kodflow/superviz.io/internal/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRequireOneTarget(t *testing.T) {
	tests := []struct {
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
			name:    "multiple arguments",
			args:    []string{"user@host1", "user@host2"},
			wantErr: true,
			errMsg:  "you must specify the target as user@host",
		},
		{
			name:    "three arguments",
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
			wantErr: false, // selon implémentation actuelle (vérifie juste présence de @)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			err := utils.RequireOneTarget(cmd, tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
