package cmd

import (
	"strings"
	"syscall"
	"testing"
	"time"
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
