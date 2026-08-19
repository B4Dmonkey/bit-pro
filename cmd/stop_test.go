package cmd

import (
	"context"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/daemon"
)

func bootoutCall() string {
	return "launchctl bootout gui/" + strconv.Itoa(os.Getuid()) + "/" + daemon.Label
}

func disableCall() string {
	return "launchctl disable gui/" + strconv.Itoa(os.Getuid()) + "/" + daemon.Label
}

func TestStopCmd_BootsOutThenDisables(t *testing.T) {
	tests := []struct {
		name string
		out  string
		code int
	}{
		{
			name: "the job is loaded",
		},
		{
			name: "the job is not loaded",
			out:  `Boot-out failed: 3: No such process`,
			code: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string

			lc := func(_ context.Context, name string, args ...string) (string, int, error) {
				calls = append(calls, strings.Join(append([]string{name}, args...), " "))

				if args[0] == "bootout" {
					return tt.out, tt.code, nil
				}

				return "", 0, nil
			}

			out, err := runWithDaemon(t, lc, stopCmdUse)
			if err != nil {
				t.Fatalf("bp stop returned error: %v", err)
			}

			if out != stopped {
				t.Errorf("bp stop output = %q, want %q", out, stopped)
			}

			bootout := slices.Index(calls, bootoutCall())
			disable := slices.Index(calls, disableCall())

			if bootout < 0 || disable < 0 || bootout > disable {
				t.Errorf("launchctl calls = %v, want bootout before disable", calls)
			}
		})
	}
}
