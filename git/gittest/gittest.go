package gittest

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func Env() []string {
	env := make([]string, 0, len(os.Environ()))

	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "GIT_") {
			env = append(env, kv)
		}
	}

	return env
}

func Run(t testing.TB, dir string, args ...string) string {
	t.Helper()

	bin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not found")
	}

	cmd := exec.CommandContext(t.Context(), bin, args...)
	cmd.Dir = dir
	cmd.Env = Env()

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String())
}

func Isolate(t testing.TB) {
	t.Helper()

	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		if !strings.HasPrefix(k, "GIT_") {
			continue
		}

		t.Setenv(k, "")

		if err := os.Unsetenv(k); err != nil {
			t.Fatalf("unsetting %s: %v", k, err)
		}
	}
}
