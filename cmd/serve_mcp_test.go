package cmd

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/B4Dmonkey/bit-pro/task"
)

const (
	testTrackID = "FOO-1"
	testTitle   = "mcp test track"
	testBody    = "the body"
	testClient  = "test"

	testNewTrackID = "BIT-1"

	testBarID      = "FOO-1.1"
	testBarTitle   = "a bar"
	testPhaseLabel = "Read surface"

	testSecondBarID    = "FOO-1.2"
	testSecondBarTitle = "a later bar"
	testOtherTrackID   = "FOO-2"
	testOtherTitle     = "another track"
	testOtherBarID     = "FOO-2.1"
	testOtherBarTitle  = "another track's bar"

	testParentKey = "parent"
	testCommitKey = "commit"
	testBranchKey = "branch"
	testBarSHA    = "6a1d3459c0ffee000000000000000000000000ab"
	testTitleKey  = "title"
	testStatusKey = "status"
	testPhaseKey  = "phase"

	testPhaseLabelKey = "phase_label"
	testApprovedKey   = "approved"
	testBodyKey       = "body"

	testTrackSentence = "A track is a top-level task"
	testBarIDExample  = "BIT-7.3"

	testRevokingFields       = "title, body, phase or phase_label revokes it"
	testTodoRevokes          = "Writing status todo revokes approval"
	testForwardKeepsApproval = "a forward move to doing or done keeps approval"
	testGitKeepsApproval     = "Sending commit or branch keeps approval"

	testNoCascade     = "does not cascade"
	testCallerRollsUp = "sets the track's status in a separate call"
)

func TestTaskReadHandler(t *testing.T) {
	t.Run("returns structured fields", func(t *testing.T) {
		dir := t.TempDir()

		seedTasks(t, dir, &task.Task{
			ID: testTrackID, Title: testTitle, Status: task.StatusTodo, Body: testBody,
		})

		got := callTool(t, mcpSession(t, dir), taskReadTool, map[string]any{"id": testTrackID})

		if got["id"] != testTrackID {
			t.Errorf("id = %v, want %s", got["id"], testTrackID)
		}

		if got[testTitleKey] != testTitle {
			t.Errorf("title = %v, want %s", got[testTitleKey], testTitle)
		}

		if got["status"] != "todo" {
			t.Errorf("status = %v, want todo", got["status"])
		}

		if got[testApprovedKey] != false {
			t.Errorf("approved = %v, want false", got[testApprovedKey])
		}

		if got[testBodyKey] != testBody {
			t.Errorf("body = %v, want %s", got[testBodyKey], testBody)
		}

		if got["parent"] != "" {
			t.Errorf("parent = %v, want empty string", got["parent"])
		}
	})

	t.Run("returns parent for bar", func(t *testing.T) {
		dir := t.TempDir()

		seedTasks(t, dir,
			&task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusTodo},
			&task.Task{ID: testBarID, Title: testBarTitle, Status: task.StatusTodo},
		)

		got := callTool(t, mcpSession(t, dir), taskReadTool, map[string]any{"id": testBarID})

		if got["parent"] != testTrackID {
			t.Errorf("parent = %v, want %s", got["parent"], testTrackID)
		}
	})

	t.Run("returns commit and branch", func(t *testing.T) {
		dir := t.TempDir()

		seedGitBars(t, dir)

		session := mcpSession(t, dir)

		tests := []struct {
			name       string
			id         string
			wantCommit string
			wantBranch string
		}{
			{name: "committed bar", id: testBarID, wantCommit: testBarSHA, wantBranch: "v2"},
			{name: "uncommitted bar", id: testSecondBarID, wantCommit: "", wantBranch: ""},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := callTool(t, session, taskReadTool, map[string]any{"id": tt.id})

				assertGitFields(t, got, tt.wantCommit, tt.wantBranch)
			})
		}
	})
}

func seedGitBars(t *testing.T, dir string) {
	t.Helper()

	seedTasks(t, dir,
		&task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing, Order: []string{testBarID, testSecondBarID}},
		&task.Task{ID: testBarID, Title: testBarTitle, Status: task.StatusDone, Commit: testBarSHA, Branch: "v2"},
		&task.Task{ID: testSecondBarID, Title: testSecondBarTitle, Status: task.StatusTodo},
	)
}

func assertGitFields(t *testing.T, got map[string]any, wantCommit, wantBranch string) {
	t.Helper()

	for key, want := range map[string]string{testCommitKey: wantCommit, testBranchKey: wantBranch} {
		gotVal, ok := got[key]
		if !ok {
			t.Errorf("%s key missing", key)

			continue
		}

		if gotVal != want {
			t.Errorf("%s = %v, want %q", key, gotVal, want)
		}
	}
}

func TestRunMCPServer(t *testing.T) {
	t.Run("resolves worktree root to main checkout", func(t *testing.T) {
		dir := t.TempDir()

		seedTasks(t, dir, &task.Task{
			ID: testTrackID, Title: testTitle, Status: task.StatusTodo, Body: testBody,
		})

		session := mcpSession(t, filepath.Join(dir, ".claude", "worktrees", "wt"))

		got := callTool(t, session, taskReadTool, map[string]any{"id": testTrackID})

		if got[testTitleKey] != testTitle {
			t.Errorf("title = %v, want %s", got[testTitleKey], testTitle)
		}
	})

	t.Run("resolves a subfolder root through the registry", func(t *testing.T) {
		mcpSandbox(t)

		dir := t.TempDir()

		path, err := project.CanonicalPath(dir)
		if err != nil {
			t.Fatal(err)
		}

		seedProject(t, orm.CreateProjectParams{Path: path, Code: testPrefix})

		sub := filepath.Join(dir, "sub")
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatal(err)
		}

		session := mcpSession(t, sub)

		created := callTool(t, session, taskCreateTool, map[string]any{testTitleKey: "Track", testBodyKey: testBody})
		if created["id"] != testNewTrackID {
			t.Fatalf("id = %v, want %s", created["id"], testNewTrackID)
		}

		read := callTool(t, session, taskReadTool, map[string]any{"id": testNewTrackID})
		if read[testTitleKey] != "Track" {
			t.Errorf("title = %v, want Track", read[testTitleKey])
		}

		stored := filepath.Join(projectStoreDir(t, dir), testTasksDir, testNewTrackID+".json")
		if _, err := os.Stat(stored); err != nil {
			t.Errorf("os.Stat(%q) returned error: %v", stored, err)
		}

		for _, bitDir := range []string{filepath.Join(sub, ".bit"), filepath.Join(dir, ".bit")} {
			if _, err := os.Stat(bitDir); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("%s: stat err = %v, want ErrNotExist", bitDir, err)
			}
		}
	})

	t.Run("unregistered root is a tool error", func(t *testing.T) {
		result := callToolResult(t, mcpSession(t, t.TempDir()), taskListTool, map[string]any{})

		assertToolErrorNames(t, result, "not a bit project; run `bp add`")
	})

	t.Run("empty root falls back to the working directory", func(t *testing.T) {
		dir := t.TempDir()
		registerProject(t, dir)
		t.Chdir(dir)

		result := callToolResult(t, mcpSession(t, ""), taskListTool, map[string]any{})
		if result.IsError {
			t.Errorf("IsError = true, want false (content %v)", result.Content)
		}
	})
}

func TestTaskListHandler(t *testing.T) {
	t.Run("returns every task as fields", func(t *testing.T) {
		dir := t.TempDir()

		seedTasks(t, dir,
			&task.Task{
				ID: testTrackID, Title: testTitle, Status: task.StatusTodo,
				Approved: true, Order: []string{testBarID},
			},
			&task.Task{
				ID: testBarID, Title: testBarTitle, Status: task.StatusDoing,
				Phase: 2, PhaseLabel: testPhaseLabel,
			},
		)

		tasks := callToolList(t, mcpSession(t, dir), taskListTool, map[string]any{})

		want := []map[string]any{
			{
				"id": testTrackID, testTitleKey: testTitle, testStatusKey: task.StatusTodo,
				testApprovedKey: true, testPhaseKey: float64(0), testPhaseLabelKey: "", testParentKey: "",
			},
			{
				"id": testBarID, testTitleKey: testBarTitle, "status": task.StatusDoing,
				testApprovedKey: false, "phase": float64(2), "phase_label": testPhaseLabel, testParentKey: testTrackID,
			},
		}

		if len(tasks) != len(want) {
			t.Fatalf("tasks = %d entries, want %d", len(tasks), len(want))
		}

		for i, w := range want {
			for key, wantVal := range w {
				if gotVal := tasks[i][key]; gotVal != wantVal {
					t.Errorf("tasks[%d][%s] = %v, want %v", i, key, gotVal, wantVal)
				}
			}

			if _, ok := tasks[i]["body"]; ok {
				t.Errorf("tasks[%d] carries a body key", i)
			}
		}
	})

	t.Run("parent returns only that tracks bars in order", func(t *testing.T) {
		dir := t.TempDir()

		seedTasks(t, dir,
			&task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusTodo, Order: []string{testSecondBarID, testBarID}},
			&task.Task{ID: testBarID, Title: testBarTitle, Status: task.StatusDone},
			&task.Task{ID: testSecondBarID, Title: testSecondBarTitle, Status: task.StatusTodo},
			&task.Task{ID: testOtherTrackID, Title: testOtherTitle, Status: task.StatusTodo},
			&task.Task{ID: testOtherBarID, Title: testOtherBarTitle, Status: task.StatusTodo},
		)

		tasks := callToolList(t, mcpSession(t, dir), taskListTool, map[string]any{testParentKey: testTrackID})

		want := []string{testSecondBarID, testBarID}
		if len(tasks) != len(want) {
			t.Fatalf("tasks = %d entries, want %d", len(tasks), len(want))
		}

		for i, wantID := range want {
			if gotID := tasks[i]["id"]; gotID != wantID {
				t.Errorf("tasks[%d][id] = %v, want %v", i, gotID, wantID)
			}
		}
	})

	t.Run("returns commit and branch", func(t *testing.T) {
		dir := t.TempDir()

		seedGitBars(t, dir)

		tasks := callToolList(t, mcpSession(t, dir), taskListTool, map[string]any{testParentKey: testTrackID})
		if len(tasks) != 2 {
			t.Fatalf("tasks = %d entries, want 2", len(tasks))
		}

		assertGitFields(t, tasks[0], testBarSHA, "v2")
		assertGitFields(t, tasks[1], "", "")
	})
}

func TestMCPToolDescriptions(t *testing.T) {
	t.Run("carry the domain", func(t *testing.T) {
		res, err := mcpSession(t, t.TempDir()).ListTools(t.Context(), nil)
		if err != nil {
			t.Fatal(err)
		}

		described := make(map[string]string, len(res.Tools))
		for _, tool := range res.Tools {
			described[tool.Name] = tool.Description
		}

		tests := []struct {
			name string
			tool string
			want []string
		}{
			{name: taskReadTool, tool: taskReadTool, want: []string{testTrackSentence, testBarIDExample}},
			{name: taskListTool, tool: taskListTool, want: []string{testTrackSentence, testBarIDExample}},
			{name: taskCreateTool, tool: taskCreateTool, want: []string{testTrackSentence, testBarIDExample}},
			{name: taskCompleteTool, tool: taskCompleteTool, want: []string{testTrackSentence}},
			{
				name: taskUpdateTool + " approval",
				tool: taskUpdateTool,
				want: []string{testRevokingFields, testTodoRevokes, testForwardKeepsApproval},
			},
			{
				name: taskUpdateTool + " rollup",
				tool: taskUpdateTool,
				want: []string{testNoCascade, testCallerRollsUp},
			},
			{
				name: taskUpdateTool + " git keeps approval",
				tool: taskUpdateTool,
				want: []string{testGitKeepsApproval},
			},
			{name: feedbackListTool, tool: feedbackListTool, want: []string{testOwnProjectOnly}},
			{name: feedbackReadTool, tool: feedbackReadTool, want: []string{testOwnProjectOnly}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, ok := described[tt.tool]
				if !ok {
					t.Fatalf("%s is not a registered tool", tt.tool)
				}

				for _, want := range tt.want {
					if !strings.Contains(got, want) {
						t.Errorf("%s description is missing %q", tt.tool, want)
					}
				}
			})
		}
	})
}
