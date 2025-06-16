package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// version is the semantic version of the application.
	version = "dev"

	// commit is the Git commit hash of the build.
	commit = "none"

	// date is the build timestamp.
	date = "unknown"

	// builtBy indicates who or what performed the build.
	builtBy = "unknown"

	// goVersion is the version of Go used to compile the binary.
	goVersion = runtime.Version()

	// osArch is the target OS/architecture of the compiled binary.
	osArch = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

// VersionInfo contains metadata about the compiled binary,
// including version, commit hash, build date, builder identity,
// Go version, and target OS/architecture.
type VersionInfo struct {
	Version   string // Version is the semantic version of the application.
	Commit    string // Commit is the Git commit hash of the build.
	BuiltAt   string // BuiltAt is the timestamp when the binary was built.
	BuiltBy   string // BuiltBy is the identifier of the build system or user.
	GoVersion string // GoVersion is the version of Go used to compile the binary.
	OSArch    string // OSArch represents the operating system and architecture.
}

// GetVersionCommand returns a Cobra command that displays version information
// about the CLI application.
//
// Returns:
//
//   - *cobra.Command: the configured Cobra command that prints version info.
func GetVersionCommand() *cobra.Command {
	return versionCmd
}

// getVersionInfo collects the current version metadata and returns it
// as a VersionInfo structure.
//
// Returns:
//
//   - VersionInfo: populated with build-time and runtime information.
func getVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   version,
		Commit:    commit,
		BuiltAt:   date,
		BuiltBy:   builtBy,
		GoVersion: goVersion,
		OSArch:    osArch,
	}
}

// format returns a formatted string representation of the VersionInfo
// suitable for CLI display.
//
// Returns:
//
//   - string: the multi-line formatted version details.
func (vi VersionInfo) format() string {
	return fmt.Sprintf(`Version:       %s
Commit:        %s
Built at:      %s
Built by:      %s
Go version:    %s
OS/Arch:       %s
`, vi.Version, vi.Commit, vi.BuiltAt, vi.BuiltBy, vi.GoVersion, vi.OSArch)
}

// versionCmd is the Cobra command used to display version information
// for the application via the "version" CLI subcommand.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Print the formatted version info to stdout
		_, err := fmt.Fprint(cmd.OutOrStdout(), getVersionInfo().format())
		return err
	},
}
