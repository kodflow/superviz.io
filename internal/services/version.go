package services

import (
	"io"

	"github.com/kodflow/superviz.io/internal/providers"
)

// VersionService handles version-related operations.
type VersionService struct {
	provider providers.VersionProvider
}

// NewVersionService creates a new version service with the given provider.
func NewVersionService(provider providers.VersionProvider) *VersionService {
	if provider == nil {
		provider = providers.DefaultVersionProvider() // Uses the singleton
	}
	return &VersionService{
		provider: provider,
	}
}

// GetVersionInfo retrieves version information through the provider.
func (s *VersionService) GetVersionInfo() providers.VersionInfo {
	return s.provider.GetVersionInfo()
}

// DisplayVersion writes formatted version information to the writer.
func (s *VersionService) DisplayVersion(w io.Writer) error {
	if w == nil {
		return ErrNilWriter
	}

	info := s.provider.GetVersionInfo()
	_, err := w.Write([]byte(info.Format())) // Plus efficace que fmt.Fprint
	return err
}

// DisplayVersionString returns the formatted version as a string.
// Useful when you need the string directly without writing to a writer.
func (s *VersionService) DisplayVersionString() string {
	return s.provider.GetVersionInfo().Format()
}
