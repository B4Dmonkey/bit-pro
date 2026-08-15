package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/claude"
)

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	return runWithStdin(t, "", args...)
}

func runWithStdin(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	return runWithRunner(t, func(context.Context, string, ...string) error { return nil }, stdin, args...)
}

func runWithRunner(t *testing.T, run claude.Runner, stdin string, args ...string) (string, error) {
	t.Helper()

	root := newRootCmd(run)
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetIn(strings.NewReader(stdin))
	root.SetArgs(args)

	err := root.Execute()

	return out.String(), err
}

func mustRun(t *testing.T, args ...string) string {
	t.Helper()

	out, err := run(t, args...)
	if err != nil {
		t.Fatalf("bp %s returned error: %v", strings.Join(args, " "), err)
	}

	return out
}

func initProject(t *testing.T, prefix string) string {
	t.Helper()

	dir := t.TempDir()
	t.Chdir(dir)
	mustRun(t, "init", "--prefix", prefix)

	return dir
}

func createTask(t *testing.T, title, description string) {
	t.Helper()
	mustRun(t, "task", "create", title, "--description", description)
}
