package task

import (
	"errors"
	"io/fs"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestStorePath_ContainsUntrustedID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   string
		want string
	}{
		{name: "plain id", id: "BIT-1", want: ".bit/tasks/BIT-1.md"},
		{name: "traversal cannot escape the tasks dir", id: "../../README", want: ".bit/tasks/README.md"},
		{name: "deep traversal cannot escape", id: "../../../../etc/passwd", want: ".bit/tasks/etc/passwd.md"},
		{name: "absolute path cannot escape", id: "/etc/passwd", want: ".bit/tasks/etc/passwd.md"},
		{name: "illegal characters are stripped", id: "a:b*c", want: ".bit/tasks/abc.md"},
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
}

func TestStoreRelocate_MovesFileOutOfList(t *testing.T) {
	t.Parallel()

	s := New(t.TempDir())
	if err := s.Save(&Task{ID: "BIT-1", Status: "done"}); err != nil {
		t.Fatalf("seeding BIT-1: %v", err)
	}

	if err := s.Relocate("BIT-1", false); err != nil {
		t.Fatalf("Relocate() returned error: %v", err)
	}

	tasks, err := s.List()
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	if slices.ContainsFunc(tasks, func(t *Task) bool { return t.ID == "BIT-1" }) {
		t.Errorf("List() still contains BIT-1 after relocate")
	}
	if _, err := os.Stat(s.archivePath("BIT-1")); err != nil {
		t.Errorf("archived file: os.Stat error = %v, want the file to exist", err)
	}
	if _, err := os.Stat(s.Path("BIT-1")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("tasks file: os.Stat error = %v, want fs.ErrNotExist", err)
	}
}

func TestStoreRelocate_CascadesToBars(t *testing.T) {
	t.Parallel()

	s := New(t.TempDir())
	for _, id := range []string{"BIT-1", "BIT-1.1", "BIT-1.2"} {
		if err := s.Save(&Task{ID: id, Status: "done"}); err != nil {
			t.Fatalf("seeding %s: %v", id, err)
		}
	}

	if err := s.Relocate("BIT-1", false); err != nil {
		t.Fatalf("Relocate() returned error: %v", err)
	}

	tasks, err := s.List()
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("List() = %v, want no tasks after cascade", tasks)
	}
	for _, id := range []string{"BIT-1", "BIT-1.1", "BIT-1.2"} {
		if _, err := os.Stat(s.archivePath(id)); err != nil {
			t.Errorf("archived %s: os.Stat error = %v, want the file to exist", id, err)
		}
		if _, err := os.Stat(s.Path(id)); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("tasks %s: os.Stat error = %v, want fs.ErrNotExist", id, err)
		}
	}
}

func TestStoreRelocate_RefusesWithUnfinishedBars(t *testing.T) {
	t.Parallel()

	s := New(t.TempDir())
	for _, seed := range []struct{ id, status string }{
		{"BIT-1", "done"},
		{"BIT-1.1", "done"},
		{"BIT-1.2", "todo"},
	} {
		if err := s.Save(&Task{ID: seed.id, Status: seed.status}); err != nil {
			t.Fatalf("seeding %s: %v", seed.id, err)
		}
	}

	err := s.Relocate("BIT-1", false)

	var unfinished *UnfinishedBarsError
	if !errors.As(err, &unfinished) {
		t.Fatalf("Relocate() error = %v, want *UnfinishedBarsError", err)
	}
	if !slices.Contains(unfinished.Bars, "BIT-1.2") {
		t.Errorf("UnfinishedBarsError.Bars = %v, want it to contain BIT-1.2", unfinished.Bars)
	}
	for _, id := range []string{"BIT-1", "BIT-1.1", "BIT-1.2"} {
		if _, err := os.Stat(s.Path(id)); err != nil {
			t.Errorf("tasks %s: os.Stat error = %v, want the file to remain", id, err)
		}
		if _, err := os.Stat(s.archivePath(id)); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("archive %s: os.Stat error = %v, want fs.ErrNotExist", id, err)
		}
	}
}

func TestStoreRelocate_ForceOverridesGuard(t *testing.T) {
	t.Parallel()

	s := New(t.TempDir())
	for _, seed := range []struct{ id, status string }{
		{"BIT-1", "done"},
		{"BIT-1.1", "done"},
		{"BIT-1.2", "todo"},
	} {
		if err := s.Save(&Task{ID: seed.id, Status: seed.status}); err != nil {
			t.Fatalf("seeding %s: %v", seed.id, err)
		}
	}

	if err := s.Relocate("BIT-1", true); err != nil {
		t.Fatalf("Relocate() returned error: %v", err)
	}

	for _, id := range []string{"BIT-1", "BIT-1.1", "BIT-1.2"} {
		if _, err := os.Stat(s.archivePath(id)); err != nil {
			t.Errorf("archived %s: os.Stat error = %v, want the file to exist", id, err)
		}
		if _, err := os.Stat(s.Path(id)); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("tasks %s: os.Stat error = %v, want fs.ErrNotExist", id, err)
		}
	}
}

func TestStoreRelocate_ContainsUntrustedID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   string
		want string
	}{
		{name: "plain id", id: "BIT-1", want: ".bit/archive/BIT-1.md"},
		{name: "traversal cannot escape the archive dir", id: "../../README", want: ".bit/archive/README.md"},
		{name: "absolute path cannot escape", id: "/etc/passwd", want: ".bit/archive/etc/passwd.md"},
		{name: "illegal characters are stripped", id: "a:b*c", want: ".bit/archive/abc.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := New(".bit").archivePath(tt.id)
			if got != tt.want {
				t.Errorf("archivePath(%q) = %q, want %q", tt.id, got, tt.want)
			}
			if !strings.HasPrefix(got, ".bit/archive/") {
				t.Errorf("archivePath(%q) = %q, escaped the archive directory", tt.id, got)
			}
		})
	}
}

func TestStoreRelocate_DropsBarFromParentOrder(t *testing.T) {
	t.Parallel()

	s := New(t.TempDir())
	if err := s.Save(&Task{ID: "BIT-1", Status: "done", Order: []string{"BIT-1.1", "BIT-1.2", "BIT-1.3"}}); err != nil {
		t.Fatalf("seeding BIT-1: %v", err)
	}
	for _, id := range []string{"BIT-1.1", "BIT-1.2", "BIT-1.3"} {
		if err := s.Save(&Task{ID: id, Status: "done"}); err != nil {
			t.Fatalf("seeding %s: %v", id, err)
		}
	}

	if err := s.Relocate("BIT-1.2", false); err != nil {
		t.Fatalf("Relocate() returned error: %v", err)
	}

	got, err := s.Load("BIT-1")
	if err != nil {
		t.Fatalf("loading BIT-1: %v", err)
	}
	if want := []string{"BIT-1.1", "BIT-1.3"}; !slices.Equal(got.Order, want) {
		t.Errorf("Order = %v, want %v", got.Order, want)
	}
}

func TestStoreRelocate_LeavesLegacyOrderUnmaterialized(t *testing.T) {
	t.Parallel()

	s := New(t.TempDir())
	if err := s.Save(&Task{ID: "BIT-1", Status: "done", Order: nil}); err != nil {
		t.Fatalf("seeding BIT-1: %v", err)
	}
	for _, id := range []string{"BIT-1.1", "BIT-1.2"} {
		if err := s.Save(&Task{ID: id, Status: "done"}); err != nil {
			t.Fatalf("seeding %s: %v", id, err)
		}
	}

	if err := s.Relocate("BIT-1.1", false); err != nil {
		t.Fatalf("Relocate() returned error: %v", err)
	}

	got, err := s.Load("BIT-1")
	if err != nil {
		t.Fatalf("loading BIT-1: %v", err)
	}
	if len(got.Order) != 0 {
		t.Errorf("Order = %v, want empty", got.Order)
	}
}

func TestStoreNextID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		existing []string
		want     string
	}{
		{name: "no tasks yet", existing: nil, want: "BIT-1"},
		{name: "continues past the highest, not the count", existing: []string{"BIT-1", "BIT-3"}, want: "BIT-4"},
		{name: "ignores other prefixes", existing: []string{"BIT-1", "OTHER-9"}, want: "BIT-2"},
		{name: "ignores non-numeric suffixes", existing: []string{"BIT-1", "BIT-abc"}, want: "BIT-2"},
		{name: "handles multi-digit ids", existing: []string{"BIT-9", "BIT-10"}, want: "BIT-11"},
		{name: "ignores dotted children", existing: []string{"BIT-1", "BIT-1.1", "BIT-1.13"}, want: "BIT-2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := New(t.TempDir())
			for _, id := range tt.existing {
				if err := s.Save(&Task{ID: id, Title: "seed", Status: "todo"}); err != nil {
					t.Fatalf("seeding %s: %v", id, err)
				}
			}

			got, err := s.NextID("BIT")
			if err != nil {
				t.Fatalf("NextID() returned error: %v", err)
			}
			if got != tt.want {
				t.Errorf("NextID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStoreNextChildID_ErrorsWhenParentMissing(t *testing.T) {
	t.Parallel()

	_, err := New(t.TempDir()).NextChildID("BIT-99")

	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("NextChildID() error = %v, want an error wrapping fs.ErrNotExist", err)
	}
	if !strings.Contains(err.Error(), "BIT-99") {
		t.Errorf("NextChildID() error = %q, want it to name the parent ID", err)
	}
}

func TestStoreNextChildID_MintsWhenParentExists(t *testing.T) {
	t.Parallel()

	s := New(t.TempDir())
	if err := s.Save(&Task{ID: "BIT-1", Title: "seed", Status: "todo"}); err != nil {
		t.Fatalf("seeding BIT-1: %v", err)
	}

	got, err := s.NextChildID("BIT-1")
	if err != nil {
		t.Fatalf("NextChildID() returned error: %v", err)
	}
	if got != "BIT-1.1" {
		t.Errorf("NextChildID() = %q, want %q", got, "BIT-1.1")
	}
}

func TestStoreNextID_ReservesArchivedIDs(t *testing.T) {
	t.Parallel()

	s := New(t.TempDir())
	if err := s.Save(&Task{ID: "BIT-1", Title: "seed", Status: "todo"}); err != nil {
		t.Fatalf("seeding BIT-1: %v", err)
	}
	if err := s.Save(&Task{ID: "BIT-2", Title: "seed", Status: "done"}); err != nil {
		t.Fatalf("seeding BIT-2: %v", err)
	}
	if err := s.Relocate("BIT-2", false); err != nil {
		t.Fatalf("Relocate(BIT-2): %v", err)
	}

	got, err := s.NextID("BIT")
	if err != nil {
		t.Fatalf("NextID() returned error: %v", err)
	}
	if got != "BIT-3" {
		t.Errorf("NextID() = %q, want %q", got, "BIT-3")
	}
}

func TestStoreNextChildID_ReservesArchivedChildren(t *testing.T) {
	t.Parallel()

	s := New(t.TempDir())
	if err := s.Save(&Task{ID: "BIT-1", Title: "seed", Status: "todo"}); err != nil {
		t.Fatalf("seeding BIT-1: %v", err)
	}
	if err := s.Save(&Task{ID: "BIT-1.1", Title: "seed", Status: "done"}); err != nil {
		t.Fatalf("seeding BIT-1.1: %v", err)
	}
	if err := s.Relocate("BIT-1.1", false); err != nil {
		t.Fatalf("Relocate(BIT-1.1): %v", err)
	}

	got, err := s.NextChildID("BIT-1")
	if err != nil {
		t.Fatalf("NextChildID() returned error: %v", err)
	}
	if got != "BIT-1.2" {
		t.Errorf("NextChildID() = %q, want %q", got, "BIT-1.2")
	}
}

func TestStoreNextID_ReservesCompletedIDs(t *testing.T) {
	t.Parallel()

	s := New(t.TempDir())
	if err := s.Save(&Task{ID: "BIT-1", Title: "seed", Status: "todo"}); err != nil {
		t.Fatalf("seeding BIT-1: %v", err)
	}
	if err := s.Save(&Task{ID: "BIT-2", Title: "seed", Status: "done"}); err != nil {
		t.Fatalf("seeding BIT-2: %v", err)
	}
	if err := s.Complete("BIT-2"); err != nil {
		t.Fatalf("Complete(BIT-2): %v", err)
	}

	got, err := s.NextID("BIT")
	if err != nil {
		t.Fatalf("NextID() returned error: %v", err)
	}
	if got != "BIT-3" {
		t.Errorf("NextID() = %q, want %q", got, "BIT-3")
	}
}

func TestStoreNextChildID_ReservesCompletedChildren(t *testing.T) {
	t.Parallel()

	s := New(t.TempDir())
	if err := s.Save(&Task{ID: "BIT-1", Title: "seed", Status: "todo"}); err != nil {
		t.Fatalf("seeding BIT-1: %v", err)
	}
	if err := s.Save(&Task{ID: "BIT-1.1", Title: "seed", Status: "done"}); err != nil {
		t.Fatalf("seeding BIT-1.1: %v", err)
	}
	if err := s.Complete("BIT-1.1"); err != nil {
		t.Fatalf("Complete(BIT-1.1): %v", err)
	}

	got, err := s.NextChildID("BIT-1")
	if err != nil {
		t.Fatalf("NextChildID() returned error: %v", err)
	}
	if got != "BIT-1.2" {
		t.Errorf("NextChildID() = %q, want %q", got, "BIT-1.2")
	}
}

func TestStoreSaveLoad_RoundTrips(t *testing.T) {
	t.Parallel()

	s := New(t.TempDir())
	want := Task{ID: "BIT-1", Title: "Title", Status: "todo", Body: "Body.\n"}

	if err := s.Save(&want); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	got, err := s.Load("BIT-1")
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if !reflect.DeepEqual(*got, want) {
		t.Errorf("Load() = %+v, want %+v", *got, want)
	}
}

func TestStoreLoad_ErrorsOnUnknownID(t *testing.T) {
	t.Parallel()

	_, err := New(t.TempDir()).Load("BIT-99")

	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Load() error = %v, want an error wrapping fs.ErrNotExist", err)
	}
	if !strings.Contains(err.Error(), "BIT-99") {
		t.Errorf("Load() error = %q, want it to name the task ID", err)
	}
}

func TestStoreList_EmptyWhenNoTasksDir(t *testing.T) {
	t.Parallel()

	tasks, err := New(t.TempDir()).List()
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("List() = %v, want no tasks", tasks)
	}
}

func TestStoreList_OrdersBarsByExplicitOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		order []string
		want  []string
	}{
		{
			name:  "explicit order overrides id sequence",
			order: []string{"BIT-1.2", "BIT-1.1"},
			want:  []string{"BIT-1", "BIT-1.2", "BIT-1.1"},
		},
		{
			name:  "no order falls back to id sequence",
			order: nil,
			want:  []string{"BIT-1", "BIT-1.1", "BIT-1.2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := New(t.TempDir())
			if err := s.Save(&Task{ID: "BIT-1", Title: "track", Status: "todo", Order: tt.order}); err != nil {
				t.Fatalf("seeding BIT-1: %v", err)
			}
			for _, id := range []string{"BIT-1.1", "BIT-1.2"} {
				if err := s.Save(&Task{ID: id, Title: "bar", Status: "todo"}); err != nil {
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
}

func TestStoreMove_Resequences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		order  []string
		id     string
		anchor string
		before bool
		want   []string
	}{
		{
			name:   "materializes then moves to front",
			order:  nil,
			id:     "BIT-1.3",
			anchor: "BIT-1.1",
			before: true,
			want:   []string{"BIT-1.3", "BIT-1.1", "BIT-1.2"},
		},
		{
			name:   "splices an existing order to the back",
			order:  []string{"BIT-1.1", "BIT-1.2", "BIT-1.3"},
			id:     "BIT-1.1",
			anchor: "BIT-1.3",
			before: false,
			want:   []string{"BIT-1.2", "BIT-1.3", "BIT-1.1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := New(t.TempDir())
			if err := s.Save(&Task{ID: "BIT-1", Title: "track", Status: "todo", Order: tt.order}); err != nil {
				t.Fatalf("seeding BIT-1: %v", err)
			}
			for _, id := range []string{"BIT-1.1", "BIT-1.2", "BIT-1.3"} {
				if err := s.Save(&Task{ID: id, Title: "bar", Status: "todo"}); err != nil {
					t.Fatalf("seeding %s: %v", id, err)
				}
			}

			if err := s.Move(tt.id, tt.anchor, tt.before); err != nil {
				t.Fatalf("Move() returned error: %v", err)
			}

			got, err := s.Load("BIT-1")
			if err != nil {
				t.Fatalf("loading BIT-1: %v", err)
			}
			if !slices.Equal(got.Order, tt.want) {
				t.Errorf("Order = %v, want %v", got.Order, tt.want)
			}
		})
	}
}

func TestStoreMove_Rejects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		id           string
		anchor       string
		wantNotExist bool
	}{
		{name: "anchor under a different track", id: "BIT-1.1", anchor: "BIT-2.1"},
		{name: "unknown bar", id: "BIT-1.9", anchor: "BIT-1.1", wantNotExist: true},
		{name: "unknown anchor", id: "BIT-1.1", anchor: "BIT-1.9", wantNotExist: true},
		{name: "moving a bar relative to itself", id: "BIT-1.1", anchor: "BIT-1.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := New(t.TempDir())
			for _, id := range []string{"BIT-1", "BIT-2"} {
				if err := s.Save(&Task{ID: id, Title: "track", Status: "todo"}); err != nil {
					t.Fatalf("seeding %s: %v", id, err)
				}
			}
			for _, id := range []string{"BIT-1.1", "BIT-1.2", "BIT-2.1"} {
				if err := s.Save(&Task{ID: id, Title: "bar", Status: "todo"}); err != nil {
					t.Fatalf("seeding %s: %v", id, err)
				}
			}

			err := s.Move(tt.id, tt.anchor, false)
			if err == nil {
				t.Fatalf("Move(%q, %q) returned nil error, want non-nil", tt.id, tt.anchor)
			}
			if tt.wantNotExist && !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("Move(%q, %q) error = %v, want it to wrap fs.ErrNotExist", tt.id, tt.anchor, err)
			}
		})
	}
}

func TestStoreConfig_RoundTrips(t *testing.T) {
	t.Parallel()

	s := New(t.TempDir())
	if err := s.SaveConfig(&Config{Prefix: "BIT"}); err != nil {
		t.Fatalf("SaveConfig() returned error: %v", err)
	}

	got, err := s.Config()
	if err != nil {
		t.Fatalf("Config() returned error: %v", err)
	}
	if got.Prefix != "BIT" {
		t.Errorf("Config().Prefix = %q, want %q", got.Prefix, "BIT")
	}
}

func TestCompareIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b string
		want int
	}{
		{name: "newer sorts before older", a: "BIT-2", b: "BIT-1", want: -1},
		{name: "older sorts after newer", a: "BIT-1", b: "BIT-2", want: 1},
		{name: "equal ids", a: "BIT-1", b: "BIT-1", want: 0},
		{name: "two-digit id sorts before one-digit", a: "BIT-10", b: "BIT-9", want: -1},
		{name: "unparseable suffix sorts last", a: "BIT-abc", b: "BIT-1", want: 1},
		{name: "track heads its own bars", a: "BIT-2", b: "BIT-2.1", want: -1},
		{name: "bars ascend, not lexically", a: "BIT-2.1", b: "BIT-2.13", want: -1},
		{name: "track dominates bar", a: "BIT-2.1", b: "BIT-1.9", want: -1},
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

func TestStoreConfig_ErrorsWhenAbsent(t *testing.T) {
	t.Parallel()

	if _, err := New(t.TempDir()).Config(); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Config() error = %v, want an error wrapping fs.ErrNotExist", err)
	}
}
