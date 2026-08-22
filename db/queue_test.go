package db

import (
	"testing"

	"github.com/B4Dmonkey/bit-pro/db/orm"
)

const (
	testProjPath  = "/tmp/proj"
	testProjPath2 = "/tmp/proj2"
	testProjCode  = "TST"
	testBarTyp    = "bar"
	testBarID     = "BIT-33.1"
)

func openTestDB(t *testing.T) *orm.Queries {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")

	sqlDB, err := Open()
	if err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}

	t.Cleanup(func() { sqlDB.Close() })

	return orm.New(sqlDB)
}

func createTestProjects(t *testing.T, q *orm.Queries, paths []string) map[string]int64 {
	t.Helper()

	ids := make(map[string]int64, len(paths))

	for _, path := range paths {
		if err := q.CreateProject(t.Context(), orm.CreateProjectParams{Path: path, Code: testProjCode}); err != nil {
			t.Fatalf("CreateProject(%q) returned error: %v", path, err)
		}

		project, err := q.GetProjectByPath(t.Context(), path)
		if err != nil {
			t.Fatalf("GetProjectByPath(%q) returned error: %v", path, err)
		}

		ids[path] = project.ID
	}

	return ids
}

func TestQueue_EnqueueAndList(t *testing.T) {
	q := openTestDB(t)
	ids := createTestProjects(t, q, []string{testProjPath})

	before, err := q.ListQueueByProject(t.Context(), ids[testProjPath])
	if err != nil {
		t.Fatalf("ListQueueByProject() returned error: %v", err)
	}

	if len(before) != 0 {
		t.Fatalf("ListQueueByProject() returned %d rows before enqueue, want 0", len(before))
	}

	if err := q.EnqueueTask(t.Context(), orm.EnqueueTaskParams{
		ProjectID: ids[testProjPath],
		TargetID:  testBarID,
		TargetTyp: testBarTyp,
	}); err != nil {
		t.Fatalf("EnqueueTask() returned error: %v", err)
	}

	rows, err := q.ListQueueByProject(t.Context(), ids[testProjPath])
	if err != nil {
		t.Fatalf("ListQueueByProject() returned error: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("ListQueueByProject() returned %d rows, want 1", len(rows))
	}

	if rows[0].TargetID != testBarID {
		t.Errorf("TargetID = %q, want %q", rows[0].TargetID, testBarID)
	}

	if rows[0].TargetTyp != testBarTyp {
		t.Errorf("TargetTyp = %q, want %q", rows[0].TargetTyp, testBarTyp)
	}
}

func TestQueue_ListEmpty(t *testing.T) {
	q := openTestDB(t)
	ids := createTestProjects(t, q, []string{testProjPath})

	rows, err := q.ListQueueByProject(t.Context(), ids[testProjPath])
	if err != nil {
		t.Fatalf("ListQueueByProject() returned error: %v", err)
	}

	if len(rows) != 0 {
		t.Errorf("ListQueueByProject() returned %d rows, want 0", len(rows))
	}
}

type queueEnqueue struct {
	project  string
	targetID string
}

func enqueueAll(t *testing.T, q *orm.Queries, ids map[string]int64, enqueues []queueEnqueue) {
	t.Helper()

	for _, e := range enqueues {
		if err := q.EnqueueTask(t.Context(), orm.EnqueueTaskParams{
			ProjectID: ids[e.project],
			TargetID:  e.targetID,
			TargetTyp: testBarTyp,
		}); err != nil {
			t.Fatalf("EnqueueTask(%q, %q) returned error: %v", e.project, e.targetID, err)
		}
	}
}

func assertQueue(t *testing.T, q *orm.Queries, ids map[string]int64, want map[string][]string) {
	t.Helper()

	for path, targetIDs := range want {
		rows, err := q.ListQueueByProject(t.Context(), ids[path])
		if err != nil {
			t.Fatalf("ListQueueByProject(%q) returned error: %v", path, err)
		}

		if len(rows) != len(targetIDs) {
			t.Fatalf("ListQueueByProject(%q) returned %d rows, want %d", path, len(rows), len(targetIDs))
		}

		for i, targetID := range targetIDs {
			if rows[i].TargetID != targetID {
				t.Errorf("%s row %d TargetID = %q, want %q", path, i, rows[i].TargetID, targetID)
			}

			if rows[i].TargetTyp != testBarTyp {
				t.Errorf("%s row %d TargetTyp = %q, want %q", path, i, rows[i].TargetTyp, testBarTyp)
			}
		}
	}
}

func TestQueue_EnqueueIsIdempotent(t *testing.T) {
	tests := []struct {
		name     string
		projects []string
		enqueues []queueEnqueue
		want     map[string][]string
	}{
		{
			name:     "same target twice",
			projects: []string{testProjPath},
			enqueues: []queueEnqueue{
				{project: testProjPath, targetID: testBarID},
				{project: testProjPath, targetID: testBarID},
			},
			want: map[string][]string{testProjPath: {testBarID}},
		},
		{
			name:     "distinct targets, one project",
			projects: []string{testProjPath},
			enqueues: []queueEnqueue{
				{project: testProjPath, targetID: testBarID},
				{project: testProjPath, targetID: "BIT-33.2"},
			},
			want: map[string][]string{testProjPath: {testBarID, "BIT-33.2"}},
		},
		{
			name:     "same target, two projects",
			projects: []string{testProjPath, testProjPath2},
			enqueues: []queueEnqueue{
				{project: testProjPath, targetID: testBarID},
				{project: testProjPath2, targetID: testBarID},
			},
			want: map[string][]string{
				testProjPath:  {testBarID},
				testProjPath2: {testBarID},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := openTestDB(t)
			ids := createTestProjects(t, q, tt.projects)

			enqueueAll(t, q, ids, tt.enqueues)
			assertQueue(t, q, ids, tt.want)
		})
	}
}
