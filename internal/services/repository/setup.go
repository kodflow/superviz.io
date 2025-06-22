// internal/services/repository/setup.go
package repository

import (
	"context"
	"fmt"
	"io"

	"github.com/kodflow/superviz.io/internal/providers"
	"github.com/kodflow/superviz.io/internal/services/repository/alpine"
	"github.com/kodflow/superviz.io/internal/services/repository/arch"
	"github.com/kodflow/superviz.io/internal/services/repository/debian"
	"github.com/kodflow/superviz.io/internal/services/repository/rhel"
	"github.com/kodflow/superviz.io/internal/transports/ssh"
)

// Setup defines the interface for repository setup operations.
type Setup interface {
	Setup(ctx context.Context, distro string, writer io.Writer) error
}

// setup implements repository setup for different distributions.
type setup struct {
	client   ssh.Client
	provider providers.InstallProvider
}

// NewSetup creates a new repository setup instance.
func NewSetup(client ssh.Client, provider providers.InstallProvider) Setup {
	return &setup{
		client:   client,
		provider: provider,
	}
}

// Setup sets up the repository for the specified distribution.
func (s *setup) Setup(ctx context.Context, distro string, writer io.Writer) error {
	switch distro {
	case "ubuntu", "debian":
		handler := debian.NewHandler(s.client)
		return handler.Setup(ctx, writer)
	case "alpine":
		handler := alpine.NewHandler(s.client)
		return handler.Setup(ctx, writer)
	case "centos", "rhel", "fedora":
		handler := rhel.NewHandler(s.client)
		return handler.Setup(ctx, writer)
	case "arch":
		handler := arch.NewHandler(s.client, s.provider)
		return handler.Setup(ctx, writer)
	default:
		return fmt.Errorf("unsupported distribution: %s", distro)
	}
}
