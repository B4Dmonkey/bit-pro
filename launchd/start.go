package launchd

import (
	"context"
	"fmt"
)

func Start(ctx context.Context, run Runner, plistPath string) (State, int, error) {
	if _, _, err := run(ctx, "launchctl", "enable", domain()+"/"+Label); err != nil {
		return StateNotRunning, 0, fmt.Errorf("enabling %s: %w", Label, err)
	}

	if _, _, err := run(ctx, "launchctl", "bootstrap", domain(), plistPath); err != nil {
		return StateNotRunning, 0, fmt.Errorf("bootstrapping %s from %s: %w", Label, plistPath, err)
	}

	return Status(ctx, run)
}
