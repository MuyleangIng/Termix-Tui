package terminal

import "testing"

func TestAppTerminalFaceUsesMesloName(t *testing.T) {
	got := appTerminalFace("MesloLGM Nerd Font")
	want := "MesloLGM Nerd Font"
	if got != want {
		t.Fatalf("appTerminalFace() = %q, want %q", got, want)
	}
}

func TestAppleTerminalFaceUsesMesloMonoName(t *testing.T) {
	got := appleTerminalFace("MesloLGM Nerd Font")
	want := "MesloLGM Nerd Font Mono"
	if got != want {
		t.Fatalf("appleTerminalFace() = %q, want %q", got, want)
	}
}
