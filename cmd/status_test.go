package cmd

import (
	"context"
	"strings"
	"testing"
)

const notRunning = "not running\n"

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
			want: "running (pid 4242)\n",
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

				return tt.out, tt.code, nil
			}

			out, err := runWithLaunchd(t, lc, statusCmdUse)
			if err != nil {
				t.Fatalf("bp status returned error: %v", err)
			}

			if out != tt.want {
				t.Errorf("bp status output = %q, want %q", out, tt.want)
			}

			want := []string{"launchctl list com.github.b4dmonkey.bit-pro"}
			if len(calls) != 1 || calls[0] != want[0] {
				t.Errorf("launchctl calls = %v, want %v", calls, want)
			}
		})
	}
}
