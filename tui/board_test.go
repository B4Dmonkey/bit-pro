package tui

import (
	"slices"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
	tea "charm.land/bubbletea/v2"
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

func TestGroupByStatus_PreservesOrderWithinColumn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		tasks  []*task.Task
		column int
		want   []string
	}{
		{
			name: "todo column keeps non-ID incoming order",
			tasks: []*task.Task{
				{ID: "BIT-1.2", Status: "todo"},
				{ID: "BIT-1.1", Status: "todo"},
				{ID: "BIT-1.3", Status: "todo"},
			},
			column: 0,
			want:   []string{"BIT-1.2", "BIT-1.1", "BIT-1.3"},
		},
		{
			name: "done column keeps non-ID incoming order",
			tasks: []*task.Task{
				{ID: "BIT-1.3", Status: "done"},
				{ID: "BIT-1.1", Status: "done"},
				{ID: "BIT-1.2", Status: "done"},
			},
			column: 2,
			want:   []string{"BIT-1.3", "BIT-1.1", "BIT-1.2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cols := groupByStatus(tt.tasks)

			var got []string
			for _, tk := range cols[tt.column] {
				got = append(got, tk.ID)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("column %d order = %v, want %v", tt.column, got, tt.want)
			}
		})
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

func TestView_BoardColumnCounts(t *testing.T) {
	t.Parallel()

	tasks := []*task.Task{
		{ID: "BIT-1", Status: "todo"},
		{ID: "BIT-2", Status: "todo"},
		{ID: "BIT-3", Status: "todo"},
		{ID: "BIT-4", Status: "todo"},
		{ID: "BIT-5", Status: "doing"},
		{ID: "BIT-6", Status: "done"},
		{ID: "BIT-7", Status: "done"},
	}

	var mdl tea.Model = New(tasks)
	mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	view := mdl.(model).View().Content
	for _, want := range []string{"To Do (4)", "Doing (1)", "Done (2)"} {
		if !strings.Contains(view, want) {
			t.Errorf("board View() missing %q", want)
		}
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

func TestUpdate_BoardActiveColumn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		keys []rune
		want int
	}{
		{"default first column", nil, 0},
		{"right advances", []rune{tea.KeyRight}, 1},
		{"right twice", []rune{tea.KeyRight, tea.KeyRight}, 2},
		{"right clamps at last", []rune{tea.KeyRight, tea.KeyRight, tea.KeyRight}, 2},
		{"left retreats from last", []rune{tea.KeyRight, tea.KeyRight, tea.KeyLeft}, 1},
		{"left clamps at first", []rune{tea.KeyLeft, tea.KeyLeft, tea.KeyLeft}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var mdl tea.Model = New([]*task.Task{
				{ID: "BIT-1", Status: "todo"},
				{ID: "BIT-2", Status: "doing"},
				{ID: "BIT-3", Status: "done"},
			})
			mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
			for _, k := range tt.keys {
				mdl, _ = mdl.Update(tea.KeyPressMsg{Code: k})
			}

			if got := mdl.(model).activeCol; got != tt.want {
				t.Errorf("activeCol = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestUpdate_BoardCardSelection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		keys    []rune
		wantCol int
		wantIdx int
	}{
		{"empty column stays at zero", []rune{tea.KeyDown}, 0, 0},
		{"down advances in active column", []rune{tea.KeyRight, tea.KeyRight, tea.KeyDown}, 2, 1},
		{"down clamps at last card", []rune{tea.KeyRight, tea.KeyRight, tea.KeyDown, tea.KeyDown, tea.KeyDown}, 2, 2},
		{"selection survives column round trip", []rune{tea.KeyRight, tea.KeyRight, tea.KeyDown, tea.KeyLeft, tea.KeyLeft, tea.KeyRight, tea.KeyRight}, 2, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var mdl tea.Model = New([]*task.Task{
				{ID: "BIT-1", Status: "doing"},
				{ID: "BIT-2", Status: "done"},
				{ID: "BIT-3", Status: "done"},
				{ID: "BIT-4", Status: "done"},
			})
			mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
			mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
			for _, k := range tt.keys {
				mdl, _ = mdl.Update(tea.KeyPressMsg{Code: k})
			}

			if got := mdl.(model).boardCols[tt.wantCol].Index(); got != tt.wantIdx {
				t.Errorf("boardCols[%d].Index() = %d, want %d", tt.wantCol, got, tt.wantIdx)
			}
		})
	}
}

func TestView_BoardHelp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		toBoard  bool
		contains []string
		absent   []string
	}{
		{"list mode shows focus", false, []string{"focus"}, []string{"column"}},
		{"board mode shows column and card", true, []string{"column", "card"}, []string{"focus"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var mdl tea.Model = New([]*task.Task{
				{ID: "BIT-1", Status: "todo"},
				{ID: "BIT-2", Status: "doing"},
				{ID: "BIT-3", Status: "done"},
			})
			mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
			if tt.toBoard {
				mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
			}

			view := mdl.(model).View().Content
			for _, want := range tt.contains {
				if !strings.Contains(view, want) {
					t.Errorf("View() missing %q", want)
				}
			}
			for _, notWant := range tt.absent {
				if strings.Contains(view, notWant) {
					t.Errorf("View() contains %q, want absent", notWant)
				}
			}
		})
	}
}

func TestUpdate_BoardEnterOpensModal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		enter bool
		want  bool
	}{
		{"no key leaves modal closed", false, false},
		{"enter opens modal", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var mdl tea.Model = New([]*task.Task{{ID: "BIT-1", Status: "todo", Body: "body"}})
			mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
			if tt.enter {
				mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			}

			if got := mdl.(model).modalOpen; got != tt.want {
				t.Errorf("modalOpen = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestView_ModalShowsBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		enter bool
		want  bool
	}{
		{"closed hides body", false, false},
		{"open shows body", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var mdl tea.Model = New([]*task.Task{{ID: "BIT-1", Status: "todo", Title: "T", Body: "MODALBODYTOKEN"}})
			mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
			if tt.enter {
				mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			}

			if got := strings.Contains(mdl.(model).View().Content, "MODALBODYTOKEN"); got != tt.want {
				t.Errorf("View contains body token = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdate_ModalCloses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  tea.KeyPressMsg
	}{
		{"q", tea.KeyPressMsg{Code: 'q', Text: "q"}},
		{"esc", tea.KeyPressMsg{Code: tea.KeyEsc}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var mdl tea.Model = New([]*task.Task{{ID: "BIT-1", Status: "todo", Body: "body"}})
			mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
			mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

			mdl, cmd := mdl.Update(tt.key)

			if mdl.(model).modalOpen {
				t.Errorf("modalOpen = true, want false")
			}
			if cmd != nil {
				t.Errorf("cmd = %T, want nil", cmd())
			}
		})
	}
}

func TestUpdate_BoardEnterEmptyColumnNoop(t *testing.T) {
	t.Parallel()

	var mdl tea.Model = New([]*task.Task{{ID: "BIT-1", Status: "doing", Body: "body"}})
	mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if got := mdl.(model).modalOpen; got {
		t.Errorf("modalOpen = %v, want false", got)
	}
}

func TestUpdate_BoardQuits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  tea.KeyPressMsg
	}{
		{"q", tea.KeyPressMsg{Code: 'q', Text: "q"}},
		{"esc", tea.KeyPressMsg{Code: tea.KeyEsc}},
		{"ctrl+c", tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var mdl tea.Model = New([]*task.Task{{ID: "BIT-1", Status: "todo"}})
			mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})

			_, cmd := mdl.Update(tt.key)

			if cmd == nil {
				t.Fatalf("%s in board mode: cmd = nil, want a quit cmd", tt.name)
			}
			if _, ok := cmd().(tea.QuitMsg); !ok {
				t.Errorf("%s in board mode: cmd() = %T, want tea.QuitMsg", tt.name, cmd())
			}
		})
	}
}
