package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
)

const (
	testTrackID = "FOO-1"
	testTitle   = "mcp test track"
	testBody    = "the body"
	testClient  = "test"

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

	testNoCascade     = "does not cascade"
	testCallerRollsUp = "sets the track's status in a separate call"
)

func TestServeMCPCmd_TaskReadReturnsStructuredFields(t *testing.T) {
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
}

func TestServeMCPCmd_TaskReadReturnsParentForBar(t *testing.T) {
	dir := t.TempDir()

	seedTasks(t, dir,
		&task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusTodo},
		&task.Task{ID: testBarID, Title: testBarTitle, Status: task.StatusTodo},
	)

	got := callTool(t, mcpSession(t, dir), taskReadTool, map[string]any{"id": testBarID})

	if got["parent"] != testTrackID {
		t.Errorf("parent = %v, want %s", got["parent"], testTrackID)
	}
}

func TestServeMCPCmd_ResolvesWorktreeRootToMainCheckout(t *testing.T) {
	dir := t.TempDir()

	seedTasks(t, dir, &task.Task{
		ID: testTrackID, Title: testTitle, Status: task.StatusTodo, Body: testBody,
	})

	session := mcpSession(t, filepath.Join(dir, ".claude", "worktrees", "wt"))

	got := callTool(t, session, taskReadTool, map[string]any{"id": testTrackID})

	if got[testTitleKey] != testTitle {
		t.Errorf("title = %v, want %s", got[testTitleKey], testTitle)
	}
}

func TestServeMCPCmd_TaskListReturnsEveryTaskAsFields(t *testing.T) {
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
}

func TestServeMCPCmd_TaskListParentReturnsOnlyThatTracksBarsInOrder(t *testing.T) {
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
}

func TestMCPToolDescriptions_CarryTheDomain(t *testing.T) {
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
}
