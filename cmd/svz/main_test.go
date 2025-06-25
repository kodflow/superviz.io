package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newStubCmd(out *bytes.Buffer, runErr error) *cobra.Command {
	cmd := &cobra.Command{
		Use: "svz",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(out, "Superviz - Declarative Process Supervisor")
			fmt.Fprintln(out, "Available Commands:")
			return runErr
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(out)
	return cmd
}

func TestRun_OK(t *testing.T) {
	var out bytes.Buffer
	exit := run(newStubCmd(&out, nil), nil)

	require.Equal(t, 0, exit)
	assert.Contains(t, out.String(), "Superviz - Declarative Process Supervisor")
	assert.Contains(t, out.String(), "Available Commands:")
}

func TestRun_KO(t *testing.T) {
	exit := run(newStubCmd(&bytes.Buffer{}, fmt.Errorf("boom")), nil)
	assert.Equal(t, 1, exit)
}

func TestMain(t *testing.T) {
	if os.Getenv("TEST_MAIN") == "1" {
		main() // Appelle os.Exit(0)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestMain$")
	cmd.Env = append(os.Environ(), "TEST_MAIN=1")
	err := cmd.Run()

	if err == nil {
		// os.Exit(0) = success → OK
		return
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		t.Errorf("main exited with code %d, want 0", exitErr.ExitCode())
	} else {
		t.Fatalf("unexpected error: %v", err)
	}
}
