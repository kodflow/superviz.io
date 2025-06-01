package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	commit    = "none"
	date      = "unknown"
	builtBy   = "unknown"
	goVersion = runtime.Version()
	osArch    = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

// VersionInfo holds metadata about the build
type VersionInfo struct {
	Version   string
	Commit    string
	BuiltAt   string
	BuiltBy   string
	GoVersion string
	OSArch    string
}

// getVersionInfo returns a VersionInfo struct populated with the current build and runtime metadata.
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

// format returns the formatted output of version info
func (vi VersionInfo) format() string {
	return fmt.Sprintf(`Version:       %s
Commit:        %s
Built at:      %s
Built by:      %s
Go version:    %s
OS/Arch:       %s
`, vi.Version, vi.Commit, vi.BuiltAt, vi.BuiltBy, vi.GoVersion, vi.OSArch)
}

// versionCmd defines the "version" command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := fmt.Fprint(cmd.OutOrStdout(), getVersionInfo().format())
		return err
	},
}

// GetVersionCommand returns a Cobra command that displays version information for the CLI application.
func GetVersionCommand() *cobra.Command {
	return versionCmd
}
