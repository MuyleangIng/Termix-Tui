package toolpath

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func Resolve(name string) (string, error) {
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}
	for _, candidate := range candidates(name) {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("%s not found", name)
}

func candidates(name string) []string {
	home, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(home, ".local", "bin", name),
		filepath.Join("/opt/homebrew/bin", name),
		filepath.Join("/usr/local/bin", name),
		filepath.Join("/usr/bin", name),
		filepath.Join("/bin", name),
	}
	if runtime.GOOS == "windows" {
		exe := name
		if filepath.Ext(exe) == "" {
			exe += ".exe"
		}
		paths = append([]string{
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "WindowsApps", exe),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", name, exe),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", name, "bin", exe),
			filepath.Join(os.Getenv("ProgramFiles"), name, exe),
			filepath.Join(os.Getenv("ProgramFiles"), name, "bin", exe),
		}, paths...)
	}
	out := paths[:0]
	for _, path := range paths {
		if path != "" && filepath.IsAbs(path) {
			out = append(out, filepath.Clean(path))
		}
	}
	return out
}

func Exists(name string) bool {
	_, err := Resolve(name)
	return err == nil
}
