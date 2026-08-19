package cmd

import (
	"context"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/launchd"
)

const (
	notRunning = "not running\n"
	runningPID = "running (pid 4242)\n"
	stopped    = "stopped\n"
)

func TestStatusCmd_ReportsNotRunning(t *testing.T) {
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
			var calls []string

			lc := func(_ context.Context, name string, args ...string) (string, int, error) {
				calls = append(calls, strings.Join(append([]string{name}, args...), " "))

				if len(args) > 0 && args[0] == "print-disabled" {
					return disabledStore(), 0, nil
				}

				return tt.out, tt.code, nil
			}

			out, err := runWithLaunchd(t, lc, statusCmdUse)
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

func disabledStore(entries ...string) string {
	return "{\n\t" + strings.Join(entries, "\n\t") + "\n}"
}

func printDisabledCall() string {
	return "launchctl print-disabled gui/" + strconv.Itoa(os.Getuid())
}

func listCall() string {
	return "launchctl list " + launchd.Label
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
			var calls []string

			lc := func(_ context.Context, name string, args ...string) (string, int, error) {
				calls = append(calls, strings.Join(append([]string{name}, args...), " "))

				if len(args) > 0 && args[0] == "print-disabled" {
					return tt.store, 0, nil
				}

				return launchctlDict, 0, nil
			}

			out, err := runWithLaunchd(t, lc, statusCmdUse)
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
