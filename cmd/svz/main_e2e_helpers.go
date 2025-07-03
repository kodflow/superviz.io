//go:build e2e

package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// getSvzBinaryPath returns the path to the svz binary for E2E tests
func getSvzBinaryPath() string {
	// Check for TEST_WORKSPACE and RUNFILES_DIR environment variables (Bazel specific)
	if runfilesDir := os.Getenv("RUNFILES_DIR"); runfilesDir != "" {
		// In Bazel, data dependencies are in runfiles under _main
		binaryPath := filepath.Join(runfilesDir, "_main", "cmd", "svz", "svz_", "svz")
		if _, err := os.Stat(binaryPath); err == nil {
			return binaryPath
		}
	}

	// Get current working directory and check for Bazel test structure
	wd, err := os.Getwd()
	if err == nil {
		// In the debug output, we saw that the working directory contains svz_/
		binaryPath := filepath.Join(wd, "svz_", "svz")
		if _, err := os.Stat(binaryPath); err == nil {
			return binaryPath
		}
	}

	// Get current test directory for relative path discovery
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)

	// Try common paths for Bazel tests
	candidates := []string{
		filepath.Join(testDir, "svz_", "svz"), // Bazel test data path
		filepath.Join(testDir, "svz"),         // Direct path
		"cmd/svz/svz_/svz",                    // Relative to workspace
		"bazel-bin/cmd/svz/svz_/svz",          // Build output
		"./svz",                               // Current directory
		"../svz",                              // Parent directory
		"svz",                                 // From PATH
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			// Test if the binary is executable
			if err := exec.Command(candidate, "--help").Run(); err == nil {
				return candidate
			}

			// If not executable, try to copy to a temporary location and make executable
			tmpDir, err := os.MkdirTemp("", "svz_e2e_")
			if err != nil {
				continue
			}

			tmpBinary := filepath.Join(tmpDir, "svz")
			if err := copyFile(candidate, tmpBinary); err != nil {
				continue
			}

			if err := os.Chmod(tmpBinary, 0755); err == nil {
				// Test if the copied binary works
				if err := exec.Command(tmpBinary, "--help").Run(); err == nil {
					return tmpBinary
				}
			}
		}
	}

	// Fallback to the data dependency path in Bazel
	return "svz"
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// isDockerAvailable checks if Docker is available for E2E tests
func isDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	return cmd.Run() == nil
}
