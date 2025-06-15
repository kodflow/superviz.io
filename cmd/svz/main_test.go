package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainExecution(t *testing.T) {
	// Sauvegarde l'état original de os.Stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Simule les arguments d'entrée (ex: aucun, donc help attendu)
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
		os.Stdout = oldStdout
	}()
	os.Args = []string{"svz"}

	// Appelle main()
	main()

	// Ferme l'écriture et lit la sortie
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Assertions sur la sortie
	assert.NotEmpty(t, output, "Expected help or CLI output")
	assert.Contains(t, output, "Superviz - Declarative Process Supervisor", "Expected help text")
	assert.Contains(t, output, "Available Commands:", "Expected list of commands")
}
