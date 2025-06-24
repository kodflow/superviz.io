// internal/services/repository/arch/handler.go
package arch

import (
	"context"
	"fmt"
	"io"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
	"github.com/kodflow/superviz.io/internal/providers"
	"github.com/kodflow/superviz.io/internal/services/repository/common"
)

// Handler handles Arch repository setup.
//
//	handler := NewHandler(client, provider)
//	err := handler.Setup(ctx, writer)
//
// Handler provides Arch Linux Pacman repository configuration
// using the common base handler functionality with provider integration.
type Handler struct {
	// Base provides common repository setup functionality
	Base *common.BaseHandler
	// provider supplies GPG key and repository information
	provider providers.InstallProvider
}

// NewHandler creates a new Arch repository handler.
//
//	client := ssh.NewClient(config)
//	provider := providers.NewInstallProvider()
//	handler := NewHandler(client, provider)
//
// Parameters:
//   - client: ssh.Client SSH client for executing commands
//   - provider: providers.InstallProvider install configuration provider
//
// Returns:
//   - handler: *Handler configured Arch repository handler
func NewHandler(client ssh.Client, provider providers.InstallProvider) *Handler {
	return &Handler{
		Base:     common.NewBaseHandler(client),
		provider: provider,
	}
}

// Setup sets up the repository for Arch systems.
//
//	handler := NewHandler(client, provider)
//	err := handler.Setup(ctx, os.Stdout)
//
// Setup configures the superviz.io Pacman repository on Arch Linux systems
// by adding repository configuration and importing GPG keys.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - writer: io.Writer for setup progress output
//
// Returns:
//   - err: error if repository setup fails
func (h *Handler) Setup(ctx context.Context, writer io.Writer) error {
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

	return h.Base.ExecuteSetup(ctx, writer, "Setting up Pacman repository...", commands)
}
