package db

import (
	"testing"

	"github.com/B4Dmonkey/bit-pro/db/orm"
)

func TestQueue_EnqueueAndList(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")

	sqlDB, err := Open()
	if err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}
	defer sqlDB.Close()

	q := orm.New(sqlDB)

	if err := q.CreateProject(t.Context(), orm.CreateProjectParams{Path: "/tmp/proj", Code: "TST"}); err != nil {
		t.Fatalf("CreateProject() returned error: %v", err)
	}

	project, err := q.GetProjectByPath(t.Context(), "/tmp/proj")
	if err != nil {
		t.Fatalf("GetProjectByPath() returned error: %v", err)
	}

	before, err := q.ListQueueByProject(t.Context(), project.ID)
	if err != nil {
		t.Fatalf("ListQueueByProject() returned error: %v", err)
	}

	if len(before) != 0 {
		t.Fatalf("ListQueueByProject() returned %d rows before enqueue, want 0", len(before))
	}

	if err := q.EnqueueTask(t.Context(), orm.EnqueueTaskParams{
		ProjectID: project.ID,
		TargetID:  "BIT-33",
		TargetTyp: "track",
	}); err != nil {
		t.Fatalf("EnqueueTask() returned error: %v", err)
	}

	rows, err := q.ListQueueByProject(t.Context(), project.ID)
	if err != nil {
		t.Fatalf("ListQueueByProject() returned error: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("ListQueueByProject() returned %d rows, want 1", len(rows))
	}

	if rows[0].TargetID != "BIT-33" {
		t.Errorf("TargetID = %q, want %q", rows[0].TargetID, "BIT-33")
	}

	if rows[0].TargetTyp != "track" {
		t.Errorf("TargetTyp = %q, want %q", rows[0].TargetTyp, "track")
	}
}

func TestQueue_ListEmpty(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")

	sqlDB, err := Open()
	if err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}
	defer sqlDB.Close()

	q := orm.New(sqlDB)

	if err := q.CreateProject(t.Context(), orm.CreateProjectParams{Path: "/tmp/proj", Code: "TST"}); err != nil {
		t.Fatalf("CreateProject() returned error: %v", err)
	}

	project, err := q.GetProjectByPath(t.Context(), "/tmp/proj")
	if err != nil {
		t.Fatalf("GetProjectByPath() returned error: %v", err)
	}

	rows, err := q.ListQueueByProject(t.Context(), project.ID)
	if err != nil {
		t.Fatalf("ListQueueByProject() returned error: %v", err)
	}

	if len(rows) != 0 {
		t.Errorf("ListQueueByProject() returned %d rows, want 0", len(rows))
	}
}
