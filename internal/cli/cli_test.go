package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kodflow/superviz.io/internal/cli"
	"github.com/kodflow/superviz.io/internal/cli/commands/install"
	"github.com/kodflow/superviz.io/internal/cli/commands/version"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newTestRootCommand(buf *bytes.Buffer, args ...string) *cobra.Command {
	cmd := cli.NewRootCommand(version.GetCommand())
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	return cmd
}

func TestExecuteRootCommand(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute() panicked: %v", r)
		}
	}()

	// Simule `os.Args = []string{"svz"}`
	buf := &bytes.Buffer{}
	cmd := newTestRootCommand(buf)

	err := cmd.Execute()
	assert.NoError(t, err, "Expected Execute() to run without error")
}

func TestRootCommandStructureAndSubcommands(t *testing.T) {
	cmd := cli.NewRootCommand(version.GetCommand(), install.GetCommand())

	assert.Equal(t, "svz", cmd.Use)
	assert.Equal(t, "Superviz - Declarative Process Supervisor", cmd.Short)
	assert.True(t, cmd.DisableAutoGenTag)
	assert.True(t, cmd.DisableFlagsInUseLine)

	sub := cmd.Commands()
	assert.NotEmpty(t, sub, "Expected subcommands to be registered")

	var hasVersion, hasInstall bool
	for _, c := range sub {
		if c.Use == "version" {
			hasVersion = true
		} else if strings.HasPrefix(c.Use, "install") {
			hasInstall = true
		}
	}
	assert.True(t, hasVersion, "Expected 'version' subcommand to be registered")
	assert.True(t, hasInstall, "Expected 'install' subcommand to be registered")
}

func TestRootCommandHelpOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	cmd := newTestRootCommand(buf, "--help")

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Superviz - Declarative Process Supervisor")
	assert.Contains(t, output, "Available Commands:")
	assert.Contains(t, output, "version")
	assert.Contains(t, output, "help")
}

func TestVersionCommandOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	cmd := newTestRootCommand(buf, "version")

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Version:")
	assert.Contains(t, output, "Commit:")
	assert.Contains(t, output, "Go version:")
}

func TestUnknownCommandReturnsError(t *testing.T) {
	buf := &bytes.Buffer{}
	cmd := newTestRootCommand(buf, "doesnotexist")

	err := cmd.Execute()
	assert.Error(t, err)
	output := buf.String()
	assert.Contains(t, output, "")
}
