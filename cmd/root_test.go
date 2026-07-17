package cmd

import (
	"strings"
	"testing"
)

func TestRootCmd_Help(t *testing.T) {
	out := mustRun(t, "--help")

	if !strings.Contains(out, "bit") {
		t.Errorf("help output missing command name %q, got:\n%s", "bit", out)
	}
}

func TestRootCmd_Version(t *testing.T) {
	out := mustRun(t, "--version")

	want := "bit version " + version + "\n"
	if out != want {
		t.Errorf("version output = %q, want %q", out, want)
	}
}

func TestRootCmd_RuntimeErrorOmitsUsage(t *testing.T) {
	initProject(t, "BIT")

	out, err := run(t, "task", "read", "BIT-99")

	if err == nil {
		t.Fatal("Execute() returned nil error, want non-nil for unknown task ID")
	}
	if strings.Contains(out, "Usage:") {
		t.Errorf("output = %q, want no usage text on a runtime failure", out)
	}
}
