package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestServeCmd_ReturnsWhenContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	out, err := runWithContext(t, ctx, serveCmdUse)
	if err != nil {
		t.Fatalf("bp serve returned error: %v", err)
	}

	if strings.Contains(out, "tick") {
		t.Errorf("bp serve logged a tick before it was cancelled:\n%s", out)
	}
}

func TestServeCmd_IsListedInHelp(t *testing.T) {
	out := mustRun(t, "--help")

	if !strings.Contains(out, "serve") {
		t.Errorf("bp --help output does not list serve:\n%s", out)
	}
}

func TestServeCmd_TicksOnlyWhenVerbose(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantTicks bool
	}{
		{name: "verbose logs ticks at debug", args: []string{serveCmdUse, "-v"}, wantTicks: true},
		{name: "default logs no ticks", args: []string{serveCmdUse}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := serveTick
			serveTick = 5 * time.Millisecond

			t.Cleanup(func() { serveTick = original })

			ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
			defer cancel()

			out, err := runWithContext(t, ctx, tt.args...)
			if err != nil {
				t.Fatalf("bp %s returned error: %v", strings.Join(tt.args, " "), err)
			}

			if !tt.wantTicks {
				if strings.Contains(out, "tick") {
					t.Errorf("bp serve logged a tick at the default level:\n%s", out)
				}

				return
			}

			if got := strings.Count(out, "tick"); got < 2 {
				t.Errorf("bp serve -v logged %d ticks, want at least 2:\n%s", got, out)
			}

			if !strings.Contains(out, "DEBUG") {
				t.Errorf("bp serve -v did not log at debug level:\n%s", out)
			}
		})
	}
}

func TestServeCmd_LogsJSONWhenOutputIsNotATerminal(t *testing.T) {
	original := serveTick
	serveTick = 5 * time.Millisecond

	t.Cleanup(func() { serveTick = original })

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	out, err := runWithContext(t, ctx, serveCmdUse, "-v")
	if err != nil {
		t.Fatalf("bp serve -v returned error: %v", err)
	}

	var ticks int

	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}

		record := map[string]any{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("line %q is not JSON: %v", line, err)
		}

		if record["msg"] == "tick" && record["level"] == "DEBUG" {
			ticks++
		}
	}

	if ticks == 0 {
		t.Errorf("bp serve -v logged no JSON tick records:\n%s", out)
	}
}

func TestNewHandler_PicksEncodingFromTheWriter(t *testing.T) {
	tests := []struct {
		name   string
		writer func(t *testing.T) io.Writer
		want   slog.Handler
	}{
		{
			name:   "a buffer is not a file",
			writer: func(*testing.T) io.Writer { return &bytes.Buffer{} },
			want:   &slog.JSONHandler{},
		},
		{
			name: "a regular file is not a terminal",
			writer: func(t *testing.T) io.Writer {
				f, err := os.CreateTemp(t.TempDir(), "")
				if err != nil {
					t.Fatalf("creating temp file: %v", err)
				}

				t.Cleanup(func() { f.Close() })

				return f
			},
			want: &slog.JSONHandler{},
		},
		{
			name: "a character device is a terminal",
			writer: func(t *testing.T) io.Writer {
				f, err := os.Open(os.DevNull)
				if err != nil {
					t.Fatalf("opening %s: %v", os.DevNull, err)
				}

				t.Cleanup(func() { f.Close() })

				return f
			},
			want: &slog.TextHandler{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newHandler(tt.writer(t), slog.LevelInfo)

			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("newHandler returned %T, want %T", got, tt.want)
			}
		})
	}
}

func TestServeCmd_LogsStartAndStop(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	out, err := runWithContext(t, ctx, serveCmdUse)
	if err != nil {
		t.Fatalf("bp serve returned error: %v", err)
	}

	var msgs []string

	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}

		record := map[string]any{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("line %q is not JSON: %v", line, err)
		}

		if record["level"] != "INFO" {
			t.Errorf("record %q is not at info level", line)
		}

		msgs = append(msgs, record["msg"].(string))
	}

	want := []string{"started", "stopped"}
	if !slices.Equal(msgs, want) {
		t.Errorf("bp serve logged %v, want %v:\n%s", msgs, want, out)
	}
}
