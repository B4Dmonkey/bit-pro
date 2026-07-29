package cmd

import (
	"testing"

	"github.com/B4Dmonkey/bit-pro/assets"
)

func TestInstructionsCmd_PrintsContract(t *testing.T) {
	want, err := assets.FS.ReadFile("bit-cli.md")
	if err != nil {
		t.Fatalf("reading embedded bit-cli.md: %v", err)
	}

	out := mustRun(t, "instructions")

	if out != string(want) {
		t.Errorf("bp instructions printed %d bytes, want the %d-byte embedded contract", len(out), len(want))
	}
}

func TestInstructionsCmd_RejectsArgs(t *testing.T) {
	if _, err := run(t, "instructions", "garbage"); err == nil {
		t.Fatal("bp instructions garbage returned nil error, want a usage error")
	}
}
