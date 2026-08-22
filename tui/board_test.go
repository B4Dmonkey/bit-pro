package tui

import (
	"slices"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/B4Dmonkey/bit-pro/task"
)

func TestGroupByStatus(t *testing.T) {
	t.Parallel()

	tasks := []*task.Task{
		{ID: ttid4, Status: task.StatusTodo, Approved: true},
		{ID: ttid2_1, Status: task.StatusDoing},
		{ID: ttid4_1, Status: task.StatusDone},
		{ID: ttid4_2, Status: task.StatusDone},
		{ID: ttid9, Status: "backlog"},
	}

	cols := groupByStatus(tasks)

	want := [3][]string{
		{ttid4},
		{ttid2_1},
		{ttid4_1, ttid4_2},
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
			if tk.ID == ttid9 {
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
				{ID: ttid1_2, Status: task.StatusTodo, Approved: true},
				{ID: ttid1_1, Status: task.StatusTodo, Approved: true},
				{ID: ttid1_3, Status: task.StatusTodo, Approved: true},
			},
			column: 0,
			want:   []string{ttid1_2, ttid1_1, ttid1_3},
		},
		{
			name: "done column keeps non-ID incoming order",
			tasks: []*task.Task{
				{ID: ttid1_3, Status: task.StatusDone},
				{ID: ttid1_1, Status: task.StatusDone},
				{ID: ttid1_2, Status: task.StatusDone},
			},
			column: 2,
			want:   []string{ttid1_3, ttid1_1, ttid1_2},
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
		{ID: ttid4, Status: task.StatusTodo, Approved: true},
		{ID: ttid2_1, Status: task.StatusDoing},
		{ID: ttid4_1, Status: task.StatusDone},
		{ID: ttid4_2, Status: task.StatusDone},
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

	if first.t.ID != ttid4_1 {
		t.Errorf("column 2 item 0 = %q, want %q", first.t.ID, ttid4_1)
	}
}

func TestDefaultColumn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cols [3][]*task.Task
		want int
	}{
		{
			name: "doing is the default when populated",
			cols: [3][]*task.Task{nil, {{ID: ttid1}}, nil},
			want: 1,
		},
		{
			name: "doing wins even when to do is also populated",
			cols: [3][]*task.Task{{{ID: ttid4}}, {{ID: ttid1}}, {{ID: ttid9}}},
			want: 1,
		},
		{
			name: "falls back to to do when doing is empty",
			cols: [3][]*task.Task{{{ID: ttid4}}, nil, nil},
			want: 0,
		},
		{
			name: "falls back to done when only done has tasks",
			cols: [3][]*task.Task{nil, nil, {{ID: ttid4}}},
			want: 2,
		},
		{
			name: "all empty defaults to zero",
			cols: [3][]*task.Task{},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := defaultColumn(tt.cols); got != tt.want {
				t.Errorf("defaultColumn(%v) = %d, want %d", tt.cols, got, tt.want)
			}
		})
	}
}

func TestFirstBarIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		items []list.Item
		want  int
	}{
		{
			name: "bar after its track",
			items: []list.Item{
				item{t: &task.Task{ID: ttid1}},
				item{t: &task.Task{ID: ttid1_1}},
			},
			want: 1,
		},
		{
			name: "bar before its track",
			items: []list.Item{
				item{t: &task.Task{ID: ttid2_1}},
				item{t: &task.Task{ID: ttid2}},
			},
			want: 0,
		},
		{
			name: "no bars falls back to first row",
			items: []list.Item{
				item{t: &task.Task{ID: ttid3}},
				item{t: &task.Task{ID: ttid4}},
			},
			want: 0,
		},
		{
			name:  "empty column",
			items: []list.Item{},
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := firstBarIndex(tt.items); got != tt.want {
				t.Errorf("firstBarIndex(%v) = %d, want %d", tt.items, got, tt.want)
			}
		})
	}
}

func TestFlattenBoard(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cols [3]list.Model
		want []boardEntry
	}{
		{
			name: "single column",
			cols: [3]list.Model{{}, newColumnList([]*task.Task{{ID: ttid1}, {ID: ttid1_1}}, nil), {}},
			want: []boardEntry{
				{col: 1, pos: 0, t: &task.Task{ID: ttid1}},
				{col: 1, pos: 1, t: &task.Task{ID: ttid1_1}},
			},
		},
		{
			name: "all three columns concatenate in order",
			cols: [3]list.Model{
				newColumnList([]*task.Task{{ID: ttid2}}, nil),
				newColumnList([]*task.Task{{ID: ttid1}, {ID: ttid1_1}}, nil),
				newColumnList([]*task.Task{{ID: ttid3}}, nil),
			},
			want: []boardEntry{
				{col: 0, pos: 0, t: &task.Task{ID: ttid2}},
				{col: 1, pos: 0, t: &task.Task{ID: ttid1}},
				{col: 1, pos: 1, t: &task.Task{ID: ttid1_1}},
				{col: 2, pos: 0, t: &task.Task{ID: ttid3}},
			},
		},
		{
			name: "empty columns are skipped not padded",
			cols: [3]list.Model{{}, newColumnList([]*task.Task{{ID: ttid1}}, nil), {}},
			want: []boardEntry{
				{col: 1, pos: 0, t: &task.Task{ID: ttid1}},
			},
		},
		{
			name: "all empty returns empty not nil panic",
			cols: [3]list.Model{{}, {}, {}},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := flattenBoard(tt.cols)
			if len(got) != len(tt.want) {
				t.Fatalf("flattenBoard() = %d entries, want %d", len(got), len(tt.want))
			}

			for i, e := range got {
				if e.col != tt.want[i].col || e.pos != tt.want[i].pos || e.t.ID != tt.want[i].t.ID {
					t.Errorf("entry %d = {col:%d pos:%d id:%q}, want {col:%d pos:%d id:%q}",
						i, e.col, e.pos, e.t.ID, tt.want[i].col, tt.want[i].pos, tt.want[i].t.ID)
				}
			}
		})
	}
}

func TestView_BoardColumnCounts(t *testing.T) {
	t.Parallel()

	tasks := []*task.Task{
		{ID: ttid1, Status: task.StatusTodo, Approved: true},
		{ID: ttid2, Status: task.StatusTodo, Approved: true},
		{ID: ttid3, Status: task.StatusTodo, Approved: true},
		{ID: ttid4, Status: task.StatusTodo, Approved: true},
		{ID: ttid5, Status: task.StatusDoing},
		{ID: "BIT-6", Status: task.StatusDone},
		{ID: "BIT-7", Status: task.StatusDone},
	}

	var mdl tea.Model = New(tasks)

	mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

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
				{ID: ttid1, Status: task.StatusTodo, Approved: true},
				{ID: ttid3, Status: task.StatusDone},
			})
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
		{
			"selection survives column round trip",
			[]rune{tea.KeyRight, tea.KeyRight, tea.KeyDown, tea.KeyLeft, tea.KeyLeft, tea.KeyRight, tea.KeyRight},
			2, 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var mdl tea.Model = New([]*task.Task{
				{ID: ttid1, Status: task.StatusDoing},
				{ID: ttid2, Status: task.StatusDone},
				{ID: ttid3, Status: task.StatusDone},
				{ID: ttid4, Status: task.StatusDone},
			})

			mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
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
		{"list mode shows focus", false, []string{ttFocus}, []string{"column"}},
		{"board mode shows column and card", true, []string{"column", "card"}, []string{ttFocus}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var mdl tea.Model = New([]*task.Task{
				{ID: ttid1, Status: task.StatusTodo},
				{ID: ttid2, Status: task.StatusDoing},
				{ID: ttid3, Status: task.StatusDone},
			})

			mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
			if !tt.toBoard {
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

			var mdl tea.Model = New([]*task.Task{{
				ID: ttid1, Status: task.StatusTodo, Approved: true, Body: ttBody,
			}})

			mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
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

			var mdl tea.Model = New([]*task.Task{{
				ID: ttid1, Status: task.StatusTodo, Approved: true, Title: "T", Body: "MODALBODYTOKEN",
			}})

			mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
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
		{keyEsc, tea.KeyPressMsg{Code: tea.KeyEsc}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var mdl tea.Model = New([]*task.Task{{
				ID: ttid1, Status: task.StatusTodo, Approved: true, Body: ttBody,
			}})

			mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
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

func TestUpdate_ModalCapturesInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		key   tea.KeyPressMsg
		check func(t *testing.T, mdl tea.Model, cmd tea.Cmd)
	}{
		{
			"ctrl+c quits",
			tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl},
			func(t *testing.T, mdl tea.Model, cmd tea.Cmd) {
				if cmd == nil {
					t.Fatalf("cmd = nil, want quit")
				}

				if _, ok := cmd().(tea.QuitMsg); !ok {
					t.Errorf("cmd() = %T, want tea.QuitMsg", cmd())
				}
			},
		},
		{
			"right pages to next task instead of switching column the old way",
			tea.KeyPressMsg{Code: tea.KeyRight},
			func(t *testing.T, mdl tea.Model, _ tea.Cmd) {
				if got := mdl.(model).boardSelected().ID; got != ttid3 {
					t.Errorf("boardSelected().ID = %q, want %q", got, ttid3)
				}

				if !mdl.(model).modalOpen {
					t.Errorf("modalOpen = false, want true")
				}
			},
		},
		{
			"tab swallowed",
			tea.KeyPressMsg{Code: tea.KeyTab},
			func(t *testing.T, mdl tea.Model, _ tea.Cmd) {
				if got := mdl.(model).mode; got != modeBoard {
					t.Errorf("mode = %v, want modeBoard", got)
				}

				if !mdl.(model).modalOpen {
					t.Errorf("modalOpen = false, want true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var mdl tea.Model = New([]*task.Task{
				{ID: ttid1, Status: task.StatusTodo, Body: ttBody},
				{ID: ttid2, Status: task.StatusDoing, Body: ttBody},
				{ID: ttid3, Status: task.StatusDone, Body: ttBody},
			})

			mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

			mdl, cmd := mdl.Update(tt.key)

			tt.check(t, mdl, cmd)
		})
	}
}

func TestUpdate_ModalScrollsLongBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  tea.KeyPressMsg
	}{
		{"down arrow", tea.KeyPressMsg{Code: tea.KeyDown}},
		{"j", tea.KeyPressMsg{Code: 'j', Text: "j"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var mdl tea.Model = New([]*task.Task{{
				ID: ttid1, Status: task.StatusTodo, Approved: true, Title: "T", Body: strings.Repeat("line\n", 500),
			}})

			mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

			opened := mdl.(model)
			if got := opened.modalViewport.YOffset(); got != 0 {
				t.Fatalf("YOffset after open = %d, want 0", got)
			}

			mdl, _ = mdl.Update(tt.key)

			scrolled := mdl.(model)
			if got := scrolled.modalViewport.YOffset(); got <= 0 {
				t.Errorf("YOffset after %s = %d, want > 0", tt.name, got)
			}

			if h := lipgloss.Height(scrolled.View().Content); h > 24 {
				t.Errorf("View height = %d, want <= 24", h)
			}
		})
	}
}

func TestUpdate_ModalPagesWithinColumn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  tea.KeyPressMsg
	}{
		{"right", tea.KeyPressMsg{Code: tea.KeyRight}},
		{"l", tea.KeyPressMsg{Code: 'l', Text: "l"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var mdl tea.Model = New([]*task.Task{
				{ID: ttid1, Status: task.StatusDoing},
				{ID: ttid1_1, Status: task.StatusDoing},
			})

			mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyUp})
			mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

			mdl, _ = mdl.Update(tt.key)

			opened := mdl.(model)
			if got := opened.boardSelected().ID; got != ttid1_1 {
				t.Errorf("boardSelected().ID = %q, want %q", got, ttid1_1)
			}

			if !opened.modalOpen {
				t.Errorf("modalOpen = false, want true")
			}
		})
	}
}

func TestUpdate_ModalPagesAcrossColumns(t *testing.T) {
	t.Parallel()

	var mdl tea.Model = New([]*task.Task{
		{ID: ttid2, Status: task.StatusTodo, Approved: true},
		{ID: ttid1, Status: task.StatusDoing},
		{ID: ttid1_1, Status: task.StatusDoing},
	})

	mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	steps := []struct {
		key     tea.KeyPressMsg
		wantID  string
		wantCol int
	}{
		{tea.KeyPressMsg{Code: tea.KeyLeft}, ttid1, 1},
		{tea.KeyPressMsg{Code: tea.KeyLeft}, ttid2, 0},
		{tea.KeyPressMsg{Code: tea.KeyRight}, ttid1, 1},
		{tea.KeyPressMsg{Code: tea.KeyRight}, ttid1_1, 1},
	}

	for _, s := range steps {
		mdl, _ = mdl.Update(s.key)

		got := mdl.(model)
		if id := got.boardSelected().ID; id != s.wantID {
			t.Fatalf("boardSelected().ID = %q, want %q", id, s.wantID)
		}

		if got.activeCol != s.wantCol {
			t.Fatalf("activeCol = %d, want %d", got.activeCol, s.wantCol)
		}
	}
}

func TestUpdate_ModalPagingClampsAtEnds(t *testing.T) {
	t.Parallel()

	var mdl tea.Model = New([]*task.Task{
		{ID: ttid2, Status: task.StatusTodo, Approved: true},
		{ID: ttid1, Status: task.StatusDoing},
		{ID: ttid1_1, Status: task.StatusDoing},
	})

	mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if got := mdl.(model).boardSelected().ID; got != ttid2 {
		t.Fatalf("boardSelected().ID = %q, want %q (clamp at start)", got, ttid2)
	}

	mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyRight})

	mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyRight})

	if got := mdl.(model).boardSelected().ID; got != ttid1_1 {
		t.Fatalf("boardSelected().ID = %q, want %q", got, ttid1_1)
	}

	mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if got := mdl.(model).boardSelected().ID; got != ttid1_1 {
		t.Fatalf("boardSelected().ID = %q, want %q (clamp at end)", got, ttid1_1)
	}
}

func TestUpdate_ModalPagingSingleTaskNoop(t *testing.T) {
	t.Parallel()

	var mdl tea.Model = New([]*task.Task{{ID: ttid1, Status: task.StatusDoing}})

	mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyLeft})

	got := mdl.(model)
	if got.boardSelected().ID != ttid1 {
		t.Errorf("boardSelected().ID = %q, want %q", got.boardSelected().ID, ttid1)
	}

	if !got.modalOpen {
		t.Errorf("modalOpen = false, want true")
	}

	mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyRight})

	got = mdl.(model)
	if got.boardSelected().ID != ttid1 {
		t.Errorf("boardSelected().ID = %q, want %q", got.boardSelected().ID, ttid1)
	}

	if !got.modalOpen {
		t.Errorf("modalOpen = false, want true")
	}
}

func TestUpdate_BoardEnterEmptyColumnNoop(t *testing.T) {
	t.Parallel()

	var mdl tea.Model = New(nil)

	mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if got := mdl.(model).modalOpen; got {
		t.Errorf("modalOpen = %v, want false", got)
	}
}

func TestGroupByStatus_UnapprovedTodosVisibleInListNotBoard(t *testing.T) {
	t.Parallel()

	m := New([]*task.Task{{ID: ttid1, Status: task.StatusTodo, Approved: false}})

	if got := len(m.Items()); got != 1 {
		t.Errorf("list Items() = %d, want 1 (unapproved todo still in list)", got)
	}

	if got := len(m.boardCols[0].Items()); got != 0 {
		t.Errorf("board todo column = %d items, want 0 (unapproved todo filtered from board)", got)
	}
}

func TestGroupByStatus_UnapprovedTodoIsHiddenFromBoard(t *testing.T) {
	t.Parallel()

	tasks := []*task.Task{{ID: ttid1, Status: task.StatusTodo, Approved: false}}

	cols := groupByStatus(tasks)

	if got := len(cols[0]); got != 0 {
		t.Errorf("todo column has %d tasks, want 0 (unapproved todo must be hidden)", got)
	}
}

func TestGroupByStatus_ApprovedTodoAppearsInBoard(t *testing.T) {
	t.Parallel()

	tasks := []*task.Task{{ID: ttid1, Status: task.StatusTodo, Approved: true}}

	cols := groupByStatus(tasks)

	if got := len(cols[0]); got != 1 {
		t.Errorf("todo column has %d tasks, want 1 (approved todo must appear)", got)
	}
}

func TestUpdate_BoardQuits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  tea.KeyPressMsg
	}{
		{"q", tea.KeyPressMsg{Code: 'q', Text: "q"}},
		{keyEsc, tea.KeyPressMsg{Code: tea.KeyEsc}},
		{"ctrl+c", tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var mdl tea.Model = New([]*task.Task{{ID: ttid1, Status: task.StatusTodo}})

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
