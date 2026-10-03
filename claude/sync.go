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
