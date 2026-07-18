package tui

import (
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
	tea "github.com/charmbracelet/bubbletea"
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

func TestUpdate_ForwardsNavigationToList(t *testing.T) {
	t.Parallel()

	tasks := []*task.Task{
		{ID: "BIT-2"},
		{ID: "BIT-2.1"},
		{ID: "BIT-1"},
	}

	m := New(tasks)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})

	if got := updated.(model).Index(); got != 1 {
		t.Errorf("after KeyDown, Index() = %d, want 1", got)
	}
}

func TestNew_EmptyList(t *testing.T) {
	t.Parallel()

	m := New(nil)

	if got := len(m.Items()); got != 0 {
		t.Errorf("New(nil) produced %d items, want 0", got)
	}
}

func TestSelected_TracksCursor(t *testing.T) {
	t.Parallel()

	tasks := []*task.Task{
		{ID: "BIT-2"},
		{ID: "BIT-2.1"},
		{ID: "BIT-1"},
	}

	m := New(tasks)

	if got := m.selected().ID; got != tasks[0].ID {
		t.Errorf("default selected().ID = %q, want %q", got, tasks[0].ID)
	}

	m.Select(2)

	if got := m.selected().ID; got != tasks[2].ID {
		t.Errorf("after Select(2), selected().ID = %q, want %q", got, tasks[2].ID)
	}
}

func TestSelected_EmptyListNil(t *testing.T) {
	t.Parallel()

	m := New(nil)

	if got := m.selected(); got != nil {
		t.Errorf("selected() on empty list = %v, want nil", got)
	}
}
