package task_test

import (
	"bytes"
	"strings"
	"testing"

	taskcmd "github.com/B4Dmonkey/bit-pro/cmd/task"
	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/B4Dmonkey/bit-pro/store"
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

func seedProject(t *testing.T, params orm.CreateProjectParams) {
	t.Helper()

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	defer sqlDB.Close()

	if err := orm.New(sqlDB).CreateProject(t.Context(), params); err != nil {
		t.Fatalf("CreateProject(%+v) returned error: %v", params, err)
	}
}

func approve(t *testing.T, id string) {
	t.Helper()

	if err := projectStore(t).SetApproved(id, true); err != nil {
		t.Fatalf("SetApproved(%q, true) returned error: %v", id, err)
	}
}

func createTask(t *testing.T, title, description string) {
	t.Helper()
	mustRun(t, taskCmdUse, "create", title, "--description", description)
}
