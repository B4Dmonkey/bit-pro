package task

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestStorePath(t *testing.T) {
	t.Parallel()

	t.Run("contains untrusted id", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			id   string
			want string
		}{
			{name: "plain id", id: tid1, want: ".bit/tasks/BIT-1.json"},
			{name: "traversal cannot escape the tasks dir", id: "../../README", want: ".bit/tasks/README.json"},
			{name: "deep traversal cannot escape", id: "../../../../etc/passwd", want: ".bit/tasks/ETC/PASSWD.json"},
			{name: "absolute path cannot escape", id: "/etc/passwd", want: ".bit/tasks/ETC/PASSWD.json"},
			{name: "illegal characters are stripped", id: "a:b*c", want: ".bit/tasks/ABC.json"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got := New(".bit").Path(tt.id)
				if got != tt.want {
					t.Errorf("Path(%q) = %q, want %q", tt.id, got, tt.want)
				}

				if !strings.HasPrefix(got, ".bit/tasks/") {
					t.Errorf("Path(%q) = %q, escaped the tasks directory", tt.id, got)
				}
			})
		}
	})
}

func TestStoreRelocate(t *testing.T) {
	t.Parallel()

	t.Run("moves file out of list", func(t *testing.T) {
		t.Parallel()

		s := New(t.TempDir())
		if err := s.Save(&Task{ID: tid1, Status: StatusDone}); err != nil {
			t.Fatalf("seeding BIT-1: %v", err)
		}

		if err := s.Relocate(tid1, false); err != nil {
			t.Fatalf("Relocate() returned error: %v", err)
		}

		tasks, err := s.List()
		if err != nil {
			t.Fatalf("List() returned error: %v", err)
		}

		if slices.ContainsFunc(tasks, func(t *Task) bool { return t.ID == tid1 }) {
			t.Errorf("List() still contains BIT-1 after relocate")
		}

		for _, name := range []string{"BIT-1.json", "BIT-1.md"} {
			if _, err := os.Stat(filepath.Join(s.archiveTasksDir(), name)); err != nil {
				t.Errorf("archived %s: os.Stat error = %v, want the file to exist", name, err)
			}

			if _, err := os.Stat(filepath.Join(s.tasksDir(), name)); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("tasks %s: os.Stat error = %v, want fs.ErrNotExist", name, err)
			}
		}
	})

	t.Run("cascades to bars", func(t *testing.T) {
		t.Parallel()

		s := New(t.TempDir())
		for _, id := range []string{tid1, tid1_1, tid1_2} {
			if err := s.Save(&Task{ID: id, Status: StatusDone}); err != nil {
				t.Fatalf("seeding %s: %v", id, err)
			}
		}

		if err := s.Relocate(tid1, false); err != nil {
			t.Fatalf("Relocate() returned error: %v", err)
		}

		tasks, err := s.List()
		if err != nil {
			t.Fatalf("List() returned error: %v", err)
		}

		if len(tasks) != 0 {
			t.Errorf("List() = %v, want no tasks after cascade", tasks)
		}

		for _, id := range []string{tid1, tid1_1, tid1_2} {
			if _, err := os.Stat(s.archivePath(id)); err != nil {
				t.Errorf("archived %s: os.Stat error = %v, want the file to exist", id, err)
			}

			if _, err := os.Stat(s.Path(id)); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("tasks %s: os.Stat error = %v, want fs.ErrNotExist", id, err)
			}
		}
	})

	t.Run("refuses with unfinished bars", func(t *testing.T) {
		t.Parallel()

		s := New(t.TempDir())
		for _, seed := range []struct{ id, status string }{
			{tid1, StatusDone},
			{tid1_1, StatusDone},
			{tid1_2, StatusTodo},
		} {
			if err := s.Save(&Task{ID: seed.id, Status: seed.status}); err != nil {
				t.Fatalf("seeding %s: %v", seed.id, err)
			}
		}

		err := s.Relocate(tid1, false)

		var unfinished *UnfinishedBarsError
		if !errors.As(err, &unfinished) {
			t.Fatalf("Relocate() error = %v, want *UnfinishedBarsError", err)
		}

		if !slices.Contains(unfinished.Bars, tid1_2) {
			t.Errorf("UnfinishedBarsError.Bars = %v, want it to contain BIT-1.2", unfinished.Bars)
		}

		for _, id := range []string{tid1, tid1_1, tid1_2} {
			if _, err := os.Stat(s.Path(id)); err != nil {
				t.Errorf("tasks %s: os.Stat error = %v, want the file to remain", id, err)
			}

			if _, err := os.Stat(s.archivePath(id)); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("archive %s: os.Stat error = %v, want fs.ErrNotExist", id, err)
			}
		}
	})

	t.Run("force overrides guard", func(t *testing.T) {
		t.Parallel()

		s := New(t.TempDir())
		for _, seed := range []struct{ id, status string }{
			{tid1, StatusDone},
			{tid1_1, StatusDone},
			{tid1_2, StatusTodo},
		} {
			if err := s.Save(&Task{ID: seed.id, Status: seed.status}); err != nil {
				t.Fatalf("seeding %s: %v", seed.id, err)
			}
		}

		if err := s.Relocate(tid1, true); err != nil {
			t.Fatalf("Relocate() returned error: %v", err)
		}

		for _, id := range []string{tid1, tid1_1, tid1_2} {
			if _, err := os.Stat(s.archivePath(id)); err != nil {
				t.Errorf("archived %s: os.Stat error = %v, want the file to exist", id, err)
			}

			if _, err := os.Stat(s.Path(id)); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("tasks %s: os.Stat error = %v, want fs.ErrNotExist", id, err)
			}
		}
	})

	t.Run("contains untrusted id", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			id   string
			want string
		}{
			{name: "plain id", id: tid1, want: ".bit/archive/tasks/BIT-1.json"},
			{name: "traversal cannot escape the archive dir", id: "../../README", want: ".bit/archive/tasks/README.json"},
			{name: "absolute path cannot escape", id: "/etc/passwd", want: ".bit/archive/tasks/ETC/PASSWD.json"},
			{name: "illegal characters are stripped", id: "a:b*c", want: ".bit/archive/tasks/ABC.json"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got := New(".bit").archivePath(tt.id)
				if got != tt.want {
					t.Errorf("archivePath(%q) = %q, want %q", tt.id, got, tt.want)
				}

				if !strings.HasPrefix(got, ".bit/archive/tasks/") {
					t.Errorf("archivePath(%q) = %q, escaped the archive directory", tt.id, got)
				}
			})
		}
	})

	t.Run("drops bar from parent order", func(t *testing.T) {
		t.Parallel()

		s := New(t.TempDir())
		if err := s.Save(&Task{ID: tid1, Status: StatusDone, Order: []string{tid1_1, tid1_2, tid1_3}}); err != nil {
			t.Fatalf("seeding BIT-1: %v", err)
		}

		for _, id := range []string{tid1_1, tid1_2, tid1_3} {
			if err := s.Save(&Task{ID: id, Status: StatusDone}); err != nil {
				t.Fatalf("seeding %s: %v", id, err)
			}
		}

		if err := s.Relocate(tid1_2, false); err != nil {
			t.Fatalf("Relocate() returned error: %v", err)
		}

		got, err := s.Load(tid1)
		if err != nil {
			t.Fatalf("loading BIT-1: %v", err)
		}

		if want := []string{tid1_1, tid1_3}; !slices.Equal(got.Order, want) {
			t.Errorf("Order = %v, want %v", got.Order, want)
		}
	})

	t.Run("leaves legacy order unmaterialized", func(t *testing.T) {
		t.Parallel()

		s := New(t.TempDir())
		if err := s.Save(&Task{ID: tid1, Status: StatusDone, Order: nil}); err != nil {
			t.Fatalf("seeding BIT-1: %v", err)
		}

		for _, id := range []string{tid1_1, tid1_2} {
			if err := s.Save(&Task{ID: id, Status: StatusDone}); err != nil {
				t.Fatalf("seeding %s: %v", id, err)
			}
		}

		if err := s.Relocate(tid1_1, false); err != nil {
			t.Fatalf("Relocate() returned error: %v", err)
		}

		got, err := s.Load(tid1)
		if err != nil {
			t.Fatalf("loading BIT-1: %v", err)
		}

		if len(got.Order) != 0 {
			t.Errorf("Order = %v, want empty", got.Order)
		}
	})
}

func TestStoreComplete(t *testing.T) {
	t.Parallel()

	t.Run("moves both files", func(t *testing.T) {
		t.Parallel()

		s := New(t.TempDir())
		if err := s.Save(&Task{ID: tid1, Status: StatusDone}); err != nil {
			t.Fatalf("seeding BIT-1: %v", err)
		}

		if err := s.Complete(tid1); err != nil {
			t.Fatalf("Complete() returned error: %v", err)
		}

		for _, name := range []string{"BIT-1.json", "BIT-1.md"} {
			if _, err := os.Stat(filepath.Join(s.completedDir(), name)); err != nil {
				t.Errorf("completed %s: os.Stat error = %v, want the file to exist", name, err)
			}

			if _, err := os.Stat(filepath.Join(s.tasksDir(), name)); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("tasks %s: os.Stat error = %v, want fs.ErrNotExist", name, err)
			}
		}
	})
}

func TestStoreNextID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		existing []string
		want     string
	}{
		{name: "no tasks yet", existing: nil, want: tid1},
		{name: "continues past the highest, not the count", existing: []string{tid1, tid3}, want: tid4},
		{name: "ignores other prefixes", existing: []string{tid1, "OTHER-9"}, want: tid2},
		{name: "ignores non-numeric suffixes", existing: []string{tid1, "BIT-abc"}, want: tid2},
		{name: "handles multi-digit ids", existing: []string{"BIT-9", "BIT-10"}, want: "BIT-11"},
		{name: "ignores dotted children", existing: []string{tid1, tid1_1, "BIT-1.13"}, want: tid2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := New(t.TempDir())
			for _, id := range tt.existing {
				if err := s.Save(&Task{ID: id, Title: tseed, Status: StatusTodo}); err != nil {
					t.Fatalf("seeding %s: %v", id, err)
				}
			}

			got, err := s.NextID(tprefix)
			if err != nil {
				t.Fatalf("NextID() returned error: %v", err)
			}

			if got != tt.want {
				t.Errorf("NextID() = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("reserves archived ids", func(t *testing.T) {
		t.Parallel()

		s := New(t.TempDir())
		if err := s.Save(&Task{ID: tid1, Title: tseed, Status: StatusTodo}); err != nil {
			t.Fatalf("seeding BIT-1: %v", err)
		}

		if err := s.Save(&Task{ID: tid2, Title: tseed, Status: StatusDone}); err != nil {
			t.Fatalf("seeding BIT-2: %v", err)
		}

		if err := s.Relocate(tid2, false); err != nil {
			t.Fatalf("Relocate(BIT-2): %v", err)
		}

		got, err := s.NextID(tprefix)
		if err != nil {
			t.Fatalf("NextID() returned error: %v", err)
		}

		if got != tid3 {
			t.Errorf("NextID() = %q, want %q", got, tid3)
		}
	})

	t.Run("reserves completed ids", func(t *testing.T) {
		t.Parallel()

		s := New(t.TempDir())
		if err := s.Save(&Task{ID: tid1, Title: tseed, Status: StatusTodo}); err != nil {
			t.Fatalf("seeding BIT-1: %v", err)
		}

		if err := s.Save(&Task{ID: tid2, Title: tseed, Status: StatusDone}); err != nil {
			t.Fatalf("seeding BIT-2: %v", err)
		}

		if err := s.Complete(tid2); err != nil {
			t.Fatalf("Complete(BIT-2): %v", err)
		}

		got, err := s.NextID(tprefix)
		if err != nil {
			t.Fatalf("NextID() returned error: %v", err)
		}

		if got != tid3 {
			t.Errorf("NextID() = %q, want %q", got, tid3)
		}
	})
}

func TestStoreNextChildID(t *testing.T) {
	t.Parallel()

	t.Run("errors when parent missing", func(t *testing.T) {
		t.Parallel()

		_, err := New(t.TempDir()).NextChildID("BIT-99")

		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("NextChildID() error = %v, want an error wrapping fs.ErrNotExist", err)
		}

		if !strings.Contains(err.Error(), "BIT-99") {
			t.Errorf("NextChildID() error = %q, want it to name the parent ID", err)
		}
	})

	t.Run("mints when parent exists", func(t *testing.T) {
		t.Parallel()

		s := New(t.TempDir())
		if err := s.Save(&Task{ID: tid1, Title: tseed, Status: StatusTodo}); err != nil {
			t.Fatalf("seeding BIT-1: %v", err)
		}

		got, err := s.NextChildID(tid1)
		if err != nil {
			t.Fatalf("NextChildID() returned error: %v", err)
		}

		if got != tid1_1 {
			t.Errorf("NextChildID() = %q, want %q", got, tid1_1)
		}
	})

	t.Run("reserves archived children", func(t *testing.T) {
		t.Parallel()

		s := New(t.TempDir())
		if err := s.Save(&Task{ID: tid1, Title: tseed, Status: StatusTodo}); err != nil {
			t.Fatalf("seeding BIT-1: %v", err)
		}

		if err := s.Save(&Task{ID: tid1_1, Title: tseed, Status: StatusDone}); err != nil {
			t.Fatalf("seeding BIT-1.1: %v", err)
		}

		if err := s.Relocate(tid1_1, false); err != nil {
			t.Fatalf("Relocate(BIT-1.1): %v", err)
		}

		got, err := s.NextChildID(tid1)
		if err != nil {
			t.Fatalf("NextChildID() returned error: %v", err)
		}

		if got != tid1_2 {
			t.Errorf("NextChildID() = %q, want %q", got, tid1_2)
		}
	})

	t.Run("reserves completed children", func(t *testing.T) {
		t.Parallel()

		s := New(t.TempDir())
		if err := s.Save(&Task{ID: tid1, Title: tseed, Status: StatusTodo}); err != nil {
			t.Fatalf("seeding BIT-1: %v", err)
		}

		if err := s.Save(&Task{ID: tid1_1, Title: tseed, Status: StatusDone}); err != nil {
			t.Fatalf("seeding BIT-1.1: %v", err)
		}

		if err := s.Complete(tid1_1); err != nil {
			t.Fatalf("Complete(BIT-1.1): %v", err)
		}

		got, err := s.NextChildID(tid1)
		if err != nil {
			t.Fatalf("NextChildID() returned error: %v", err)
		}

		if got != tid1_2 {
			t.Errorf("NextChildID() = %q, want %q", got, tid1_2)
		}
	})
}

func TestStoreSave(t *testing.T) {
	t.Parallel()

	t.Run("writes a json record beside the body", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		s := NewProject(root, tprefix)
		body := "## Why\n\n- [ ] a checkbox\n"

		if err := s.Save(&Task{ID: "BIT-7", Title: "Ship it", Status: StatusTodo, Body: body}); err != nil {
			t.Fatalf("Save() returned error: %v", err)
		}

		gotBody, err := os.ReadFile(filepath.Join(root, "tasks", "BIT-7.md"))
		if err != nil {
			t.Fatalf("reading body: %v", err)
		}

		if string(gotBody) != body {
			t.Errorf("body = %q, want %q", gotBody, body)
		}

		raw, err := os.ReadFile(filepath.Join(root, "tasks", "BIT-7.json"))
		if err != nil {
			t.Fatalf("reading record: %v", err)
		}

		var got map[string]any
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("unmarshaling record: %v", err)
		}

		want := map[string]any{
			"id":          "BIT-7",
			"title":       "Ship it",
			"status":      StatusTodo,
			"approved":    false,
			"phase":       float64(0),
			"phase_label": "",
			"order":       []any{},
			"content":     "BIT-7.md",
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("record = %v, want %v", got, want)
		}

		if !strings.Contains(string(raw), `"order": []`) {
			t.Errorf("record = %s, want order written as []", raw)
		}

		if !strings.HasSuffix(string(raw), "}\n") {
			t.Errorf("record = %q, want a trailing newline", raw)
		}
	})
}

func TestStoreLoad(t *testing.T) {
	t.Parallel()

	t.Run("round trips a saved task", func(t *testing.T) {
		t.Parallel()

		s := New(t.TempDir())
		want := Task{
			ID:         tid1,
			Title:      "Title",
			Status:     StatusDoing,
			Approved:   true,
			Phase:      2,
			PhaseLabel: "records",
			Order:      []string{tid1_2, tid1_1},
			Body:       "Body.\n\nMore body.\n",
		}

		if err := s.Save(&want); err != nil {
			t.Fatalf("Save() returned error: %v", err)
		}

		got, err := s.Load(tid1)
		if err != nil {
			t.Fatalf("Load() returned error: %v", err)
		}

		if !reflect.DeepEqual(*got, want) {
			t.Errorf("Load() = %+v, want %+v", *got, want)
		}
	})

	t.Run("errors on unknown id", func(t *testing.T) {
		t.Parallel()

		_, err := New(t.TempDir()).Load("BIT-99")

		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("Load() error = %v, want an error wrapping fs.ErrNotExist", err)
		}

		if !strings.Contains(err.Error(), "BIT-99") {
			t.Errorf("Load() error = %q, want it to name the task ID", err)
		}
	})
}

func TestStoreList(t *testing.T) {
	t.Parallel()

	t.Run("empty when no tasks dir", func(t *testing.T) {
		t.Parallel()

		tasks, err := New(t.TempDir()).List()
		if err != nil {
			t.Fatalf("List() returned error: %v", err)
		}

		if len(tasks) != 0 {
			t.Errorf("List() = %v, want no tasks", tasks)
		}
	})

	t.Run("ignores a stray body", func(t *testing.T) {
		t.Parallel()

		s := New(t.TempDir())
		if err := s.Save(&Task{ID: tid1, Title: tseed, Status: StatusTodo}); err != nil {
			t.Fatalf("seeding BIT-1: %v", err)
		}

		if err := os.WriteFile(filepath.Join(s.tasksDir(), "BIT-9.md"), []byte("orphan\n"), fileMode); err != nil {
			t.Fatalf("writing stray body: %v", err)
		}

		tasks, err := s.List()
		if err != nil {
			t.Fatalf("List() returned error: %v", err)
		}

		if len(tasks) != 1 {
			t.Errorf("List() = %v, want 1 task", tasks)
		}

		got, err := s.NextID(tprefix)
		if err != nil {
			t.Fatalf("NextID() returned error: %v", err)
		}

		if got != tid2 {
			t.Errorf("NextID() = %q, want %q", got, tid2)
		}
	})

	t.Run("orders bars by explicit order", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			order []string
			want  []string
		}{
			{
				name:  "explicit order overrides id sequence",
				order: []string{tid1_2, tid1_1},
				want:  []string{tid1, tid1_2, tid1_1},
			},
			{
				name:  "no order falls back to id sequence",
				order: nil,
				want:  []string{tid1, tid1_1, tid1_2},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				s := New(t.TempDir())
				if err := s.Save(&Task{ID: tid1, Title: ttrack, Status: StatusTodo, Order: tt.order}); err != nil {
					t.Fatalf("seeding BIT-1: %v", err)
				}

				for _, id := range []string{tid1_1, tid1_2} {
					if err := s.Save(&Task{ID: id, Title: tbar, Status: StatusTodo}); err != nil {
						t.Fatalf("seeding %s: %v", id, err)
					}
				}

				tasks, err := s.List()
				if err != nil {
					t.Fatalf("List() returned error: %v", err)
				}

				got := make([]string, len(tasks))
				for i, task := range tasks {
					got[i] = task.ID
				}

				if !slices.Equal(got, tt.want) {
					t.Errorf("List() order = %v, want %v", got, tt.want)
				}
			})
		}
	})
}

func TestStoreMove(t *testing.T) {
	t.Parallel()

	t.Run("resequences", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name          string
			order         []string
			id            string
			before, after string
			want          []string
		}{
			{
				name:   "materializes then moves to front",
				order:  nil,
				id:     tid1_3,
				before: tid1_1,
				want:   []string{tid1_3, tid1_1, tid1_2},
			},
			{
				name:  "splices an existing order to the back",
				order: []string{tid1_1, tid1_2, tid1_3},
				id:    tid1_1,
				after: tid1_3,
				want:  []string{tid1_2, tid1_3, tid1_1},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				s := New(t.TempDir())
				if err := s.Save(&Task{ID: tid1, Title: ttrack, Status: StatusTodo, Order: tt.order}); err != nil {
					t.Fatalf("seeding BIT-1: %v", err)
				}

				for _, id := range []string{tid1_1, tid1_2, tid1_3} {
					if err := s.Save(&Task{ID: id, Title: tbar, Status: StatusTodo}); err != nil {
						t.Fatalf("seeding %s: %v", id, err)
					}
				}

				if err := s.Move(tt.id, tt.before, tt.after); err != nil {
					t.Fatalf("Move() returned error: %v", err)
				}

				got, err := s.Load(tid1)
				if err != nil {
					t.Fatalf("loading BIT-1: %v", err)
				}

				if !slices.Equal(got.Order, tt.want) {
					t.Errorf("Order = %v, want %v", got.Order, tt.want)
				}
			})
		}
	})

	t.Run("rejects", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name         string
			id           string
			anchor       string
			wantNotExist bool
		}{
			{name: "anchor under a different track", id: tid1_1, anchor: tid2_1},
			{name: "unknown bar", id: tid1_9, anchor: tid1_1, wantNotExist: true},
			{name: "unknown anchor", id: tid1_1, anchor: tid1_9, wantNotExist: true},
			{name: "moving a bar relative to itself", id: tid1_1, anchor: tid1_1},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				s := New(t.TempDir())
				for _, id := range []string{tid1, tid2} {
					if err := s.Save(&Task{ID: id, Title: ttrack, Status: StatusTodo}); err != nil {
						t.Fatalf("seeding %s: %v", id, err)
					}
				}

				for _, id := range []string{tid1_1, tid1_2, tid2_1} {
					if err := s.Save(&Task{ID: id, Title: tbar, Status: StatusTodo}); err != nil {
						t.Fatalf("seeding %s: %v", id, err)
					}
				}

				err := s.Move(tt.id, "", tt.anchor)
				if err == nil {
					t.Fatalf("Move(%q, %q) returned nil error, want non-nil", tt.id, tt.anchor)
				}

				if tt.wantNotExist && !errors.Is(err, fs.ErrNotExist) {
					t.Errorf("Move(%q, %q) error = %v, want it to wrap fs.ErrNotExist", tt.id, tt.anchor, err)
				}
			})
		}
	})

	t.Run("rejects anchor pair", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name          string
			before, after string
		}{
			{name: "both anchors", before: tid1_1, after: tid1_1},
			{name: "neither anchor"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				seed := []string{tid1_1, tid1_2}

				s := New(t.TempDir())
				if err := s.Save(&Task{ID: tid1, Title: ttrack, Status: StatusTodo, Order: seed}); err != nil {
					t.Fatalf("seeding %s: %v", tid1, err)
				}

				for _, id := range seed {
					if err := s.Save(&Task{ID: id, Title: tbar, Status: StatusTodo}); err != nil {
						t.Fatalf("seeding %s: %v", id, err)
					}
				}

				if err := s.Move(tid1_2, tt.before, tt.after); err == nil {
					t.Fatalf("Move(%q, %q, %q) returned nil error, want non-nil", tid1_2, tt.before, tt.after)
				}

				got, err := s.Load(tid1)
				if err != nil {
					t.Fatalf("loading %s: %v", tid1, err)
				}

				if !slices.Equal(got.Order, seed) {
					t.Errorf("Order = %v, want %v unchanged", got.Order, seed)
				}
			})
		}
	})
}

func TestCompareIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b string
		want int
	}{
		{name: "newer sorts before older", a: tid2, b: tid1, want: -1},
		{name: "older sorts after newer", a: tid1, b: tid2, want: 1},
		{name: "equal ids", a: tid1, b: tid1, want: 0},
		{name: "two-digit id sorts before one-digit", a: "BIT-10", b: "BIT-9", want: -1},
		{name: "unparseable suffix sorts last", a: "BIT-abc", b: tid1, want: 1},
		{name: "track heads its own bars", a: tid2, b: tid2_1, want: -1},
		{name: "bars ascend, not lexically", a: tid2_1, b: "BIT-2.13", want: -1},
		{name: "track dominates bar", a: tid2_1, b: tid1_9, want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := compareIDs(tt.a, tt.b)
			switch {
			case tt.want < 0 && got >= 0:
				t.Errorf("compareIDs(%q, %q) = %d, want negative", tt.a, tt.b, got)
			case tt.want > 0 && got <= 0:
				t.Errorf("compareIDs(%q, %q) = %d, want positive", tt.a, tt.b, got)
			case tt.want == 0 && got != 0:
				t.Errorf("compareIDs(%q, %q) = %d, want 0", tt.a, tt.b, got)
			}
		})
	}
}

func TestStoreCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		seed      []*Task
		params    CreateParams
		wantID    string
		wantOrder []string
	}{
		{
			name:   "mints the first track in an empty store",
			params: CreateParams{Title: ttrack, Body: "why\n\nbecause"},
			wantID: tid1,
		},
		{
			name:   "mints past an existing track",
			seed:   []*Task{{ID: tid1, Title: tseed, Status: StatusTodo}},
			params: CreateParams{Title: ttrack, Body: "why\n\nbecause"},
			wantID: tid2,
		},
		{
			name:   "mints a dotted child under a parent",
			seed:   []*Task{{ID: tid1, Title: tseed, Status: StatusTodo}},
			params: CreateParams{Title: tbar, Parent: tid1, Phase: 2, PhaseLabel: "Plan writes"},
			wantID: tid1_1,
		},
		{
			name: "splices a child into an explicit order",
			seed: []*Task{
				{ID: tid1, Title: tseed, Status: StatusTodo, Order: []string{tid1_1, tid1_2}},
				{ID: tid1_1, Title: tbar, Status: StatusTodo},
				{ID: tid1_2, Title: tbar, Status: StatusTodo},
			},
			params:    CreateParams{Title: tbar, Parent: tid1, After: tid1_1},
			wantID:    tid1_3,
			wantOrder: []string{tid1_1, tid1_3, tid1_2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := NewProject(t.TempDir(), tprefix)

			for _, seed := range tt.seed {
				if err := s.Save(seed); err != nil {
					t.Fatalf("seeding %s: %v", seed.ID, err)
				}
			}

			got, err := s.Create(tt.params)
			if err != nil {
				t.Fatalf("Create() returned error: %v", err)
			}

			loaded, err := s.Load(tt.wantID)
			if err != nil {
				t.Fatalf("loading %s: %v", tt.wantID, err)
			}

			assertCreated(t, "Create()", got, tt.params, tt.wantID)
			assertCreated(t, "Load()", loaded, tt.params, tt.wantID)

			if tt.wantOrder == nil {
				return
			}

			track, err := s.Load(tt.params.Parent)
			if err != nil {
				t.Fatalf("loading %s: %v", tt.params.Parent, err)
			}

			if !slices.Equal(track.Order, tt.wantOrder) {
				t.Errorf("Order = %v, want %v", track.Order, tt.wantOrder)
			}
		})
	}

	t.Run("mints from the project code", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			code    string
			creates int
			want    string
		}{
			{name: "first track", code: "EX", creates: 1, want: "EX-1"},
			{name: "second track in the same store", code: "EX", creates: 2, want: "EX-2"},
			{name: "another code", code: "ZZ", creates: 1, want: "ZZ-1"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				root := t.TempDir()
				s := NewProject(root, tt.code)

				var got *Task

				for range tt.creates {
					var err error

					got, err = s.Create(CreateParams{Title: ttrack})
					if err != nil {
						t.Fatalf("Create() returned error: %v", err)
					}
				}

				if got.ID != tt.want {
					t.Errorf("Create() ID = %q, want %q", got.ID, tt.want)
				}
			})
		}
	})

	t.Run("rejects unknown parent", func(t *testing.T) {
		t.Parallel()

		s := New(t.TempDir())

		_, err := s.Create(CreateParams{Title: tbar, Parent: tid1_9})

		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("Create() error = %v, want an error wrapping fs.ErrNotExist", err)
		}
	})
}

func assertCreated(t *testing.T, what string, got *Task, want CreateParams, wantID string) {
	t.Helper()

	if got.ID != wantID {
		t.Errorf("%s ID = %q, want %q", what, got.ID, wantID)
	}

	if got.Status != StatusTodo {
		t.Errorf("%s Status = %q, want %q", what, got.Status, StatusTodo)
	}

	if got.Title != want.Title {
		t.Errorf("%s Title = %q, want %q", what, got.Title, want.Title)
	}

	if got.Body != want.Body {
		t.Errorf("%s Body = %q, want %q", what, got.Body, want.Body)
	}

	if got.Phase != want.Phase {
		t.Errorf("%s Phase = %d, want %d", what, got.Phase, want.Phase)
	}

	if got.PhaseLabel != want.PhaseLabel {
		t.Errorf("%s PhaseLabel = %q, want %q", what, got.PhaseLabel, want.PhaseLabel)
	}
}

func TestStoreUpdate(t *testing.T) {
	t.Parallel()

	t.Run("applies only set fields", func(t *testing.T) {
		t.Parallel()

		const (
			oldTitle = "Old title"
			oldLabel = "Old label"
			oldBody  = "Old body."
		)

		seed := Task{
			ID:         tid1,
			Title:      oldTitle,
			Status:     StatusTodo,
			Phase:      3,
			PhaseLabel: oldLabel,
			Body:       oldBody,
		}

		tests := []struct {
			name  string
			patch Patch
			want  Task
		}{
			{
				name:  "an empty patch changes nothing",
				patch: Patch{},
				want:  seed,
			},
			{
				name:  "a set title is written",
				patch: Patch{Title: ptr("New title")},
				want:  Task{ID: tid1, Title: "New title", Status: StatusTodo, Phase: 3, PhaseLabel: oldLabel, Body: oldBody},
			},
			{
				name:  "body and status are written together",
				patch: Patch{Body: ptr("New body."), Status: ptr(StatusDoing)},
				want:  Task{ID: tid1, Title: oldTitle, Status: StatusDoing, Phase: 3, PhaseLabel: oldLabel, Body: "New body."},
			},
			{
				name:  "an explicitly empty title is written",
				patch: Patch{Title: ptr("")},
				want:  Task{ID: tid1, Title: "", Status: StatusTodo, Phase: 3, PhaseLabel: oldLabel, Body: oldBody},
			},
			{
				name:  "a zero phase and empty label are written",
				patch: Patch{Phase: ptr(0), PhaseLabel: ptr("")},
				want:  Task{ID: tid1, Title: oldTitle, Status: StatusTodo, Phase: 0, PhaseLabel: "", Body: oldBody},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				s := New(t.TempDir())

				s0 := seed
				if err := s.Save(&s0); err != nil {
					t.Fatalf("seeding %s: %v", seed.ID, err)
				}

				got, err := s.Update(tid1, tt.patch)
				if err != nil {
					t.Fatalf("Update() returned error: %v", err)
				}

				loaded, err := s.Load(tid1)
				if err != nil {
					t.Fatalf("loading %s: %v", tid1, err)
				}

				if !reflect.DeepEqual(*got, tt.want) {
					t.Errorf("Update() = %+v, want %+v", *got, tt.want)
				}

				if !reflect.DeepEqual(*loaded, tt.want) {
					t.Errorf("Load() = %+v, want %+v", *loaded, tt.want)
				}
			})
		}
	})

	t.Run("approval revocation", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name         string
			approved     bool
			patch        Patch
			wantApproved bool
		}{
			{name: "a title change revokes", approved: true, patch: Patch{Title: ptr("x")}},
			{name: "a body change revokes", approved: true, patch: Patch{Body: ptr("x")}},
			{name: "a phase change revokes", approved: true, patch: Patch{Phase: ptr(2)}},
			{name: "a phase-label change revokes", approved: true, patch: Patch{PhaseLabel: ptr("x")}},
			{name: "sending a task back to todo revokes", approved: true, patch: Patch{Status: ptr(StatusTodo)}},
			{
				name:         "a forward status move keeps approval",
				approved:     true,
				patch:        Patch{Status: ptr(StatusDone)},
				wantApproved: true,
			},
			{name: "an empty patch keeps approval", approved: true, patch: Patch{}, wantApproved: true},
			{name: "an unapproved task stays unapproved", approved: false, patch: Patch{Title: ptr("x")}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				s := New(t.TempDir())
				if err := s.Save(&Task{ID: tid1, Title: "T", Status: StatusDoing, Approved: tt.approved}); err != nil {
					t.Fatalf("seeding %s: %v", tid1, err)
				}

				got, err := s.Update(tid1, tt.patch)
				if err != nil {
					t.Fatalf("Update() returned error: %v", err)
				}

				loaded, err := s.Load(tid1)
				if err != nil {
					t.Fatalf("loading %s: %v", tid1, err)
				}

				if got.Approved != tt.wantApproved {
					t.Errorf("Update() Approved = %v, want %v", got.Approved, tt.wantApproved)
				}

				if loaded.Approved != tt.wantApproved {
					t.Errorf("Load() Approved = %v, want %v", loaded.Approved, tt.wantApproved)
				}
			})
		}
	})
}

func ptr[T any](v T) *T {
	return &v
}
