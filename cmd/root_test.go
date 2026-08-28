package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
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

func TestBitDir_OutsideWorktreeUsesRelativeDotBit(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")

	out := mustRun(t, "task", "list")

	if !strings.Contains(out, "BIT-1") {
		t.Errorf("output = %q, want output to contain BIT-1 from default .bit dir", out)
	}
}

func TestBitDir_InsideClaudeWorktreeResolvesToMainCheckout(t *testing.T) {
	root := initProject(t, "BIT")
	createTask(t, "Track", "...")

	worktree := filepath.Join(root, ".claude", "worktrees", "hazy-pondering-star")
	if err := os.MkdirAll(filepath.Join(worktree, ".bit"), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) returned error: %v", worktree, err)
	}

	t.Chdir(worktree)

	out := mustRun(t, "task", "list")

	if !strings.Contains(out, "BIT-1") {
		t.Errorf("output = %q, want output to contain BIT-1 from the main checkout's .bit", out)
	}
}

func TestBitDir_NestedWorktreeResolvesToOutermostCheckout(t *testing.T) {
	root := initProject(t, "BIT")
	createTask(t, "Track", "...")

	outer := filepath.Join(root, ".claude", "worktrees", "outer")
	if err := os.MkdirAll(filepath.Join(outer, ".bit"), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) returned error: %v", outer, err)
	}

	nested := filepath.Join(outer, ".claude", "worktrees", "inner")
	if err := os.MkdirAll(filepath.Join(nested, ".bit"), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) returned error: %v", nested, err)
	}

	t.Chdir(nested)

	out := mustRun(t, "task", "list")

	if !strings.Contains(out, "BIT-1") {
		t.Errorf("output = %q, want output to contain BIT-1 from the outermost checkout's .bit", out)
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

func TestSignalContext_CancelsOnTerminationSignals(t *testing.T) {
	tests := []struct {
		name string
		sig  syscall.Signal
	}{
		{name: "SIGTERM", sig: syscall.SIGTERM},
		{name: "SIGINT", sig: syscall.SIGINT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, stop := signalContext()
			defer stop()

			if err := syscall.Kill(syscall.Getpid(), tt.sig); err != nil {
				t.Fatalf("Kill(%v) returned error: %v", tt.sig, err)
			}

			select {
			case <-ctx.Done():
			case <-time.After(2 * time.Second):
				t.Fatal("context was not cancelled within 2s of the signal")
			}
		})
	}
}

func TestExecute_BehindPluginWritesNoticeToStderr(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")

	prev := pluginState
	pluginState = func() (string, string, bool) { return v010, v020, true }

	t.Cleanup(func() { pluginState = prev })

	stdout, stderr, err := runSplit(t, "task", "list")
	if err != nil {
		t.Fatalf("bp task list returned error: %v", err)
	}

	want := "bp: bit plugin 0.1.0 → 0.2.0 available — run: claude plugin update bit@bit-pro --scope project\n"
	if stderr != want {
		t.Errorf("stderr = %q, want %q", stderr, want)
	}

	if !strings.Contains(stdout, "BIT-1") {
		t.Errorf("stdout = %q, want it to contain BIT-1", stdout)
	}

	if strings.Contains(stdout, "bp: bit plugin") {
		t.Errorf("stdout = %q, want the notice to stay off stdout", stdout)
	}
}

func TestExecute_NoPluginStateIsSilent(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")

	stdout, stderr, err := runSplit(t, "task", "list")
	if err != nil {
		t.Fatalf("bp task list returned error: %v", err)
	}

	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	if !strings.Contains(stdout, "BIT-1") {
		t.Errorf("stdout = %q, want it to contain BIT-1", stdout)
	}
}

func TestBehind(t *testing.T) {
	tests := []struct {
		name      string
		installed string
		latest    string
		want      bool
	}{
		{name: "minor below", installed: v010, latest: v020, want: true},
		{name: "equal", installed: v010, latest: v010, want: false},
		{name: "minor above", installed: v020, latest: v010, want: false},
		{name: "two-digit minor below", installed: v090, latest: "0.10.0", want: true},
		{name: "two-digit minor above", installed: "0.10.0", latest: v090, want: false},
		{name: "major above", installed: "1.0.0", latest: v090, want: false},
		{name: "unparseable installed", installed: "4ebbe7cd5eff", latest: v010, want: false},
		{name: "empty latest", installed: v010, latest: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := behind(tt.installed, tt.latest); got != tt.want {
				t.Errorf("behind(%q, %q) = %v, want %v", tt.installed, tt.latest, got, tt.want)
			}
		})
	}
}
