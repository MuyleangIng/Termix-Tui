//go:build windows

package shell

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func documentsDir(home string) string {
	if userHome, err := os.UserHomeDir(); err == nil && !samePath(home, userHome) {
		return filepath.Join(home, "Documents")
	}
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Explorer\User Shell Folders`, registry.QUERY_VALUE)
	if err != nil {
		return filepath.Join(home, "Documents")
	}
	defer key.Close()

	value, _, err := key.GetStringValue("Personal")
	if err != nil || strings.TrimSpace(value) == "" {
		return filepath.Join(home, "Documents")
	}
	value = os.ExpandEnv(value)
	if !filepath.IsAbs(value) {
		return filepath.Join(home, "Documents")
	}
	return filepath.Clean(value)
}

func samePath(left, right string) bool {
	leftAbs, leftErr := filepath.Abs(left)
	rightAbs, rightErr := filepath.Abs(right)
	if leftErr != nil || rightErr != nil {
		return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
	}
	return strings.EqualFold(filepath.Clean(leftAbs), filepath.Clean(rightAbs))
}
