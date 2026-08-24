package cmd

import (
	"path/filepath"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
)

const (
	testCreateBody = "## Why\n\nBecause `sed` owns the file format today.\n\n## Summary\n\nSix tools."
	testUpdateBody = "## Why\n\nnew reason\n\n## Decisions\n\n- settled"
)

func seedConfig(t *testing.T, dir string) {
	t.Helper()

	if err := task.New(filepath.Join(dir, ".bit")).SaveConfig(&task.Config{Prefix: testCode}); err != nil {
		t.Fatal(err)
	}
}

func TestServeMCPCmd_TaskCreateMintsATrack(t *testing.T) {
	dir := t.TempDir()
	seedConfig(t, dir)

	got := callTool(t, mcpSession(t, dir), taskCreateTool, map[string]any{
		testTitleKey: testTitle,
		"body":       testCreateBody,
	})

	if got["id"] != testTrackID {
		t.Fatalf("id = %v, want %s", got["id"], testTrackID)
	}

	created, err := task.New(filepath.Join(dir, ".bit")).Load(testTrackID)
	if err != nil {
		t.Fatal(err)
	}

	if created.Title != testTitle {
		t.Errorf("Title = %q, want %q", created.Title, testTitle)
	}

	if created.Status != task.StatusTodo {
		t.Errorf("Status = %q, want %q", created.Status, task.StatusTodo)
	}

	if created.Body != testCreateBody {
		t.Errorf("Body = %q, want %q", created.Body, testCreateBody)
	}
}

func TestServeMCPCmd_TaskUpdateRewritesBodyAndReportsRevocation(t *testing.T) {
	dir := t.TempDir()
	seedTasks(t, dir, &task.Task{
		ID: testTrackID, Title: testTitle, Status: task.StatusTodo, Approved: true, Body: "## Why\n\nold reason",
	})

	got := callTool(t, mcpSession(t, dir), taskUpdateTool, map[string]any{
		"id":   testTrackID,
		"body": testUpdateBody,
	})

	if got["id"] != testTrackID {
		t.Errorf("id = %v, want %s", got["id"], testTrackID)
	}

	if got["approved"] != false {
		t.Errorf("approved = %v, want false", got["approved"])
	}

	updated, err := task.New(filepath.Join(dir, ".bit")).Load(testTrackID)
	if err != nil {
		t.Fatal(err)
	}

	if updated.Body != testUpdateBody {
		t.Errorf("Body = %q, want %q", updated.Body, testUpdateBody)
	}

	if updated.Approved {
		t.Error("Approved = true, want false")
	}
}
