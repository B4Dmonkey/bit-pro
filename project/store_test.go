package project

import (
	"errors"
	"io/fs"
	"slices"
	"testing"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/task"
)

const (
	acmeTrack = "ACME-1"
	acmeBar   = "ACME-1.1"
	bitSecond = "BIT-2"
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
		createTask(t, acme, task.CreateParams{Title: "bar", Parent: acmeTrack})

		stores, err := OpenAll(t.Context())
		if err != nil {
			t.Fatalf("OpenAll() returned error: %v", err)
		}

		want := []string{acmeTrack, acmeBar, bitSecond, "BIT-1"}
		if got := listIDs(t, stores); !slices.Equal(got, want) {
			t.Errorf("List() IDs = %v, want %v", got, want)
		}
	})
	t.Run("hides removed projects", func(t *testing.T) {
		isolateData(t)
		registerCodes(t, testCode, outerCode)

		createTask(t, storeFor(t, testCode), task.CreateParams{Title: "kept"})
		createTask(t, storeFor(t, outerCode), task.CreateParams{Title: "removed"})

		sqlDB, err := db.Open()
		if err != nil {
			t.Fatalf("db.Open() returned error: %v", err)
		}

		t.Cleanup(func() { sqlDB.Close() })

		q := orm.New(sqlDB)

		projects, err := q.ListProjects(t.Context())
		if err != nil {
			t.Fatalf("ListProjects() returned error: %v", err)
		}

		i := slices.IndexFunc(projects, func(p orm.Project) bool { return p.Code == outerCode })
		if i < 0 {
			t.Fatalf("ListProjects() has no %s project", outerCode)
		}

		if err := q.SetProjectRemoved(t.Context(), orm.SetProjectRemovedParams{Removed: 1, ID: projects[i].ID}); err != nil {
			t.Fatalf("SetProjectRemoved() returned error: %v", err)
		}

		stores, err := OpenAll(t.Context())
		if err != nil {
			t.Fatalf("OpenAll() returned error: %v", err)
		}

		want := []string{"BIT-1"}
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

func TestStoresFor(t *testing.T) {
	isolateData(t)
	registerCodes(t, testCode, outerCode)

	bit := storeFor(t, testCode)
	createTask(t, bit, task.CreateParams{Title: "first"})
	createTask(t, bit, task.CreateParams{Title: "second"})

	acme := storeFor(t, outerCode)
	createTask(t, acme, task.CreateParams{Title: "track"})
	createTask(t, acme, task.CreateParams{Title: "bar", Parent: acmeTrack})

	stores, err := OpenAll(t.Context())
	if err != nil {
		t.Fatalf("OpenAll() returned error: %v", err)
	}

	tests := []struct {
		name        string
		id          string
		wantID      string
		wantProject string
		wantErr     error
	}{
		{name: "track in first project", id: bitSecond, wantID: bitSecond, wantProject: testCode},
		{name: "bar in second project", id: acmeBar, wantID: acmeBar, wantProject: outerCode},
		{name: "lowercase id", id: "acme-1", wantID: acmeTrack, wantProject: outerCode},
		{name: "unknown code", id: "FOO-1", wantErr: fs.ErrNotExist},
		{name: "no code prefix", id: "../../README", wantErr: fs.ErrNotExist},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := stores.For(tt.id)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("For(%q) error = %v, want %v", tt.id, err, tt.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("For(%q) returned error: %v", tt.id, err)
			}

			got, err := s.Load(tt.id)
			if err != nil {
				t.Fatalf("Load(%q) returned error: %v", tt.id, err)
			}

			if got.ID != tt.wantID || got.Project != tt.wantProject {
				t.Errorf("Load(%q) = %s in %s, want %s in %s", tt.id, got.ID, got.Project, tt.wantID, tt.wantProject)
			}
		})
	}
}
