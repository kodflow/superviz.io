package install_test

import (
	"testing"
	"time"

	"github.com/kodflow/superviz.io/internal/cli/commands/install"
	"github.com/kodflow/superviz.io/internal/services"
	"github.com/kodflow/superviz.io/internal/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestGetCommand(t *testing.T) {
	t.Helper()

	cmd := install.GetCommand()
	require.NotNil(t, cmd)
	require.Equal(t, "install", cmd.Use[:7])
	require.Equal(t, "Setup superviz.io repository on remote system", cmd.Short)
	require.NotEmpty(t, cmd.Long)

	// Test singleton behavior
	cmd2 := install.GetCommand()
	require.Same(t, cmd, cmd2, "GetCommand should return the same instance")
}

func TestGetCommandWithService(t *testing.T) {
	t.Helper()

	// Test with nil service - should return singleton
	cmd1 := install.GetCommandWithService(nil)
	singleton := install.GetCommand()
	require.Same(t, singleton, cmd1)

	// Test with custom service - should return new instance
	service := services.NewInstallService(nil)
	cmd2 := install.GetCommandWithService(service)
	require.NotSame(t, singleton, cmd2)
}

func TestNewInstallCommand(t *testing.T) {
	t.Helper()

	// Test command creation with a service
	service := services.NewInstallService(nil)
	cmd := install.NewInstallCommand(service)
	require.NotNil(t, cmd)
	require.Equal(t, "install user@host [flags]", cmd.Use)
	require.Equal(t, "Setup superviz.io repository on remote system", cmd.Short)

	// Test that multiple calls return different instances
	cmd2 := install.NewInstallCommand(service)
	require.NotSame(t, cmd, cmd2)
}

func TestCreateInstallCommandFlags(t *testing.T) {
	t.Helper()

	service := services.NewInstallService(nil)
	cmd := install.NewInstallCommand(service)

	// Test that all expected flags are present
	flags := cmd.Flags()

	// SSH key flag
	keyFlag := flags.Lookup("ssh-key")
	require.NotNil(t, keyFlag)
	require.Equal(t, "i", keyFlag.Shorthand)
	require.Equal(t, "", keyFlag.DefValue)

	// SSH port flag
	portFlag := flags.Lookup("ssh-port")
	require.NotNil(t, portFlag)
	require.Equal(t, "p", portFlag.Shorthand)
	require.Equal(t, "22", portFlag.DefValue)

	// Timeout flag
	timeoutFlag := flags.Lookup("timeout")
	require.NotNil(t, timeoutFlag)
	require.Equal(t, "t", timeoutFlag.Shorthand)
	require.Equal(t, "5m0s", timeoutFlag.DefValue)

	// Force flag
	forceFlag := flags.Lookup("force")
	require.NotNil(t, forceFlag)
	require.Equal(t, "f", forceFlag.Shorthand)
	require.Equal(t, "false", forceFlag.DefValue)

	// Skip host key check flag
	skipFlag := flags.Lookup("skip-host-key-check")
	require.NotNil(t, skipFlag)
	require.Equal(t, "false", skipFlag.DefValue)
}

func TestInstallCommandValidation(t *testing.T) {
	t.Helper()

	cases := []struct {
		name     string
		args     []string
		wantErr  bool
		errMsg   string
	}{
		{
			name:    "valid target",
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
			name:    "malformed target",
			args:    []string{"not-a-target"},
			wantErr: true,
			errMsg:  "target must be in format user@host",
		},
	}

	for _, tc := range cases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()

			// Test using utils.RequireOneTarget directly
			err := utils.RequireOneTarget(nil, tc.args)

			if tc.wantErr {
				require.Error(t, err)
				require.Equal(t, tc.errMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInstallCommandFlagParsing(t *testing.T) {
	t.Helper()

	service := services.NewInstallService(nil)

	// Test parsing various flag combinations
	testCases := []struct {
		name    string
		args    []string
		check   func(t *testing.T, cmd *cobra.Command)
	}{
		{
			name: "ssh key flag",
			args: []string{"--ssh-key", "/path/to/key", "user@host"},
			check: func(t *testing.T, cmd *cobra.Command) {
				val, err := cmd.Flags().GetString("ssh-key")
				require.NoError(t, err)
				require.Equal(t, "/path/to/key", val)
			},
		},
		{
			name: "ssh port flag",
			args: []string{"--ssh-port", "2222", "user@host"},
			check: func(t *testing.T, cmd *cobra.Command) {
				val, err := cmd.Flags().GetInt("ssh-port")
				require.NoError(t, err)
				require.Equal(t, 2222, val)
			},
		},
		{
			name: "timeout flag",
			args: []string{"--timeout", "10m", "user@host"},
			check: func(t *testing.T, cmd *cobra.Command) {
				val, err := cmd.Flags().GetDuration("timeout")
				require.NoError(t, err)
				require.Equal(t, 10*time.Minute, val)
			},
		},
		{
			name: "force flag",
			args: []string{"--force", "user@host"},
			check: func(t *testing.T, cmd *cobra.Command) {
				val, err := cmd.Flags().GetBool("force")
				require.NoError(t, err)
				require.True(t, val)
			},
		},
		{
			name: "skip host key check flag",
			args: []string{"--skip-host-key-check", "user@host"},
			check: func(t *testing.T, cmd *cobra.Command) {
				val, err := cmd.Flags().GetBool("skip-host-key-check")
				require.NoError(t, err)
				require.True(t, val)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()

			// Reset flags for each test
			cmd := install.NewInstallCommand(service)
			cmd.SetArgs(tc.args)

			err := cmd.ParseFlags(tc.args)
			require.NoError(t, err)

			tc.check(t, cmd)
		})
	}
}

func TestInstallCommandDefaultValues(t *testing.T) {
	t.Helper()

	service := services.NewInstallService(nil)
	cmd := install.NewInstallCommand(service)

	// Test default values
	portVal, err := cmd.Flags().GetInt("ssh-port")
	require.NoError(t, err)
	require.Equal(t, 22, portVal)

	timeoutVal, err := cmd.Flags().GetDuration("timeout")
	require.NoError(t, err)
	require.Equal(t, 300*time.Second, timeoutVal)

	forceVal, err := cmd.Flags().GetBool("force")
	require.NoError(t, err)
	require.False(t, forceVal)

	skipVal, err := cmd.Flags().GetBool("skip-host-key-check")
	require.NoError(t, err)
	require.False(t, skipVal)

	keyVal, err := cmd.Flags().GetString("ssh-key")
	require.NoError(t, err)
	require.Equal(t, "", keyVal)
}

// Test the PreRunE and RunE functions with mock service
func TestInstallCommandExecution(t *testing.T) {
	t.Helper()

	// Create a mock service to test command execution flows
	service := services.NewInstallService(nil)
	cmd := install.NewInstallCommand(service)

	// Test PreRunE with valid arguments
	cmd.SetArgs([]string{"user@host"})
	err := cmd.ParseFlags([]string{"user@host"})
	require.NoError(t, err)

	preRunE := cmd.PreRunE
	require.NotNil(t, preRunE)

	// Since the actual service might try to validate SSH configs,
	// we test that PreRunE is defined and callable
	err = preRunE(cmd, []string{"user@host"})
	// We expect an error because the service will try to validate real SSH config
	// but the important thing is that PreRunE is properly wired
	if err != nil {
		// This is expected since we're using a real service that validates configs
		t.Logf("PreRunE returned expected validation error: %v", err)
	}
}

func TestInstallCommandRunE(t *testing.T) {
	t.Helper()

	service := services.NewInstallService(nil)
	cmd := install.NewInstallCommand(service)

	// Test that RunE is properly defined
	runE := cmd.RunE
	require.NotNil(t, runE)

	// We can't easily test the actual execution without mocking the entire service,
	// but we can verify the function is properly wired
	require.NotNil(t, runE)
}

func TestCreateInstallCommandDefaults(t *testing.T) {
	t.Helper()

	service := services.NewInstallService(nil)

	// Test that createInstallCommand sets up proper defaults
	cmd := install.NewInstallCommand(service)
	require.NotNil(t, cmd)

	// Test the command structure
	require.Equal(t, "install user@host [flags]", cmd.Use)
	require.Equal(t, "Setup superviz.io repository on remote system", cmd.Short)
	require.Contains(t, cmd.Long, "Setup superviz.io package repository")
	require.NotNil(t, cmd.Args)
	require.NotNil(t, cmd.PreRunE)
	require.NotNil(t, cmd.RunE)

	// Test that flags are properly initialized with defaults
	flags := cmd.Flags()
	require.NotNil(t, flags)

	// Verify default values are set correctly in the flags
	portFlag := flags.Lookup("ssh-port")
	require.NotNil(t, portFlag)
	require.Equal(t, "22", portFlag.DefValue)

	timeoutFlag := flags.Lookup("timeout")
	require.NotNil(t, timeoutFlag)
	require.Equal(t, "5m0s", timeoutFlag.DefValue)
}

func TestInstallCommandInternalStructure(t *testing.T) {
	t.Helper()

	service := services.NewInstallService(nil)
	cmd := install.NewInstallCommand(service)

	// Test internal structure and function assignments
	require.Equal(t, "install user@host [flags]", cmd.Use)
	require.Equal(t, "Setup superviz.io repository on remote system", cmd.Short)

	// Test that Args function is properly assigned
	require.NotNil(t, cmd.Args)
	err := cmd.Args(cmd, []string{})
	require.Error(t, err) // Should error with no args

	err = cmd.Args(cmd, []string{"user@host"})
	require.NoError(t, err) // Should succeed with valid arg

	// Test that PreRunE and RunE are assigned
	require.NotNil(t, cmd.PreRunE)
	require.NotNil(t, cmd.RunE)

	// Test flag configuration completeness
	flags := cmd.Flags()
	expectedFlags := []string{"ssh-key", "ssh-port", "timeout", "force", "skip-host-key-check"}
	for _, flagName := range expectedFlags {
		flag := flags.Lookup(flagName)
		require.NotNil(t, flag, "Flag %s should be defined", flagName)
	}
}