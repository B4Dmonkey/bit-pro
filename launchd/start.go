package launchd

import (
	"context"
	"fmt"
)

func Start(ctx context.Context, run Runner, plistPath string) (State, int, bool, error) {
	loaded, pid, err := listJob(ctx, run)
	if err != nil {
		return StateNotRunning, 0, false, err
	}

	if loaded && pid != 0 {
		return StateRunning, pid, true, nil
	}

	if _, _, err := run(ctx, "launchctl", "enable", domain()+"/"+Label); err != nil {
		return StateNotRunning, 0, false, fmt.Errorf("enabling %s: %w", Label, err)
	}

	if loaded {
		if _, _, err := run(ctx, "launchctl", "kickstart", domain()+"/"+Label); err != nil {
			return StateNotRunning, 0, false, fmt.Errorf("kickstarting %s: %w", Label, err)
		}
	} else {
		if _, _, err := run(ctx, "launchctl", "bootstrap", domain(), plistPath); err != nil {
			return StateNotRunning, 0, false, fmt.Errorf("bootstrapping %s from %s: %w", Label, plistPath, err)
		}
	}

	state, pid, err := Status(ctx, run)

	return state, pid, false, err
}
