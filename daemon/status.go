package daemon

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
)

var (
	pidPattern      = regexp.MustCompile(`"PID"\s*=\s*(\d+)`)
	disabledPattern = regexp.MustCompile(regexp.QuoteMeta(`"`+Label+`"`) + `\s*=>\s*disabled`)
)

func listJob(ctx context.Context, run Runner) (bool, int, error) {
	out, code, err := run(ctx, "launchctl", "list", Label)
	if err != nil {
		return false, 0, fmt.Errorf("asking launchd about %s: %w", Label, err)
	}

	if code != 0 {
		return false, 0, nil
	}

	match := pidPattern.FindStringSubmatch(out)
	if match == nil {
		return true, 0, nil
	}

	pid, err := strconv.Atoi(match[1])
	if err != nil {
		return true, 0, fmt.Errorf("reading the pid of %s from %q: %w", Label, match[1], err)
	}

	return true, pid, nil
}

func Status(ctx context.Context, run Runner) (State, int, error) {
	disabled, _, err := run(ctx, "launchctl", "print-disabled", domain())
	if err != nil {
		return StateNotRunning, 0, fmt.Errorf("asking launchd which labels are disabled: %w", err)
	}

	if disabledPattern.MatchString(disabled) {
		return StateStopped, 0, nil
	}

	loaded, pid, err := listJob(ctx, run)
	if err != nil {
		return StateNotRunning, 0, err
	}

	if !loaded || pid == 0 {
		return StateNotRunning, 0, nil
	}

	return StateRunning, pid, nil
}
