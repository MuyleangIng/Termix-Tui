package cli

import "testing"

func TestIsTUIExecutable(t *testing.T) {
	tests := map[string]bool{
		"termix":                          false,
		"termix.exe":                      false,
		"termix-tui":                      true,
		"termix-tui.exe":                  true,
		`C:\Users\Dev\bin\termix-tui.exe`: true,
	}
	for name, want := range tests {
		if got := isTUIExecutable(name); got != want {
			t.Fatalf("isTUIExecutable(%q) = %v, want %v", name, got, want)
		}
	}
}
