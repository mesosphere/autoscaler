package utils

import (
	"os"
	"os/exec"
	"path/filepath"
)

// GetCmdPathPreferVendor returns the path to a binary given the name, preferring binaries vendored by our tooling
func GetCmdPathPreferVendor(name string) string {
	if ex, err := os.Executable(); err == nil {
		vendored := filepath.Join(filepath.Dir(ex), name)
		if _, err := exec.LookPath(vendored); err == nil {
			return vendored
		}
	}
	return name
}
