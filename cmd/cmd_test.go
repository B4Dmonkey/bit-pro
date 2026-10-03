package cmd

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/claude"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/B4Dmonkey/bit-pro/store"
	"github.com/B4Dmonkey/bit-pro/task"
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

	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")

	dir := t.TempDir()
	t.Chdir(dir)

	path, err := project.CanonicalPath(dir)
	if err != nil {
		t.Fatalf("CanonicalPath(%q) returned error: %v", dir, err)
	}

	seedProject(t, orm.CreateProjectParams{Path: path, Code: prefix})

	return dir
}

func projectStore(t *testing.T) *task.Store {
	t.Helper()

	s, err := project.OpenStore(t.Context(), ".")
	if err != nil {
		t.Fatalf("project.OpenStore(.) returned error: %v", err)
	}

	return s
}

func storeDir(t *testing.T) string {
	t.Helper()

	p, err := project.Find(t.Context(), ".")
	if err != nil {
		t.Fatalf("project.Find(.) returned error: %v", err)
	}

	dir, err := store.ProjectDir(p.Code)
	if err != nil {
		t.Fatalf("store.ProjectDir(%q) returned error: %v", p.Code, err)
	}

	return dir
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

	root := newRootCmd(func(context.Context, string, ...string) error { return nil })
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetIn(strings.NewReader(""))
	root.SetArgs(args)

	err := execute(context.Background(), root)

	return stdout.String(), stderr.String(), err
}
