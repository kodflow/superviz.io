package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/kodflow/superviz.io/internal/cli"
	"github.com/stretchr/testify/assert"
)

func TestExecuteCLI(t *testing.T) {
	// Capture la sortie standard
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Exécute la commande CLI
	err := cli.GetCLICommand().Execute()
	assert.NoError(t, err, "Expected Execute() to run without error")

	// Ferme et récupère la sortie
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	output := buf.String()
	if output == "" {
		t.Error("Expected output from CLI execution, got empty string")
	}
}
