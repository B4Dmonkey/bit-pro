package cmd

import (
	"bytes"
	"context"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestRootCmd(t *testing.T) {
	t.Run("help", func(t *testing.T) {
		out := mustRun(t, "--help")

		if !strings.Contains(out, "bp") {
			t.Errorf("help output missing command name %q, got:\n%s", "bp", out)
		}
	})

	t.Run("version", func(t *testing.T) {
		out := mustRun(t, "--version")

		want := "bp version " + version + "\n"
		if out != want {
			t.Errorf("version output = %q, want %q", out, want)
		}
	})

	t.Run("has no instructions command", func(t *testing.T) {
		out, err := run(t, "instructions")
		if err == nil {
			t.Fatalf("bp instructions returned nil error, want an unknown-command error; output:\n%s", out)
		}

		if !strings.Contains(err.Error(), "unknown command") {
			t.Errorf("error = %q, want it to contain %q", err, "unknown command")
		}
	})

	t.Run("runtime error omits usage", func(t *testing.T) {
		initProject(t, "BIT")

		out, err := run(t, "task", "read", "BIT-99")
		if err == nil {
			t.Fatal("Execute() returned nil error, want non-nil for unknown task ID")
		}

		if strings.Contains(out, "Usage:") {
			t.Errorf("output = %q, want no usage text on a runtime failure", out)
		}
	})
}

func TestSignalContext(t *testing.T) {
	t.Run("cancels on termination signals", func(t *testing.T) {
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
	})
}

func TestExecute(t *testing.T) {
	t.Run("behind plugin writes notice to stderr", func(t *testing.T) {
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
	})

	t.Run("no plugin state is silent", func(t *testing.T) {
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
	})

	t.Run("suppressed command writes no notice", func(t *testing.T) {
		prev := pluginState
		pluginState = func() (string, string, bool) { return v010, v020, true }

		t.Cleanup(func() { pluginState = prev })

		root := newRootCmd(func(context.Context, string, ...string) error { return nil })
		root.AddCommand(&cobra.Command{
			Use:         "quiet",
			Annotations: map[string]string{quietAnnotation: quietEnabled},
			RunE:        func(*cobra.Command, []string) error { return nil },
		})

		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}

		root.SetOut(stdout)
		root.SetErr(stderr)
		root.SetIn(strings.NewReader(""))
		root.SetArgs([]string{"quiet"})

		if err := execute(context.Background(), root); err != nil {
			t.Fatalf("execute returned error: %v", err)
		}

		if stderr.String() != "" {
			t.Errorf("stderr = %q, want empty", stderr.String())
		}
	})

	t.Run("fires the refresh", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")

		fired := 0

		prev := refreshMarketplace
		refreshMarketplace = func() { fired++ }

		t.Cleanup(func() { refreshMarketplace = prev })

		stdout, stderr, err := runSplit(t, "task", "list")
		if err != nil {
			t.Fatalf("bp task list returned error: %v", err)
		}

		if fired != 1 {
			t.Errorf("refresh fired %d times, want 1", fired)
		}

		if stderr != "" {
			t.Errorf("stderr = %q, want it empty", stderr)
		}

		if !strings.Contains(stdout, "BIT-1") {
			t.Errorf("stdout = %q, want it to contain BIT-1", stdout)
		}
	})
}

func TestSuppressed(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: tuiCmdUse, args: []string{tuiCmdUse}, want: true},
		{name: "serve mcp", args: []string{serveCmdUse, serveMCPCmdUse}, want: true},
		{name: "task list", args: []string{"task", "list"}, want: false},
		{name: "root", args: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := newRootCmd(func(context.Context, string, ...string) error { return nil })

			cmd, _, err := root.Find(tt.args)
			if err != nil {
				t.Fatalf("Find(%v) returned error: %v", tt.args, err)
			}

			if got := suppressed(cmd); got != tt.want {
				t.Errorf("suppressed(%q) = %v, want %v", cmd.CommandPath(), got, tt.want)
			}
		})
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
