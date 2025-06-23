// internal/services/repository/rhel/handler.go
package rhel

import (
	"context"
	"fmt"
	"io"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
	"github.com/kodflow/superviz.io/internal/services/repository/common"
)

// Handler handles RHEL/CentOS/Fedora repository setup.
type Handler struct {
	client ssh.Client
	sudo   *common.SudoHelper
}

// NewHandler creates a new RHEL repository handler.
func NewHandler(client ssh.Client) *Handler {
	return &Handler{
		client: client,
		sudo:   common.NewSudoHelper(client),
	}
}

// Setup sets up the repository for RHEL/CentOS/Fedora systems.
func (h *Handler) Setup(ctx context.Context, writer io.Writer) error {
	if _, err := fmt.Fprintf(writer, "Setting up YUM/DNF repository...\n"); err != nil {
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

	repoContent := `[superviz]
name=Superviz.io Repository
baseurl=https://repo.superviz.io/rpm/
enabled=1
gpgcheck=1
gpgkey=https://repo.superviz.io/rpm/RPM-GPG-KEY-superviz`

	commands := []string{
		// Create repository file
		fmt.Sprintf("cat > /tmp/superviz.repo << 'EOF'\n%s\nEOF", repoContent),
		"cp /tmp/superviz.repo /etc/yum.repos.d/superviz.repo",
		"rm /tmp/superviz.repo",

		// Import GPG key
		"rpm --import https://repo.superviz.io/rpm/RPM-GPG-KEY-superviz",

		// Update package cache
		"if command -v dnf >/dev/null 2>&1; then dnf clean all; elif command -v yum >/dev/null 2>&1; then yum clean all; fi",
	}

	// Apply sudo prefix where needed
	commands = h.sudo.AddPrefix(commands, needSudo)

	executor := common.NewCommandExecutor(h.client)
	return executor.Execute(ctx, commands, writer)
}
