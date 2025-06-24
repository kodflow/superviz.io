// internal/services/repository/common/base_handler.go
package common

import (
	"context"
	"fmt"
	"io"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
)

// BaseHandler provides common functionality for all repository handlers.
//
//	handler := NewBaseHandler(client)
//	err := handler.ExecuteSetup(ctx, writer, "Setting up repository...", commands)
//
// BaseHandler eliminates code duplication by providing shared setup logic
// for all distribution-specific repository handlers.
type BaseHandler struct {
	// client provides SSH connectivity for executing commands
	client ssh.Client
	// sudo handles privilege escalation detection and command modification
	sudo *SudoHelper
}

// NewBaseHandler creates a new base handler with the given SSH client.
//
//	client := ssh.NewClient(config)
//	handler := NewBaseHandler(client)
//
// Parameters:
//   - client: ssh.Client SSH client for executing commands
//
// Returns:
//   - handler: *BaseHandler configured base handler instance
func NewBaseHandler(client ssh.Client) *BaseHandler {
	return &BaseHandler{
		client: client,
		sudo:   NewSudoHelper(client),
	}
}

// ExecuteSetup performs the common setup workflow for repository configuration.
//
//	commands := []string{"apt update", "apt install -y curl"}
//	err := handler.ExecuteSetup(ctx, writer, "Setting up APT repository...", commands)
//
// ExecuteSetup handles the complete repository setup workflow including
// sudo detection, command prefix application, and command execution.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - writer: io.Writer for progress output
//   - setupMessage: string initial setup message to display
//   - commands: []string list of commands to execute
//
// Returns:
//   - err: error if setup fails at any stage
func (h *BaseHandler) ExecuteSetup(ctx context.Context, writer io.Writer, setupMessage string, commands []string) error {
	// Write initial setup message
	if _, err := fmt.Fprintf(writer, "%s\n", setupMessage); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}

	// Detect sudo requirement
	needSudo, err := h.sudo.IsNeeded(ctx)
	if err != nil {
		return fmt.Errorf("failed to detect sudo requirement: %w", err)
	}

	// Notify about sudo usage
	if needSudo {
		if _, err := fmt.Fprintf(writer, "Using sudo for system operations...\n"); err != nil {
			return fmt.Errorf("failed to write to output: %w", err)
		}
	}

	// Apply sudo prefix where needed
	commands = h.sudo.AddPrefix(commands, needSudo)

	// Execute commands
	executor := NewCommandExecutor(h.client)
	return executor.Execute(ctx, commands, writer)
}
