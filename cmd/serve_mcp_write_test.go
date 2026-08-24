package cmd

import (
	"path/filepath"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
)

const (
	testCreateBody = "## Why\n\nBecause `sed` owns the file format today.\n\n## Summary\n\nSix tools."
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
