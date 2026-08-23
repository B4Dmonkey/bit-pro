package cmd

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	testTrackID = "FOO-1"
	testTitle   = "mcp test track"
	testBody    = "the body"
	testClient  = "test"

	testBarID      = "FOO-1.1"
	testBarTitle   = "a bar"
	testPhaseLabel = "Read surface"
)

func TestServeMCPCmd_TaskReadReturnsStructuredFields(t *testing.T) {
	dir := t.TempDir()

	store := task.New(filepath.Join(dir, ".bit"))
	if err := store.Save(&task.Task{
		ID: testTrackID, Title: testTitle, Status: task.StatusTodo, Body: testBody,
	}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())

	serverT, clientT := mcp.NewInMemoryTransports()

	errCh := make(chan error, 1)
	go func() { errCh <- runMCPServer(ctx, dir, serverT) }()

	t.Cleanup(func() {
		cancel()
		<-errCh
	})

	session, err := mcp.NewClient(&mcp.Implementation{Name: testClient, Version: "1"}, nil).Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatal(err)
	}

	defer session.Close()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      taskReadTool,
		Arguments: map[string]string{"id": testTrackID},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.IsError {
		t.Fatalf("task_read returned error: %v", result.Content)
	}

	b, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any

	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}

	if got["id"] != testTrackID {
		t.Errorf("id = %v, want %s", got["id"], testTrackID)
	}

	if got["title"] != testTitle {
		t.Errorf("title = %v, want %s", got["title"], testTitle)
	}

	if got["status"] != "todo" {
		t.Errorf("status = %v, want todo", got["status"])
	}

	if got["approved"] != false {
		t.Errorf("approved = %v, want false", got["approved"])
	}

	if got["body"] != testBody {
		t.Errorf("body = %v, want %s", got["body"], testBody)
	}

	if got["parent"] != "" {
		t.Errorf("parent = %v, want empty string", got["parent"])
	}
}

func TestServeMCPCmd_TaskReadReturnsParentForBar(t *testing.T) {
	dir := t.TempDir()

	store := task.New(filepath.Join(dir, ".bit"))
	if err := store.Save(&task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusTodo}); err != nil {
		t.Fatal(err)
	}

	if err := store.Save(&task.Task{ID: "FOO-1.1", Title: "a bar", Status: task.StatusTodo}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())

	serverT, clientT := mcp.NewInMemoryTransports()

	errCh := make(chan error, 1)
	go func() { errCh <- runMCPServer(ctx, dir, serverT) }()

	t.Cleanup(func() {
		cancel()
		<-errCh
	})

	session, err := mcp.NewClient(&mcp.Implementation{Name: testClient, Version: "1"}, nil).Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatal(err)
	}

	defer session.Close()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      taskReadTool,
		Arguments: map[string]string{"id": "FOO-1.1"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.IsError {
		t.Fatalf("task_read returned error: %v", result.Content)
	}

	b, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any

	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}

	if got["parent"] != testTrackID {
		t.Errorf("parent = %v, want %s", got["parent"], testTrackID)
	}
}

func TestServeMCPCmd_ResolvesWorktreeRootToMainCheckout(t *testing.T) {
	dir := t.TempDir()

	store := task.New(filepath.Join(dir, ".bit"))
	if err := store.Save(&task.Task{
		ID: testTrackID, Title: testTitle, Status: task.StatusTodo, Body: testBody,
	}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())

	serverT, clientT := mcp.NewInMemoryTransports()

	errCh := make(chan error, 1)
	go func() { errCh <- runMCPServer(ctx, filepath.Join(dir, ".claude", "worktrees", "wt"), serverT) }()

	t.Cleanup(func() {
		cancel()
		<-errCh
	})

	session, err := mcp.NewClient(&mcp.Implementation{Name: testClient, Version: "1"}, nil).Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatal(err)
	}

	defer session.Close()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      taskReadTool,
		Arguments: map[string]string{"id": testTrackID},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.IsError {
		t.Fatalf("task_read returned error: %v", result.Content)
	}

	b, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any

	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}

	if got["title"] != testTitle {
		t.Errorf("title = %v, want %s", got["title"], testTitle)
	}
}

func TestServeMCPCmd_TaskListReturnsEveryTaskAsFields(t *testing.T) {
	dir := t.TempDir()

	store := task.New(filepath.Join(dir, ".bit"))
	if err := store.Save(&task.Task{
		ID: testTrackID, Title: testTitle, Status: task.StatusTodo,
		Approved: true, Order: []string{testBarID},
	}); err != nil {
		t.Fatal(err)
	}

	if err := store.Save(&task.Task{
		ID: testBarID, Title: testBarTitle, Status: task.StatusDoing,
		Phase: 2, PhaseLabel: testPhaseLabel,
	}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())

	serverT, clientT := mcp.NewInMemoryTransports()

	errCh := make(chan error, 1)
	go func() { errCh <- runMCPServer(ctx, dir, serverT) }()

	t.Cleanup(func() {
		cancel()
		<-errCh
	})

	session, err := mcp.NewClient(&mcp.Implementation{Name: testClient, Version: "1"}, nil).Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatal(err)
	}

	defer session.Close()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      taskListTool,
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.IsError {
		t.Fatalf("task_list returned error: %v", result.Content)
	}

	b, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}

	var got struct {
		Tasks []map[string]any `json:"tasks"`
	}

	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}

	want := []map[string]any{
		{
			"id": testTrackID, "title": testTitle, "status": task.StatusTodo,
			"approved": true, "phase": float64(0), "phase_label": "", "parent": "",
		},
		{
			"id": testBarID, "title": testBarTitle, "status": task.StatusDoing,
			"approved": false, "phase": float64(2), "phase_label": testPhaseLabel, "parent": testTrackID,
		},
	}

	if len(got.Tasks) != len(want) {
		t.Fatalf("tasks = %d entries, want %d", len(got.Tasks), len(want))
	}

	for i, w := range want {
		for key, wantVal := range w {
			if gotVal := got.Tasks[i][key]; gotVal != wantVal {
				t.Errorf("tasks[%d][%s] = %v, want %v", i, key, gotVal, wantVal)
			}
		}

		if _, ok := got.Tasks[i]["body"]; ok {
			t.Errorf("tasks[%d] carries a body key", i)
		}
	}
}
