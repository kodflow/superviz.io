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
//
//	handler := NewHandler(client)
//	err := handler.Setup(ctx, writer)
//
// Handler provides RHEL/CentOS/Fedora YUM/DNF repository configuration
// using the common base handler functionality.
type Handler struct {
	// Base provides common repository setup functionality
	Base *common.BaseHandler
}

// NewHandler creates a new RHEL repository handler.
//
//	client := ssh.NewClient(config)
//	handler := NewHandler(client)
//
// Parameters:
//   - client: ssh.Client SSH client for executing commands
//
// Returns:
//   - handler: *Handler configured RHEL repository handler
func NewHandler(client ssh.Client) *Handler {
	return &Handler{
		Base: common.NewBaseHandler(client),
	}
}

// Setup sets up the repository for RHEL/CentOS/Fedora systems.
//
//	handler := NewHandler(client)
//	err := handler.Setup(ctx, os.Stdout)
//
// Setup configures the superviz.io YUM/DNF repository on RHEL-based systems
// by creating repository configuration and importing GPG keys.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - writer: io.Writer for setup progress output
//
// Returns:
//   - err: error if repository setup fails
func (h *Handler) Setup(ctx context.Context, writer io.Writer) error {
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

	return h.Base.ExecuteSetup(ctx, writer, "Setting up YUM/DNF repository...", commands)
}
