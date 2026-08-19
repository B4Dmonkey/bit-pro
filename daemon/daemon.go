package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const Label = "com.github.b4dmonkey.bit-pro"

func domain() string {
	return "gui/" + strconv.Itoa(os.Getuid())
}

type Runner func(ctx context.Context, name string, args ...string) (out string, code int, err error)

func ExecRunner(ctx context.Context, name string, args ...string) (string, int, error) {
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
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

type State int

const (
	StateNotRunning State = iota
	StateRunning
	StateStopped
)

func (s State) String() string {
	switch s {
	case StateRunning:
		return "running"
	case StateStopped:
		return "stopped"
	case StateNotRunning:
		return "not running"
	}

	return "not running"
}
