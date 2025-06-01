package cli_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/kodflow/superviz.io/internal/cli"
	"github.com/stretchr/testify/assert"
)

func TestExecuteRootCommand(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute() panicked: %v", r)
		}
	}()

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"svz"}

	cmd := cli.GetCLICommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	assert.NoError(t, err, "Expected Execute() to run without error")
}

func TestRootCommandStructureAndSubcommands(t *testing.T) {
	cmd := cli.GetCLICommand()
	assert.Equal(t, "svz", cmd.Use)
	assert.Equal(t, "Superviz - Declarative Process Supervisor", cmd.Short)
	assert.True(t, cmd.DisableAutoGenTag)
	assert.True(t, cmd.DisableFlagsInUseLine)

	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "version" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected 'version' subcommand to be registered")
}

func TestRootCommandHelpOutput(t *testing.T) {
	cmd := cli.GetCLICommand()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)
	output := buf.String()

	assert.Contains(t, output, "Superviz - Declarative Process Supervisor")
	assert.Contains(t, output, "Available Commands:")
	assert.Contains(t, output, "version")
	assert.Contains(t, output, "help")
}

func TestVersionCommandOutput(t *testing.T) {
	cmd := cli.GetCLICommand()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	assert.NoError(t, err)
	output := buf.String()

	assert.Contains(t, output, "Version:")
	assert.Contains(t, output, "Commit:")
	assert.Contains(t, output, "Go version:")
}

func TestUnknownCommandReturnsError(t *testing.T) {
	cmd := cli.GetCLICommand()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"doesnotexist"})

	err := cmd.Execute()
	assert.Error(t, err)
	output := buf.String()
	assert.Contains(t, output, "unknown command")
}
