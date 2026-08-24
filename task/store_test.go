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
		{name: "plain id", id: tid1, want: ".bit/tasks/BIT-1.md"},
		{name: "traversal cannot escape the tasks dir", id: "../../README", want: ".bit/tasks/README.md"},
		{name: "deep traversal cannot escape", id: "../../../../etc/passwd", want: ".bit/tasks/ETC/PASSWD.md"},
		{name: "absolute path cannot escape", id: "/etc/passwd", want: ".bit/tasks/ETC/PASSWD.md"},
		{name: "illegal characters are stripped", id: "a:b*c", want: ".bit/tasks/ABC.md"},
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

	if _, err := os.Stat(s.archivePath(tid1)); err != nil {
		t.Errorf("archived file: os.Stat error = %v, want the file to exist", err)
	}

	if _, err := os.Stat(s.Path(tid1)); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("tasks file: os.Stat error = %v, want fs.ErrNotExist", err)
	}
}

func TestStoreRelocate_CascadesToBars(t *testing.T) {
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
}

func TestStoreRelocate_RefusesWithUnfinishedBars(t *testing.T) {
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
}

func TestStoreRelocate_ForceOverridesGuard(t *testing.T) {
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
}

func TestStoreRelocate_ContainsUntrustedID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   string
		want string
	}{
		{name: "plain id", id: tid1, want: ".bit/archive/tasks/BIT-1.md"},
		{name: "traversal cannot escape the archive dir", id: "../../README", want: ".bit/archive/tasks/README.md"},
		{name: "absolute path cannot escape", id: "/etc/passwd", want: ".bit/archive/tasks/ETC/PASSWD.md"},
		{name: "illegal characters are stripped", id: "a:b*c", want: ".bit/archive/tasks/ABC.md"},
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
}

func TestStoreRelocate_DropsBarFromParentOrder(t *testing.T) {
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
}

func TestStoreRelocate_LeavesLegacyOrderUnmaterialized(t *testing.T) {
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
}

func TestStoreNextID_ReservesArchivedIDs(t *testing.T) {
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
}

func TestStoreNextChildID_ReservesArchivedChildren(t *testing.T) {
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
}

func TestStoreNextID_ReservesCompletedIDs(t *testing.T) {
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
}

func TestStoreNextChildID_ReservesCompletedChildren(t *testing.T) {
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
}

func TestStoreSaveLoad_RoundTrips(t *testing.T) {
	t.Parallel()

	s := New(t.TempDir())
	want := Task{ID: tid1, Title: "Title", Status: StatusTodo, Body: "Body.\n"}

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
}

func TestStoreMove_Resequences(t *testing.T) {
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
}

func TestStoreMove_Rejects(t *testing.T) {
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
}

func TestStoreMove_RejectsAnchorPair(t *testing.T) {
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
}

func TestStoreConfig_RoundTrips(t *testing.T) {
	t.Parallel()

	s := New(t.TempDir())
	if err := s.SaveConfig(&Config{Prefix: tprefix}); err != nil {
		t.Fatalf("SaveConfig() returned error: %v", err)
	}

	got, err := s.Config()
	if err != nil {
		t.Fatalf("Config() returned error: %v", err)
	}

	if got.Prefix != tprefix {
		t.Errorf("Config().Prefix = %q, want %q", got.Prefix, tprefix)
	}
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

func TestStoreConfig_ErrorsWhenAbsent(t *testing.T) {
	t.Parallel()

	if _, err := New(t.TempDir()).Config(); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Config() error = %v, want an error wrapping fs.ErrNotExist", err)
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

			s := New(t.TempDir())
			if err := s.SaveConfig(&Config{Prefix: tprefix}); err != nil {
				t.Fatalf("SaveConfig() returned error: %v", err)
			}

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
}

func TestStoreCreate_RejectsUnknownParent(t *testing.T) {
	t.Parallel()

	s := New(t.TempDir())

	_, err := s.Create(CreateParams{Title: tbar, Parent: tid1_9})

	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Create() error = %v, want an error wrapping fs.ErrNotExist", err)
	}
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

func TestStoreUpdate_AppliesOnlySetFields(t *testing.T) {
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
}

func TestStoreUpdate_ApprovalRevocation(t *testing.T) {
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
}

func ptr[T any](v T) *T {
	return &v
}
