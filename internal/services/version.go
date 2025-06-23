// Package services provides business logic services for superviz.io operations
package services

import (
	"io"

	"github.com/kodflow/superviz.io/internal/providers"
)

// VersionService handles version-related operations and formatting.
//
// VersionService provides methods to retrieve and display version information
// through configurable providers, enabling dependency injection and testing.
type VersionService struct {
	// provider supplies version information data
	provider providers.VersionProvider
}

// NewVersionService creates a new version service with the given provider.
//
// NewVersionService initializes a version service with the specified provider,
// using a default provider if none is provided for convenience.
//
// Parameters:
//   - provider: Version provider instance (nil for default)
//
// Returns:
//   - VersionService instance ready for use
func NewVersionService(provider providers.VersionProvider) *VersionService {
	if provider == nil {
		provider = providers.DefaultVersionProvider() // Uses the singleton
	}
	return &VersionService{
		provider: provider,
	}
}

// GetVersionInfo retrieves version information through the configured provider.
//
// GetVersionInfo provides access to the complete version information
// including version number, build details, and other metadata.
//
// Returns:
//   - VersionInfo containing complete version metadata
func (s *VersionService) GetVersionInfo() providers.VersionInfo {
	return s.provider.GetVersionInfo()
}

// DisplayVersion writes formatted version information to the specified writer.
//
// DisplayVersion formats the version information and writes it to the provided
// writer, suitable for command-line output or logging purposes.
//
// Parameters:
//   - w: Writer to output the formatted version information
//
// Returns:
//   - Error if writing fails or writer is nil
func (s *VersionService) DisplayVersion(w io.Writer) error {
	if w == nil {
		return ErrNilWriter
	}

	info := s.provider.GetVersionInfo()
	_, err := w.Write([]byte(info.Format())) // Plus efficace que fmt.Fprint
	return err
}

// DisplayVersionString returns the formatted version as a string.
//
// DisplayVersionString provides the formatted version information as a string
// for cases where direct string access is needed without writing to a writer.
//
// Returns:
//   - Formatted version string
func (s *VersionService) DisplayVersionString() string {
	return s.provider.GetVersionInfo().Format()
}
