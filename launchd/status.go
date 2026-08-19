package launchd

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
)

var pidPattern = regexp.MustCompile(`"PID"\s*=\s*(\d+)`)

func Status(ctx context.Context, run Runner) (State, int, error) {
	out, code, err := run(ctx, "launchctl", "list", Label)
	if err != nil {
		return StateNotRunning, 0, fmt.Errorf("asking launchd about %s: %w", Label, err)
	}

	if code != 0 {
		return StateNotRunning, 0, nil
	}

	match := pidPattern.FindStringSubmatch(out)
	if match == nil {
		return StateNotRunning, 0, nil
	}

	pid, err := strconv.Atoi(match[1])
	if err != nil {
		return StateNotRunning, 0, fmt.Errorf("reading the pid of %s from %q: %w", Label, match[1], err)
	}

	return StateRunning, pid, nil
}
