package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/kodflow/superviz.io/internal/cli"
)

func TestExecuteCLI(t *testing.T) {
	// Capture la sortie standard
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Exécute la commande CLI
	cli.Execute()

	// Ferme et récupère la sortie
	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	output := buf.String()
	if output == "" {
		t.Error("Expected output from CLI execution, got empty string")
	}
}
