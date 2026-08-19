package launchd

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

func Status(ctx context.Context, run Runner) (State, int, error) {
	disabled, _, err := run(ctx, "launchctl", "print-disabled", domain())
	if err != nil {
		return StateNotRunning, 0, fmt.Errorf("asking launchd which labels are disabled: %w", err)
	}

	if disabledPattern.MatchString(disabled) {
		return StateStopped, 0, nil
	}

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
