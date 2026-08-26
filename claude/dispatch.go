package claude

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"unicode"
)

type DirRunner func(ctx context.Context, dir, name string, args ...string) (out string, code int, err error)

func ExecDirRunner(ctx context.Context, dir, name string, args ...string) (string, int, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))

	var exitErr *exec.ExitError

	switch {
	case err == nil:
		return text, 0, nil
	case errors.As(err, &exitErr):
		return text, exitErr.ExitCode(), nil
	default:
		return text, 0, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, text)
	}
}

func WorktreeName(trackID, title string) string {
	return trackID + "-" + slug(title)
}

func slug(s string) string {
	var b strings.Builder

	dash := false

	for _, r := range strings.ToLower(s) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if dash && b.Len() > 0 {
				b.WriteByte('-')
			}

			dash = false

			b.WriteRune(r)
		default:
			dash = true
		}
	}

	return b.String()
}

type Agent struct {
	Name string `json:"name"`
	Cwd  string `json:"cwd"`
}

func Agents(ctx context.Context, run DirRunner) ([]Agent, error) {
	out, code, err := run(ctx, "", "claude", "agents", "--json")
	if err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}

	if code != 0 {
		return nil, fmt.Errorf("listing sessions: claude exited %d: %s", code, out)
	}

	var agents []Agent
	if err := json.Unmarshal([]byte(out), &agents); err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}

	return agents, nil
}

func Spawn(ctx context.Context, run DirRunner, dir, name, bar string) error {
	out, code, err := run(ctx, dir, "claude",
		"--bg", "--agent", "bit:bot-dev", "-w", name, "-n", name, "/bit:do "+bar)
	if err != nil {
		return fmt.Errorf("spawning a session for %s: %w", bar, err)
	}

	if code != 0 {
		return fmt.Errorf("spawning a session for %s: claude exited %d: %s", bar, code, out)
	}

	return nil
}
