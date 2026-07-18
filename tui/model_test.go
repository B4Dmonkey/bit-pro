package tui

import (
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
)

func TestNew_PreservesStoreOrder(t *testing.T) {
	t.Parallel()

	tasks := []*task.Task{
		{ID: "BIT-2"},
		{ID: "BIT-2.1"},
		{ID: "BIT-1"},
	}

	m := New(tasks)

	items := m.Items()
	if len(items) != len(tasks) {
		t.Fatalf("New produced %d items, want %d", len(items), len(tasks))
	}
	for i, it := range items {
		got := it.(item).t.ID
		if got != tasks[i].ID {
			t.Errorf("item[%d].ID = %q, want %q", i, got, tasks[i].ID)
		}
	}
}

func TestNew_EmptyList(t *testing.T) {
	t.Parallel()

	m := New(nil)

	if got := len(m.Items()); got != 0 {
		t.Errorf("New(nil) produced %d items, want 0", got)
	}
}
