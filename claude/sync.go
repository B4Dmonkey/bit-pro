package claude

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Runner func(ctx context.Context, name string, args ...string) error

func ExecRunner(ctx context.Context, name string, args ...string) error {
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

func SyncPlugin(ctx context.Context, run Runner) error {
	if err := run(ctx, "claude", "plugin", "marketplace", "update", "bit-pro"); err != nil {
		return fmt.Errorf("refreshing the bit-pro marketplace: %w", err)
	}
	if err := run(ctx, "claude", "plugin", "update", "bit@bit-pro", "--scope", "project"); err != nil {
		if err := run(ctx, "claude", "plugin", "install", "bit@bit-pro", "--scope", "project"); err != nil {
			return fmt.Errorf("installing the bit plugin: %w", err)
		}
	}
	return nil
}
