// internal/services/repository/debian/handler.go
package debian

import (
	"context"
	"io"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
	"github.com/kodflow/superviz.io/internal/services/repository/common"
)

// Handler handles Debian/Ubuntu repository setup.
//
//	handler := NewHandler(client)
//	err := handler.Setup(ctx, writer)
//
// Handler provides Debian/Ubuntu APT repository configuration
// using the common base handler functionality.
type Handler struct {
	// Base provides common repository setup functionality
	Base *common.BaseHandler
}

// NewHandler creates a new Debian repository handler.
//
//	client := ssh.NewClient(config)
//	handler := NewHandler(client)
//
// Parameters:
//   - client: ssh.Client SSH client for executing commands
//
// Returns:
//   - handler: *Handler configured Debian repository handler
func NewHandler(client ssh.Client) *Handler {
	return &Handler{
		Base: common.NewBaseHandler(client),
	}
}

// Setup sets up the repository for Debian/Ubuntu systems.
//
//	handler := NewHandler(client)
//	err := handler.Setup(ctx, os.Stdout)
//
// Setup configures the superviz.io APT repository on Debian/Ubuntu systems
// by installing dependencies, adding GPG keys, and configuring the repository.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - writer: io.Writer for setup progress output
//
// Returns:
//   - err: error if repository setup fails
func (h *Handler) Setup(ctx context.Context, writer io.Writer) error {
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

	return h.Base.ExecuteSetup(ctx, writer, "Setting up APT repository...", commands)
}
