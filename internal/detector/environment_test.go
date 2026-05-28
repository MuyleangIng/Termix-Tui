package detector

import "testing"

func TestDetectCommandDoesNotRunVersionCommand(t *testing.T) {
	state := detectCommand("missing", "definitely-not-a-real-termix-test-command")
	if state.Installed {
		t.Fatalf("expected missing command to be marked uninstalled")
	}

	found := detectCommand("sentinel", "go")
	if found.Installed && found.Version != "" {
		t.Fatalf("detectCommand should not execute version commands, got version %q", found.Version)
	}
}
