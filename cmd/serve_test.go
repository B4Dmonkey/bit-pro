package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/task"
)

func fastTick(t *testing.T) {
	t.Helper()

	tick, runner := serveTick, serveRunner

	serveTick = 5 * time.Millisecond
	serveRunner = func(context.Context, string, string, ...string) (string, int, error) {
		return "[]", 0, nil
	}

	t.Cleanup(func() { serveTick, serveRunner = tick, runner })
}

func TestServeCmd_WritesProjectCounts(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")

	dir := t.TempDir()

	store := task.New(filepath.Join(dir, ".bit"))
	if err := store.Save(&task.Task{ID: "ACME-1", Title: "a track", Status: task.StatusTodo}); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	seedProject(t, orm.CreateProjectParams{Path: dir, Code: "ACME"})

	fastTick(t)

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	if _, err := runWithContext(t, ctx, serveCmdUse, serveDaemonCmdUse); err != nil {
		t.Fatalf("bp serve daemon returned error: %v", err)
	}

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	defer sqlDB.Close()

	var backlog, todo, done, completed int64
	if err := sqlDB.QueryRowContext(t.Context(),
		"SELECT backlog, todo, done, completed FROM projects WHERE code = ?", "ACME",
	).Scan(&backlog, &todo, &done, &completed); err != nil {
		if err == sql.ErrNoRows {
			t.Fatal("project row not found")
		}

		t.Fatalf("Scan() returned error: %v", err)
	}

	if backlog != 1 || todo != 0 || done != 0 || completed != 0 {
		t.Errorf("counts = (%d, %d, %d, %d), want (1, 0, 0, 0)", backlog, todo, done, completed)
	}
}

func TestServeCmd_CountsBacklogAndTodo(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")

	dir := t.TempDir()

	store := task.New(filepath.Join(dir, ".bit"))
	for _, tr := range []*task.Task{
		{ID: "ACME-1", Title: "unapproved track", Status: task.StatusTodo, Approved: false},
		{ID: "ACME-2", Title: "approved track a", Status: task.StatusTodo, Approved: true},
		{ID: "ACME-3", Title: "approved track b", Status: task.StatusTodo, Approved: true},
		{ID: "ACME-3.1", Title: "bar under acme-3", Status: task.StatusTodo, Approved: true},
	} {
		if err := store.Save(tr); err != nil {
			t.Fatalf("Save(%s) returned error: %v", tr.ID, err)
		}
	}

	seedProject(t, orm.CreateProjectParams{Path: dir, Code: "ACME"})

	fastTick(t)

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	if _, err := runWithContext(t, ctx, serveCmdUse, serveDaemonCmdUse); err != nil {
		t.Fatalf("bp serve daemon returned error: %v", err)
	}

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	defer sqlDB.Close()

	var backlog, todo, done, completed int64
	if err := sqlDB.QueryRowContext(t.Context(),
		"SELECT backlog, todo, done, completed FROM projects WHERE code = ?", "ACME",
	).Scan(&backlog, &todo, &done, &completed); err != nil {
		if err == sql.ErrNoRows {
			t.Fatal("project row not found")
		}

		t.Fatalf("Scan() returned error: %v", err)
	}

	if backlog != 1 || todo != 2 || done != 0 || completed != 0 {
		t.Errorf("counts = (%d, %d, %d, %d), want (1, 2, 0, 0)", backlog, todo, done, completed)
	}
}

func TestServeCmd_ReturnsWhenContextCancelled(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	out, err := runWithContext(t, ctx, serveCmdUse, serveDaemonCmdUse)
	if err != nil {
		t.Fatalf("bp serve daemon returned error: %v", err)
	}

	if strings.Contains(out, "tick") {
		t.Errorf("bp serve daemon logged a tick before it was cancelled:\n%s", out)
	}
}

func TestServeCmd_DaemonIsListedInServeHelp(t *testing.T) {
	out := mustRun(t, "serve", "--help")

	if !strings.Contains(out, "daemon") {
		t.Errorf("bp serve --help output does not list daemon:\n%s", out)
	}
}

func TestServeMCPCmd_IsListedInServeHelp(t *testing.T) {
	out := mustRun(t, "serve", "--help")

	if !strings.Contains(out, "mcp") {
		t.Errorf("bp serve --help output does not list mcp:\n%s", out)
	}
}

func TestServeCmd_IsListedInHelp(t *testing.T) {
	out := mustRun(t, "serve", "--help")

	if !strings.Contains(out, "daemon") {
		t.Errorf("bp serve --help output does not list daemon:\n%s", out)
	}
}

func TestServeCmd_TicksOnlyWhenVerbose(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")

	tests := []struct {
		name      string
		args      []string
		wantTicks bool
	}{
		{name: "verbose logs ticks at debug", args: []string{serveCmdUse, serveDaemonCmdUse, "-v"}, wantTicks: true},
		{name: "default logs no ticks", args: []string{serveCmdUse, serveDaemonCmdUse}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fastTick(t)

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
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")

	fastTick(t)

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	out, err := runWithContext(t, ctx, serveCmdUse, serveDaemonCmdUse, "-v")
	if err != nil {
		t.Fatalf("bp serve daemon -v returned error: %v", err)
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

func TestServeCmd_SkipsAProjectItCannotRead(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, aaaDir string)
	}{
		{
			name: "unparseable task file",
			setup: func(t *testing.T, aaaDir string) {
				t.Helper()

				if err := os.MkdirAll(filepath.Join(aaaDir, ".bit", "tasks"), 0o755); err != nil {
					t.Fatalf("MkdirAll: %v", err)
				}

				p := filepath.Join(aaaDir, ".bit", "tasks", "AAA-1.md")
				if err := os.WriteFile(p, []byte("not frontmatter"), 0o600); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
			},
		},
		{
			name:  "no .bit/ directory",
			setup: func(*testing.T, string) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runSkipTest(t, tt.setup)
		})
	}
}

func runSkipTest(t *testing.T, setup func(*testing.T, string)) {
	t.Helper()

	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")

	aaaDir := t.TempDir()
	zzzDir := t.TempDir()

	setup(t, aaaDir)

	tr := &task.Task{ID: "ZZZ-1", Title: "a track", Status: task.StatusTodo, Approved: true}
	if err := task.New(filepath.Join(zzzDir, ".bit")).Save(tr); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	seedProject(t, orm.CreateProjectParams{Path: aaaDir, Code: "AAA"})
	seedProject(t, orm.CreateProjectParams{Path: zzzDir, Code: "ZZZ"})

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	if _, err := sqlDB.ExecContext(t.Context(),
		"UPDATE projects SET backlog = 9 WHERE code = ?", "AAA",
	); err != nil {
		sqlDB.Close()
		t.Fatalf("UPDATE backlog: %v", err)
	}

	sqlDB.Close()

	fastTick(t)

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	if _, err := runWithContext(t, ctx, serveCmdUse, serveDaemonCmdUse); err != nil {
		t.Fatalf("bp serve daemon returned error: %v", err)
	}

	sqlDB2, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	defer sqlDB2.Close()

	var aaaBacklog int64
	if err := sqlDB2.QueryRowContext(t.Context(),
		"SELECT backlog FROM projects WHERE code = ?", "AAA",
	).Scan(&aaaBacklog); err != nil {
		t.Fatalf("Scan AAA backlog: %v", err)
	}

	var zzzTodo int64
	if err := sqlDB2.QueryRowContext(t.Context(),
		"SELECT todo FROM projects WHERE code = ?", "ZZZ",
	).Scan(&zzzTodo); err != nil {
		t.Fatalf("Scan ZZZ todo: %v", err)
	}

	if aaaBacklog != 9 {
		t.Errorf("AAA.backlog = %d, want 9 (project should be skipped, not zeroed)", aaaBacklog)
	}

	if zzzTodo != 1 {
		t.Errorf("ZZZ.todo = %d, want 1 (tick should continue past broken AAA)", zzzTodo)
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
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	out, err := runWithContext(t, ctx, serveCmdUse, serveDaemonCmdUse)
	if err != nil {
		t.Fatalf("bp serve daemon returned error: %v", err)
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
