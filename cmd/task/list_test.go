package task_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestTaskListCmd(t *testing.T) {
	t.Run("shows newest first", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "First", "...")
		createTask(t, "Second", "...")

		out := mustRun(t, "task", "list")

		want := "BIT-2\ttodo\tSecond\t\t\nBIT-1\ttodo\tFirst\t\t\n"
		if out != want {
			t.Errorf("output = %q, want %q", out, want)
		}
	})

	t.Run("orders numerically not lexically", func(t *testing.T) {
		initProject(t, "BIT")

		for i := 1; i <= 10; i++ {
			createTask(t, fmt.Sprintf("T%d", i), "...")
		}

		out := mustRun(t, "task", "list")

		ids := make([]string, 0, 10)
		for line := range strings.SplitSeq(strings.TrimRight(out, "\n"), "\n") {
			ids = append(ids, strings.Split(line, "\t")[0])
		}

		want := []string{"BIT-10", "BIT-9", "BIT-8", "BIT-7", "BIT-6", "BIT-5", "BIT-4", "BIT-3", "BIT-2", "BIT-1"}
		if !slices.Equal(ids, want) {
			t.Errorf("ID order = %v, want %v", ids, want)
		}
	})

	t.Run("groups bars under their track", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "One", "...")
		createTask(t, "Two", "...")
		mustRun(t, "task", "create", "One.1", "-d", "...", "--parent", "BIT-1")
		mustRun(t, "task", "create", "One.2", "-d", "...", "--parent", "BIT-1")
		mustRun(t, "task", "create", "Two.1", "-d", "...", "--parent", "BIT-2")

		out := mustRun(t, "task", "list")

		ids := make([]string, 0, 5)
		for line := range strings.SplitSeq(strings.TrimRight(out, "\n"), "\n") {
			ids = append(ids, strings.Split(line, "\t")[0])
		}

		want := []string{"BIT-2", "BIT-2.1", "BIT-1", firstBarID, secondBarID}
		if !slices.Equal(ids, want) {
			t.Errorf("ID order = %v, want %v", ids, want)
		}
	})

	t.Run("filters to parent bars", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "One", "...")
		createTask(t, "Two", "...")
		mustRun(t, "task", "create", "One.1", "-d", "...", "--parent", "BIT-1")
		mustRun(t, "task", "create", "One.2", "-d", "...", "--parent", "BIT-1")
		mustRun(t, "task", "create", "Two.1", "-d", "...", "--parent", "BIT-2")

		out := mustRun(t, "task", "list", "--parent", "BIT-1")

		want := firstBarID + "\ttodo\tOne.1\t\t\nBIT-1.2\ttodo\tOne.2\t\t\n"
		if out != want {
			t.Errorf("output = %q, want %q", out, want)
		}
	})

	t.Run("lowercase parent still lists the bars", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")
		mustRun(t, "task", "create", "Bar one", "-d", "...", "--parent", "BIT-1")
		mustRun(t, "task", "create", "Bar two", "-d", "...", "--parent", "BIT-1")

		out := mustRun(t, "task", "list", "--parent", "bit-1")

		want := firstBarID + "\ttodo\tBar one\t\t\nBIT-1.2\ttodo\tBar two\t\t\n"
		if out != want {
			t.Errorf("output = %q, want %q", out, want)
		}
	})

	t.Run("hand edited lowercase order still ranks bars", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")
		mustRun(t, "task", "create", "Bar one", "-d", "...", "--parent", "BIT-1")
		mustRun(t, "task", "create", "Bar two", "-d", "...", "--parent", "BIT-1")

		track := "---\nid: BIT-1\ntitle: Track\nstatus: todo\norder:\n    - bit-1.2\n    - bit-1.1\n---\nHand-edited.\n"
		if err := os.WriteFile(filepath.Join(storeDir(t), "tasks", "BIT-1.md"), []byte(track), 0o600); err != nil {
			t.Fatalf("os.WriteFile(tasks/BIT-1.md) error = %v", err)
		}

		out := mustRun(t, "task", "list", "--parent", "BIT-1")

		want := "BIT-1.2\ttodo\tBar two\t\t\n" + firstBarID + "\ttodo\tBar one\t\t\n"
		if out != want {
			t.Errorf("output = %q, want %q", out, want)
		}
	})

	t.Run("parent with no bars", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Lonely", "...")

		out := mustRun(t, "task", "list", "--parent", "BIT-9")

		if out != "" {
			t.Errorf("output = %q, want empty output for a parent with no bars", out)
		}
	})

	t.Run("shows phase on bars", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")
		mustRun(t, "task", "create", "Bar one", "-d", "...", "--parent", "BIT-1",
			"--phase", "1", "--phase-label", "First")
		mustRun(t, "task", "create", "Bar two", "-d", "...", "--parent", "BIT-1",
			"--phase", "2", "--phase-label", "Second")

		out := mustRun(t, "task", "list")

		want := "BIT-1\ttodo\tTrack\t\t\n" +
			"BIT-1.1\ttodo\tBar one\t\tphase 1 — First\n" +
			"BIT-1.2\ttodo\tBar two\t\tphase 2 — Second\n"
		if out != want {
			t.Errorf("output = %q, want %q", out, want)
		}
	})

	t.Run("empty when no tasks", func(t *testing.T) {
		initProject(t, "BIT")

		out := mustRun(t, "task", "list")

		if out != "" {
			t.Errorf("output = %q, want empty output when no tasks exist", out)
		}
	})

	t.Run("shows approved marker", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")
		approve(t, trackID)

		out := mustRun(t, "task", "list")

		want := "BIT-1\ttodo\tTrack\tapproved\t\n"
		if out != want {
			t.Errorf("output = %q, want %q", out, want)
		}
	})

	t.Run("unapproved shows empty field", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")

		out := mustRun(t, "task", "list")

		want := "BIT-1\ttodo\tTrack\t\t\n"
		if out != want {
			t.Errorf("output = %q, want %q", out, want)
		}
	})
}
