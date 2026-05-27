package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProjectStacks(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "go.mod"), "module example.com/app\n")
	write(t, filepath.Join(dir, "package.json"), `{"dependencies":{"next":"latest","react":"latest"}}`)
	write(t, filepath.Join(dir, "vite.config.ts"), "export default {}\n")

	got := Detect(dir)
	want := map[Stack]bool{Go: true, Node: true, Vite: true}
	for _, stack := range got {
		delete(want, stack)
	}
	if len(want) != 0 {
		t.Fatalf("missing stacks: %#v; got %#v", want, got)
	}
}

func write(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}
