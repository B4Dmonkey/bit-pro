package project

import (
	"slices"
	"testing"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/task"
)

func isolateData(t *testing.T) {
	t.Helper()

	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")
}

func registerCodes(t *testing.T, codes ...string) {
	t.Helper()

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	t.Cleanup(func() { sqlDB.Close() })

	q := orm.New(sqlDB)
	for _, code := range codes {
		if err := q.CreateProject(t.Context(), orm.CreateProjectParams{Path: t.TempDir(), Code: code}); err != nil {
			t.Fatalf("CreateProject(%q) returned error: %v", code, err)
		}
	}
}

func createTask(t *testing.T, s *task.Store, p task.CreateParams) {
	t.Helper()

	if _, err := s.Create(p); err != nil {
		t.Fatalf("Create(%+v) returned error: %v", p, err)
	}
}

func storeFor(t *testing.T, code string) *task.Store {
	t.Helper()

	s, err := StoreFor(Project{Code: code})
	if err != nil {
		t.Fatalf("StoreFor(%q) returned error: %v", code, err)
	}

	return s
}

func listIDs(t *testing.T, s *Stores) []string {
	t.Helper()

	tasks, err := s.List()
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}

	ids := make([]string, 0, len(tasks))
	for _, tk := range tasks {
		ids = append(ids, tk.ID)
	}

	return ids
}

func TestStoresList(t *testing.T) {
	t.Run("groups projects by code", func(t *testing.T) {
		isolateData(t)
		registerCodes(t, testCode, outerCode)

		bit := storeFor(t, testCode)
		createTask(t, bit, task.CreateParams{Title: "first"})
		createTask(t, bit, task.CreateParams{Title: "second"})

		acme := storeFor(t, outerCode)
		createTask(t, acme, task.CreateParams{Title: "track"})
		createTask(t, acme, task.CreateParams{Title: "bar", Parent: "ACME-1"})

		stores, err := OpenAll(t.Context())
		if err != nil {
			t.Fatalf("OpenAll() returned error: %v", err)
		}

		want := []string{"ACME-1", "ACME-1.1", "BIT-2", "BIT-1"}
		if got := listIDs(t, stores); !slices.Equal(got, want) {
			t.Errorf("List() IDs = %v, want %v", got, want)
		}
	})
	t.Run("empty registry lists nothing", func(t *testing.T) {
		isolateData(t)

		stores, err := OpenAll(t.Context())
		if err != nil {
			t.Fatalf("OpenAll() returned error: %v", err)
		}

		if got := listIDs(t, stores); len(got) != 0 {
			t.Errorf("List() IDs = %v, want none", got)
		}
	})
}
