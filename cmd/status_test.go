package cmd

import (
	"context"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/daemon"
	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
)

const (
	printDisabled = "print-disabled"
	listSubcmd    = "list"

	notRunning = "not running\n"
	runningPID = "running (pid 4242)\n"
	stopped    = "stopped\n"
)

func TestStatusCmd_ReportsNotRunning(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")

	out, err := run(t, statusCmdUse)
	if err != nil {
		t.Fatalf("bp status returned error: %v", err)
	}

	if out != notRunning {
		t.Errorf("bp status output = %q, want %q", out, notRunning)
	}
}

const launchctlDict = `{
	"Label" = "com.github.b4dmonkey.bit-pro";
	"OnDemand" = false;
	"LastExitStatus" = 0;
	"PID" = 4242;
	"Program" = "/Users/operator/go/bin/bp";
	"ProgramArguments" = (
		"/Users/operator/go/bin/bp";
		"serve";
	);
};`

func TestStatusCmd_ReportsWhatLaunchctlSays(t *testing.T) {
	tests := []struct {
		name string
		out  string
		code int
		want string
	}{
		{
			name: "loaded with a pid",
			out:  launchctlDict,
			code: 0,
			want: runningPID,
		},
		{
			name: "loaded without a pid",
			out:  strings.ReplaceAll(launchctlDict, "\t\"PID\" = 4242;\n", ""),
			code: 0,
			want: notRunning,
		},
		{
			name: "not loaded",
			out:  `Could not find service "com.github.b4dmonkey.bit-pro" in domain for port`,
			code: 113,
			want: notRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("XDG_DATA_HOME", "")

			var calls []string

			lc := func(_ context.Context, name string, args ...string) (string, int, error) {
				calls = append(calls, strings.Join(append([]string{name}, args...), " "))

				if len(args) > 0 && args[0] == printDisabled {
					return disabledStore(), 0, nil
				}

				return tt.out, tt.code, nil
			}

			out, err := runWithDaemon(t, lc, statusCmdUse)
			if err != nil {
				t.Fatalf("bp status returned error: %v", err)
			}

			if out != tt.want {
				t.Errorf("bp status output = %q, want %q", out, tt.want)
			}

			want := []string{printDisabledCall(), listCall()}
			if !slices.Equal(calls, want) {
				t.Errorf("launchctl calls = %v, want %v", calls, want)
			}
		})
	}
}

func TestStatusCmd_ShowsProjectCounts(t *testing.T) {
	tests := []struct {
		name  string
		lc    daemon.Runner
		state string
	}{
		{name: "loaded with a pid", lc: loadedWithPID, state: runningPID},
		{name: "no daemon loaded", lc: nothingLoaded, state: notRunning},
		{name: "disabled in the store", lc: labelDisabled, state: stopped},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("XDG_DATA_HOME", "")

			seedCountedProjects(t)

			out, err := runWithDaemon(t, tt.lc, statusCmdUse)
			if err != nil {
				t.Fatalf("bp status returned error: %v", err)
			}

			want := normalizeSpaces(tt.state) + " ACE backlog:2 todo:1 done:4 MID backlog:0 todo:3 done:12"
			if got := normalizeSpaces(out); got != want {
				t.Errorf("bp status output = %q, want %q", got, want)
			}

			if strings.Contains(out, "completed:") {
				t.Errorf("bp status output = %q, want no completed: column", out)
			}
		})
	}
}

func loadedWithPID(_ context.Context, _ string, args ...string) (string, int, error) {
	if len(args) > 0 && args[0] == printDisabled {
		return disabledStore(), 0, nil
	}

	return launchctlDict, 0, nil
}

func labelDisabled(_ context.Context, _ string, args ...string) (string, int, error) {
	if len(args) > 0 && args[0] == printDisabled {
		return disabledStore(`"` + daemon.Label + `" => disabled`), 0, nil
	}

	return launchctlDict, 0, nil
}

func seedCountedProjects(t *testing.T) {
	t.Helper()

	seedProject(t, orm.CreateProjectParams{Path: "/tmp/status-ace", Code: aceCode})
	seedProject(t, orm.CreateProjectParams{Path: "/tmp/status-mid", Code: midCode})

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	defer sqlDB.Close()

	q := orm.New(sqlDB)

	projects, err := q.ListProjects(t.Context())
	if err != nil {
		t.Fatalf("ListProjects() returned error: %v", err)
	}

	for _, p := range projects {
		counts := orm.UpdateProjectCountsParams{ID: p.ID}

		switch p.Code {
		case aceCode:
			counts.Backlog, counts.Todo, counts.Done, counts.Completed = 2, 1, 4, 7
		case midCode:
			counts.Backlog, counts.Todo, counts.Done, counts.Completed = 0, 3, 12, 2
		}

		if err := q.UpdateProjectCounts(t.Context(), counts); err != nil {
			t.Fatalf("UpdateProjectCounts(%s) returned error: %v", p.Code, err)
		}
	}
}

func disabledStore(entries ...string) string {
	return "{\n\t" + strings.Join(entries, "\n\t") + "\n}"
}

func printDisabledCall() string {
	return "launchctl print-disabled gui/" + strconv.Itoa(os.Getuid())
}

func listCall() string {
	return "launchctl list " + daemon.Label
}

func TestStatusCmd_ReportsStoppedFromTheDisabledStore(t *testing.T) {
	tests := []struct {
		name  string
		store string
		want  string
	}{
		{
			name:  "the label is disabled",
			store: disabledStore(`"com.apple.mdworker" => enabled`, `"com.github.b4dmonkey.bit-pro" => disabled`),
			want:  stopped,
		},
		{
			name:  "the label was re-enabled",
			store: disabledStore(`"com.github.b4dmonkey.bit-pro" => enabled`),
			want:  runningPID,
		},
		{
			name:  "the label is absent",
			store: disabledStore(`"com.apple.mdworker" => enabled`, `"com.github.b4dmonkey.bit-pro-other" => disabled`),
			want:  runningPID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("XDG_DATA_HOME", "")

			var calls []string

			lc := func(_ context.Context, name string, args ...string) (string, int, error) {
				calls = append(calls, strings.Join(append([]string{name}, args...), " "))

				if len(args) > 0 && args[0] == printDisabled {
					return tt.store, 0, nil
				}

				return launchctlDict, 0, nil
			}

			out, err := runWithDaemon(t, lc, statusCmdUse)
			if err != nil {
				t.Fatalf("bp status returned error: %v", err)
			}

			if out != tt.want {
				t.Errorf("bp status output = %q, want %q", out, tt.want)
			}

			want := []string{printDisabledCall(), listCall()}
			if tt.want == stopped {
				want = []string{printDisabledCall()}
			}

			if !slices.Equal(calls, want) {
				t.Errorf("launchctl calls = %v, want %v", calls, want)
			}
		})
	}
}
