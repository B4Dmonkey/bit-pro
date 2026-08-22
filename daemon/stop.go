package daemon

import (
	"context"
	"fmt"
)

func Bootout(ctx context.Context, run Runner) {
	_, _, _ = run(ctx, "launchctl", "bootout", domain()+"/"+Label)
}

func Stop(ctx context.Context, run Runner) error {
	if _, _, err := run(ctx, "launchctl", "bootout", domain()+"/"+Label); err != nil {
		return fmt.Errorf("booting out %s: %w", Label, err)
	}

	_, code, err := run(ctx, "launchctl", "disable", domain()+"/"+Label)
	if err != nil {
		return fmt.Errorf("disabling %s: %w", Label, err)
	}

	if code != 0 {
		return fmt.Errorf("disabling %s: launchctl exited %d", Label, code)
	}

	return nil
}
