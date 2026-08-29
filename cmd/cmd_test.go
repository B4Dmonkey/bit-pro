package cmd

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/claude"
	"github.com/B4Dmonkey/bit-pro/daemon"
)

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	return runWithStdin(t, "", args...)
}

func runWithStdin(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	return runWithRunner(t, func(context.Context, string, ...string) error { return nil }, stdin, args...)
}

func nothingLoaded(context.Context, string, ...string) (string, int, error) {
	return "", 113, nil
}

func runWithDaemon(t *testing.T, lc daemon.Runner, args ...string) (string, error) {
	t.Helper()

	root := newRootCmd(func(context.Context, string, ...string) error { return nil }, lc)
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetIn(strings.NewReader(""))
	root.SetArgs(args)

	err := root.Execute()

	return out.String(), err
}

func runWithRunner(t *testing.T, run claude.Runner, stdin string, args ...string) (string, error) {
	t.Helper()

	root := newRootCmd(run, nothingLoaded)
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetIn(strings.NewReader(stdin))
	root.SetArgs(args)

	err := root.Execute()

	return out.String(), err
}

func runWithContext(t *testing.T, ctx context.Context, args ...string) (string, error) {
	t.Helper()

	root := newRootCmd(func(context.Context, string, ...string) error { return nil }, nothingLoaded)
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetIn(strings.NewReader(""))
	root.SetArgs(args)

	err := root.ExecuteContext(ctx)

	return out.String(), err
}

func normalizeSpaces(s string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
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

func mcpRegisterCall() []string {
	return []string{claudeBin, serveMCPCmdUse, addCmdUse, "bit", "--", "bp", serveCmdUse, serveMCPCmdUse}
}

func pluginSyncCalls() [][]string {
	return [][]string{
		{claudeBin, "plugin", "marketplace", updateCmd, "bit-pro"},
		{claudeBin, "plugin", updateCmd, "bit@bit-pro", "--scope", "project"},
		mcpRegisterCall(),
	}
}

func createTask(t *testing.T, title, description string) {
	t.Helper()
	mustRun(t, "task", "create", title, "--description", description)
}

func TestMain(m *testing.M) {
	refreshMarketplace = func() {}

	os.Exit(m.Run())
}

func runSplit(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	root := newRootCmd(func(context.Context, string, ...string) error { return nil }, nothingLoaded)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetIn(strings.NewReader(""))
	root.SetArgs(args)

	err := execute(context.Background(), root)

	return stdout.String(), stderr.String(), err
}
