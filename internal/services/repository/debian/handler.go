// internal/services/repository/debian/handler.go
package debian

import (
	"context"
	"fmt"
	"io"

	"github.com/kodflow/superviz.io/internal/services/repository/common"
	"github.com/kodflow/superviz.io/internal/transports/ssh"
)

// Handler handles Debian/Ubuntu repository setup.
type Handler struct {
	client ssh.Client
	sudo   *common.SudoHelper
}

// NewHandler creates a new Debian repository handler.
func NewHandler(client ssh.Client) *Handler {
	return &Handler{
		client: client,
		sudo:   common.NewSudoHelper(client),
	}
}

// Setup sets up the repository for Debian/Ubuntu systems.
func (h *Handler) Setup(ctx context.Context, writer io.Writer) error {
	if _, err := fmt.Fprintf(writer, "Setting up APT repository...\n"); err != nil {
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
		// Install required packages
		"apt update",
		"apt install -y curl gnupg lsb-release",

		// Add GPG key (download to temp, then move with appropriate permissions)
		"curl -fsSL https://repo.superviz.io/gpg -o /tmp/superviz.gpg",
		"gpg --dearmor < /tmp/superviz.gpg > /tmp/superviz.gpg.dearmored",
		"cp /tmp/superviz.gpg.dearmored /usr/share/keyrings/superviz.gpg",
		"rm /tmp/superviz.gpg /tmp/superviz.gpg.dearmored",

		// Add repository
		`echo "deb [signed-by=/usr/share/keyrings/superviz.gpg] https://repo.superviz.io/apt $(lsb_release -cs) main" > /etc/apt/sources.list.d/superviz.list`,

		// Update package list
		"apt update",
	}

	// Apply sudo prefix where needed
	commands = h.sudo.AddPrefix(commands, needSudo)

	executor := common.NewCommandExecutor(h.client)
	return executor.Execute(ctx, commands, writer)
}
