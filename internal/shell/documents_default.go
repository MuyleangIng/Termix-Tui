//go:build !windows

package shell

import "path/filepath"

func documentsDir(home string) string {
	return filepath.Join(home, "Documents")
}
