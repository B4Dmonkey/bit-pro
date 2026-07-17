package task

import (
	"errors"
	"io/fs"
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
	if *got != want {
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

func TestStoreDelete_ErrorsOnUnknownID(t *testing.T) {
	t.Parallel()

	err := New(t.TempDir()).Delete("BIT-99")

	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Delete() error = %v, want an error wrapping fs.ErrNotExist", err)
	}
	if !strings.Contains(err.Error(), "BIT-99") {
		t.Errorf("Delete() error = %q, want it to name the task ID", err)
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
