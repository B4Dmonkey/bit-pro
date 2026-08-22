package task_test

import (
	"bytes"
	"strings"
	"testing"

	taskcmd "github.com/B4Dmonkey/bit-pro/cmd/task"
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

const taskCmdUse = taskcmd.CmdUse

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()

	return runWithStdin(t, "", args...)
}

func runWithStdin(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()

	root := &cobra.Command{
		Use:           "bp",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(taskcmd.NewCmd())

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

	if err := task.New(".bit").SaveConfig(&task.Config{Prefix: prefix}); err != nil {
		t.Fatalf("SaveConfig(%q) returned error: %v", prefix, err)
	}

	return dir
}

func approve(t *testing.T, id string) {
	t.Helper()

	if err := task.New(".bit").SetApproved(id, true); err != nil {
		t.Fatalf("SetApproved(%q, true) returned error: %v", id, err)
	}
}

func createTask(t *testing.T, title, description string) {
	t.Helper()
	mustRun(t, taskCmdUse, "create", title, "--description", description)
}
