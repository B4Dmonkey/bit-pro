package cmd

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRootCmd_Help(t *testing.T) {
	out := mustRun(t, "--help")

	if !strings.Contains(out, "bp") {
		t.Errorf("help output missing command name %q, got:\n%s", "bp", out)
	}
}

func TestRootCmd_Version(t *testing.T) {
	out := mustRun(t, "--version")

	want := "bp version " + version + "\n"
	if out != want {
		t.Errorf("version output = %q, want %q", out, want)
	}
}

func TestBitDirEnvVar_RoutesListToCanonicalDir(t *testing.T) {
	dir1 := initProject(t, "BIT")
	createTask(t, "Track", "...")

	t.Chdir(t.TempDir())
	t.Setenv("BIT_DIR", filepath.Join(dir1, ".bit"))

	out := mustRun(t, "task", "list")

	if !strings.Contains(out, "BIT-1") {
		t.Errorf("output = %q, want output to contain BIT-1 from dir routed via BIT_DIR", out)
	}
}

func TestBitDirEnvVar_DefaultIsRelativeDotBit(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")

	out := mustRun(t, "task", "list")

	if !strings.Contains(out, "BIT-1") {
		t.Errorf("output = %q, want output to contain BIT-1 from default .bit dir", out)
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
