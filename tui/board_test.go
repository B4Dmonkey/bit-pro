package tui

import (
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
)

func TestGroupByStatus(t *testing.T) {
	t.Parallel()

	tasks := []*task.Task{
		{ID: "BIT-4", Status: "todo"},
		{ID: "BIT-2.1", Status: "doing"},
		{ID: "BIT-4.1", Status: "done"},
		{ID: "BIT-4.2", Status: "done"},
		{ID: "BIT-9", Status: "backlog"},
	}

	cols := groupByStatus(tasks)

	want := [3][]string{
		{"BIT-4"},
		{"BIT-2.1"},
		{"BIT-4.1", "BIT-4.2"},
	}
	for i, wantIDs := range want {
		if len(cols[i]) != len(wantIDs) {
			t.Fatalf("column %d has %d tasks, want %d", i, len(cols[i]), len(wantIDs))
		}
		for j, id := range wantIDs {
			if cols[i][j].ID != id {
				t.Errorf("column %d task %d = %q, want %q", i, j, cols[i][j].ID, id)
			}
		}
	}

	for i, col := range cols {
		for _, tk := range col {
			if tk.ID == "BIT-9" {
				t.Errorf("unmapped task BIT-9 leaked into column %d", i)
			}
		}
	}
}

func TestBoardColumns_FromGrouping(t *testing.T) {
	t.Parallel()

	tasks := []*task.Task{
		{ID: "BIT-4", Status: "todo"},
		{ID: "BIT-2.1", Status: "doing"},
		{ID: "BIT-4.1", Status: "done"},
		{ID: "BIT-4.2", Status: "done"},
	}

	m := New(tasks)

	want := [3]int{1, 1, 2}
	for i, n := range want {
		if got := len(m.boardCols[i].Items()); got != n {
			t.Errorf("column %d has %d items, want %d", i, got, n)
		}
	}

	first, ok := m.boardCols[2].Items()[0].(item)
	if !ok {
		t.Fatalf("column 2 item 0 is %T, want item", m.boardCols[2].Items()[0])
	}
	if first.t.ID != "BIT-4.1" {
		t.Errorf("column 2 item 0 = %q, want %q", first.t.ID, "BIT-4.1")
	}
}

func TestGroupByStatus_Empty(t *testing.T) {
	t.Parallel()

	cols := groupByStatus(nil)

	for i, col := range cols {
		if len(col) != 0 {
			t.Errorf("column %d has %d tasks, want 0", i, len(col))
		}
	}
}
