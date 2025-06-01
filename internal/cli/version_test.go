package cli_test

import (
	"bytes"
	"testing"

	"github.com/kodflow/superviz.io/internal/cli"
	"github.com/stretchr/testify/assert"
)

func TestVersionCommandOutputFields(t *testing.T) {
	buf := &bytes.Buffer{}
	cmd := cli.GetCLICommand()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	assert.NoError(t, err)
	output := buf.String()

	expectedFields := []string{
		"Version:",
		"Commit:",
		"Built at:",
		"Built by:",
		"Go version:",
		"OS/Arch:",
	}

	for _, field := range expectedFields {
		assert.Contains(t, output, field)
	}
}

func TestVersionCommandDefaultValues(t *testing.T) {
	buf := &bytes.Buffer{}
	cmd := cli.GetCLICommand()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	assert.NoError(t, err)
	output := buf.String()

	defaults := map[string]string{
		"Version:":    "dev",
		"Commit:":     "none",
		"Built at:":   "unknown",
		"Built by:":   "unknown",
		"Go version:": "unknown",
		"OS/Arch:":    "unknown",
	}

	for _, val := range defaults {
		assert.Contains(t, output, val)
	}
}
