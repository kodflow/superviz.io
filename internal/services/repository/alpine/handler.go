// internal/services/repository/alpine/handler.go
package alpine

import (
	"context"
	"fmt"
	"io"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
	"github.com/kodflow/superviz.io/internal/services/repository/common"
)

// Handler handles Alpine repository setup.
type Handler struct {
	client ssh.Client
	sudo   *common.SudoHelper
}

// NewHandler creates a new Alpine repository handler.
func NewHandler(client ssh.Client) *Handler {
	return &Handler{
		client: client,
		sudo:   common.NewSudoHelper(client),
	}
}

// Setup sets up the repository for Alpine systems.
func (h *Handler) Setup(ctx context.Context, writer io.Writer) error {
	if _, err := fmt.Fprintf(writer, "Setting up APK repository...\n"); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}

	// Detect if sudo is needed
	needSudo, err := h.sudo.IsNeeded(ctx)
	if err != nil {
		return fmt.Errorf("failed to detect sudo requirement: %w", err)
	}

	if needSudo {
		if _, err := fmt.Fprintf(writer, "Using sudo for system operations...\n"); err != nil {
			return fmt.Errorf("failed to write to output: %w", err)
		}
	}

	commands := []string{
		// Add repository
		"echo 'https://repo.superviz.io/alpine/v$(cat /etc/alpine-release | cut -d'.' -f1-2)/main' >> /etc/apk/repositories",

		// Add public key
		"wget -O /tmp/superviz.rsa.pub https://repo.superviz.io/alpine/superviz.rsa.pub",
		"cp /tmp/superviz.rsa.pub /etc/apk/keys/superviz.rsa.pub",
		"rm /tmp/superviz.rsa.pub",

		// Update package index
		"apk update",
	}

	// Apply sudo prefix where needed
	commands = h.sudo.AddPrefix(commands, needSudo)

	executor := common.NewCommandExecutor(h.client)
	return executor.Execute(ctx, commands, writer)
}
