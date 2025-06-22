// internal/services/repository/arch/handler.go
package arch

import (
	"context"
	"fmt"
	"io"

	"github.com/kodflow/superviz.io/internal/providers"
	"github.com/kodflow/superviz.io/internal/services/repository/common"
	"github.com/kodflow/superviz.io/internal/transports/ssh"
)

// Handler handles Arch repository setup.
type Handler struct {
	client   ssh.Client
	sudo     *common.SudoHelper
	provider providers.InstallProvider
}

// NewHandler creates a new Arch repository handler.
func NewHandler(client ssh.Client, provider providers.InstallProvider) *Handler {
	return &Handler{
		client:   client,
		sudo:     common.NewSudoHelper(client),
		provider: provider,
	}
}

// Setup sets up the repository for Arch systems.
func (h *Handler) Setup(ctx context.Context, writer io.Writer) error {
	if _, err := fmt.Fprintf(writer, "Setting up Pacman repository...\n"); err != nil {
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

	// Get GPG key ID from provider
	gpgKeyID := h.provider.GetGPGKeyID()

	commands := []string{
		// Create temporary config addition
		"cat >> /tmp/superviz-pacman.conf << 'EOF'\n\n[superviz]\nServer = https://repo.superviz.io/arch/$arch\nEOF",

		// Add to pacman.conf
		"cat /tmp/superviz-pacman.conf >> /etc/pacman.conf",
		"rm /tmp/superviz-pacman.conf",

		// Import key
		fmt.Sprintf("pacman-key --recv-keys %s", gpgKeyID),
		fmt.Sprintf("pacman-key --lsign-key %s", gpgKeyID),

		// Update package database
		"pacman -Sy",
	}

	// Apply sudo prefix where needed
	commands = h.sudo.AddPrefix(commands, needSudo)

	executor := common.NewCommandExecutor(h.client)
	return executor.Execute(ctx, commands, writer)
}
