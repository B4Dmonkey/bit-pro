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

func TestSplitWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		total int
	}{
		{"zero width", 0},
		{"one column", 1},
		{"typical terminal", 120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			listW, detailW := splitWidth(tt.total)

			if listW < 0 {
				t.Errorf("splitWidth(%d) listW = %d, want >= 0", tt.total, listW)
			}
			if detailW < 0 {
				t.Errorf("splitWidth(%d) detailW = %d, want >= 0", tt.total, detailW)
			}
			if listW+detailW > tt.total {
				t.Errorf("splitWidth(%d) listW+detailW = %d, want <= %d", tt.total, listW+detailW, tt.total)
			}
		})
	}

	t.Run("typical terminal splits 40/60 with detail wider", func(t *testing.T) {
		t.Parallel()

		listW, detailW := splitWidth(120)

		if listW <= 0 || detailW <= 0 {
			t.Fatalf("splitWidth(120) = (%d, %d), want both > 0", listW, detailW)
		}
		if detailW <= listW {
			t.Errorf("splitWidth(120) detailW = %d, listW = %d, want detailW > listW", detailW, listW)
		}
	})
}

func TestUpdate_EscQuitsFromList(t *testing.T) {
	t.Parallel()

	m := New([]*task.Task{{ID: "BIT-1"}})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("after KeyEsc in list, cmd = nil, want a quit cmd")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("after KeyEsc in list, cmd() = %T, want tea.QuitMsg", cmd())
	}
}
