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
)

func TestServeMCPCmd_TaskReadReturnsStructuredFields(t *testing.T) {
	dir := t.TempDir()

	store := task.New(filepath.Join(dir, ".bit"))
	if err := store.Save(&task.Task{
		ID: testTrackID, Title: testTitle, Status: task.StatusTodo, Body: "the body",
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

	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, clientT, nil)
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

	if got["body"] != "the body" {
		t.Errorf("body = %v, want the body", got["body"])
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

	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, clientT, nil)
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
