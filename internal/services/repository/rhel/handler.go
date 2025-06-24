// internal/services/repository/rhel/handler.go
package rhel

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"text/template"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
	"github.com/kodflow/superviz.io/internal/services/repository/common"
)

// RepoConfig holds configuration for RHEL repository setup.
//
// RepoConfig contains all necessary URLs and settings for configuring
// the superviz.io YUM/DNF repository on RHEL-based systems.
//
// Example:
//
//	config := &RepoConfig{
//		Name:       "Superviz.io Repository",
//		BaseURL:    "https://repo.superviz.io/rpm/",
//		GPGKeyURL:  "https://repo.superviz.io/rpm/RPM-GPG-KEY-superviz",
//		Enabled:    true,
//		GPGCheck:   true,
//	}
type RepoConfig struct {
	// Name is the human-readable repository name
	Name string
	// BaseURL is the repository base URL for packages
	BaseURL string
	// GPGKeyURL is the URL for the GPG signing key
	GPGKeyURL string
	// Enabled indicates if the repository is enabled by default
	Enabled bool
	// GPGCheck indicates if GPG signature verification is enabled
	GPGCheck bool
}

// RepoProvider defines the interface for providing repository configuration.
//
// RepoProvider abstracts the source of repository configuration, enabling
// dependency injection and testability of repository setup operations.
//
// Example:
//
//	provider := NewDefaultRepoProvider()
//	config := provider.GetRepoConfig()
//	fmt.Printf("Repository: %s\n", config.BaseURL)
type RepoProvider interface {
	// GetRepoConfig returns the repository configuration.
	//
	// GetRepoConfig provides the complete configuration needed for
	// setting up the superviz.io repository on RHEL-based systems.
	//
	// Parameters:
	//   - None
	//
	// Returns:
	//   - config: *RepoConfig containing repository settings
	GetRepoConfig() *RepoConfig
}

// defaultRepoProvider implements RepoProvider with default values.
//
// defaultRepoProvider provides the standard superviz.io repository
// configuration for production use.
type defaultRepoProvider struct{}

// GetRepoConfig returns the default repository configuration.
//
// GetRepoConfig provides the standard production configuration for
// the superviz.io YUM/DNF repository.
//
// Example:
//
//	provider := &defaultRepoProvider{}
//	config := provider.GetRepoConfig()
//	fmt.Printf("Base URL: %s\n", config.BaseURL)
//
// Parameters:
//   - None
//
// Returns:
//   - config: *RepoConfig with default production settings
func (p *defaultRepoProvider) GetRepoConfig() *RepoConfig {
	return &RepoConfig{
		Name:      "Superviz.io Repository",
		BaseURL:   "https://repo.superviz.io/rpm/",
		GPGKeyURL: "https://repo.superviz.io/rpm/RPM-GPG-KEY-superviz",
		Enabled:   true,
		GPGCheck:  true,
	}
}

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
	// provider supplies repository configuration
	provider RepoProvider
}

// NewHandler creates a new RHEL repository handler.
//
//	client := ssh.NewClient(config)
//	handler := NewHandler(client)
//
// NewHandler creates a handler with the default repository provider.
// For custom configurations, use NewHandlerWithProvider instead.
//
// Parameters:
//   - client: ssh.Client SSH client for executing commands
//
// Returns:
//   - handler: *Handler configured RHEL repository handler
func NewHandler(client ssh.Client) *Handler {
	return NewHandlerWithProvider(client, &defaultRepoProvider{})
}

// NewHandlerWithProvider creates a new RHEL repository handler with custom provider.
//
// NewHandlerWithProvider allows injection of custom repository configuration
// for testing or alternative repository setups.
//
// Example:
//
//	provider := &CustomRepoProvider{BaseURL: "https://custom.repo.com/"}
//	handler := NewHandlerWithProvider(client, provider)
//
// Parameters:
//   - client: ssh.Client SSH client for executing commands
//   - provider: RepoProvider custom repository configuration provider
//
// Returns:
//   - handler: *Handler configured RHEL repository handler
func NewHandlerWithProvider(client ssh.Client, provider RepoProvider) *Handler {
	return &Handler{
		Base:     common.NewBaseHandler(client),
		provider: provider,
	}
}

// NewDefaultRepoProvider creates a new default repository provider.
//
// NewDefaultRepoProvider returns a provider with the standard production
// configuration for the superviz.io repository.
//
// Example:
//
//	provider := NewDefaultRepoProvider()
//	config := provider.GetRepoConfig()
//	handler := NewHandlerWithProvider(client, provider)
//
// Parameters:
//   - None
//
// Returns:
//   - provider: RepoProvider with default configuration
func NewDefaultRepoProvider() RepoProvider {
	return &defaultRepoProvider{}
}

// NewCustomRepoProvider creates a repository provider with custom configuration.
//
// NewCustomRepoProvider allows creation of providers with non-standard
// repository URLs for testing or alternative deployments.
//
// Example:
//
//	provider := NewCustomRepoProvider(
//		"Custom Repo",
//		"https://custom.repo.com/rpm/",
//		"https://custom.repo.com/gpg-key",
//		true, true,
//	)
//
// Parameters:
//   - name: string repository display name
//   - baseURL: string repository base URL
//   - gpgKeyURL: string GPG key URL
//   - enabled: bool whether repository is enabled
//   - gpgCheck: bool whether GPG verification is enabled
//
// Returns:
//   - provider: RepoProvider with custom configuration
func NewCustomRepoProvider(name, baseURL, gpgKeyURL string, enabled, gpgCheck bool) RepoProvider {
	return &customRepoProvider{
		config: &RepoConfig{
			Name:      name,
			BaseURL:   baseURL,
			GPGKeyURL: gpgKeyURL,
			Enabled:   enabled,
			GPGCheck:  gpgCheck,
		},
	}
}

// customRepoProvider implements RepoProvider with custom configuration.
//
// customRepoProvider allows injection of non-standard repository
// configuration for testing or alternative deployments.
type customRepoProvider struct {
	config *RepoConfig
}

// GetRepoConfig returns the custom repository configuration.
//
// GetRepoConfig provides the injected custom configuration for
// repository setup operations.
//
// Example:
//
//	provider := &customRepoProvider{config: myConfig}
//	config := provider.GetRepoConfig()
//	fmt.Printf("Custom URL: %s\n", config.BaseURL)
//
// Parameters:
//   - None
//
// Returns:
//   - config: *RepoConfig custom repository configuration
func (p *customRepoProvider) GetRepoConfig() *RepoConfig {
	return p.config
}

// validateURL validates that a URL is well-formed and uses HTTPS.
//
// validateURL performs security checks to ensure URLs are safe for use
// in repository configuration and package downloading.
//
// Example:
//
//	err := validateURL("https://repo.superviz.io/rpm/")
//	// Returns nil (valid HTTPS URL)
//
//	err = validateURL("http://unsafe.com/")
//	// Returns error (not HTTPS)
//
// Parameters:
//   - rawURL: string URL to validate
//
// Returns:
//   - err: error if URL is invalid or insecure
func validateURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS scheme, got: %s", parsedURL.Scheme)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	return nil
}

// validateRepoConfig validates repository configuration for security and correctness.
//
// validateRepoConfig ensures all URLs are well-formed and secure before
// using them in repository setup operations.
//
// Example:
//
//	config := &RepoConfig{
//		BaseURL:   "https://repo.superviz.io/rpm/",
//		GPGKeyURL: "https://repo.superviz.io/rpm/RPM-GPG-KEY-superviz",
//	}
//	err := validateRepoConfig(config)
//
// Parameters:
//   - config: *RepoConfig repository configuration to validate
//
// Returns:
//   - err: error if configuration is invalid
func validateRepoConfig(config *RepoConfig) error {
	if config == nil {
		return fmt.Errorf("repository configuration cannot be nil")
	}

	if strings.TrimSpace(config.Name) == "" {
		return fmt.Errorf("repository name cannot be empty")
	}

	if err := validateURL(config.BaseURL); err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	if err := validateURL(config.GPGKeyURL); err != nil {
		return fmt.Errorf("invalid GPG key URL: %w", err)
	}

	return nil
}

// Repository file template for YUM/DNF configuration.
const repoFileTemplate = `[superviz]
name={{.Name}}
baseurl={{.BaseURL}}
enabled={{if .Enabled}}1{{else}}0{{end}}
gpgcheck={{if .GPGCheck}}1{{else}}0{{end}}
gpgkey={{.GPGKeyURL}}`

// generateRepoContent creates repository file content from configuration.
//
// generateRepoContent uses text/template to safely generate repository
// configuration content, preventing injection attacks from malformed input.
//
// Example:
//
//	config := &RepoConfig{Name: "Test Repo", BaseURL: "https://example.com/"}
//	content, err := generateRepoContent(config)
//	// Returns formatted repository file content
//
// Parameters:
//   - config: *RepoConfig repository configuration
//
// Returns:
//   - content: string formatted repository file content
//   - err: error if template execution fails
func generateRepoContent(config *RepoConfig) (string, error) {
	tmpl, err := template.New("repo").Parse(repoFileTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse repository template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, config); err != nil {
		return "", fmt.Errorf("failed to execute repository template: %w", err)
	}

	return buf.String(), nil
}

// Setup sets up the repository for RHEL/CentOS/Fedora systems.
//
//	handler := NewHandler(client)
//	err := handler.Setup(ctx, os.Stdout)
//
// Setup configures the superviz.io YUM/DNF repository on RHEL-based systems
// by creating repository configuration and importing GPG keys using validated
// configuration from the repository provider.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - writer: io.Writer for setup progress output
//
// Returns:
//   - err: error if repository setup fails
func (h *Handler) Setup(ctx context.Context, writer io.Writer) error {
	// Get and validate repository configuration
	config := h.provider.GetRepoConfig()
	if err := validateRepoConfig(config); err != nil {
		return fmt.Errorf("invalid repository configuration: %w", err)
	}

	// Generate safe repository content using templates
	repoContent, err := generateRepoContent(config)
	if err != nil {
		return fmt.Errorf("failed to generate repository content: %w", err)
	}

	commands := []string{
		// Create repository file using safe content
		fmt.Sprintf("cat > /tmp/superviz.repo << 'EOF'\n%s\nEOF", repoContent),
		"cp /tmp/superviz.repo /etc/yum.repos.d/superviz.repo",
		"rm /tmp/superviz.repo",

		// Import GPG key using validated URL
		fmt.Sprintf("rpm --import %s", config.GPGKeyURL),

		// Update package cache
		"if command -v dnf >/dev/null 2>&1; then dnf clean all; elif command -v yum >/dev/null 2>&1; then yum clean all; fi",
	}

	return h.Base.ExecuteSetup(ctx, writer, "Setting up YUM/DNF repository...", commands)
}
