// internal/services/repository/alpine/handler.go
package alpine

import (
	"context"
	"io"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
	"github.com/kodflow/superviz.io/internal/services/repository/common"
)

// Handler handles Alpine repository setup.
//
//	handler := NewHandler(client)
//	err := handler.Setup(ctx, writer)
//
// Handler provides Alpine Linux APK repository configuration
// using the common base handler functionality.
type Handler struct {
	// Base provides common repository setup functionality
	Base *common.BaseHandler
}

// NewHandler creates a new Alpine repository handler.
//
//	client := ssh.NewClient(config)
//	handler := NewHandler(client)
//
// Parameters:
//   - client: ssh.Client SSH client for executing commands
//
// Returns:
//   - handler: *Handler configured Alpine repository handler
func NewHandler(client ssh.Client) *Handler {
	return &Handler{
		Base: common.NewBaseHandler(client),
	}
}

// Setup sets up the repository for Alpine systems.
//
//	handler := NewHandler(client)
//	err := handler.Setup(ctx, os.Stdout)
//
// Setup configures the superviz.io APK repository on Alpine Linux systems
// by adding the repository URL and importing the public key.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - writer: io.Writer for setup progress output
//
// Returns:
//   - err: error if repository setup fails
func (h *Handler) Setup(ctx context.Context, writer io.Writer) error {
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

	return h.Base.ExecuteSetup(ctx, writer, "Setting up APK repository...", commands)
}
