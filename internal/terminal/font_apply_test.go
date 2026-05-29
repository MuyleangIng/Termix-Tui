package terminal

import "testing"

func TestTerminalFaceUsesMesloMonoName(t *testing.T) {
	got := terminalFace("MesloLGM Nerd Font")
	want := "MesloLGM Nerd Font Mono"
	if got != want {
		t.Fatalf("terminalFace() = %q, want %q", got, want)
	}
}

func TestTerminalFaceUsesCaskaydiaName(t *testing.T) {
	got := terminalFace("Cascadia Code Nerd Font")
	want := "CaskaydiaCove Nerd Font"
	if got != want {
		t.Fatalf("terminalFace() = %q, want %q", got, want)
	}
}
