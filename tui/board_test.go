package tui

import (
	"slices"
	"testing"

	"charm.land/bubbles/v2/list"
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

	t.Run("preserves order within column", func(t *testing.T) {
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
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		cols := groupByStatus(nil)

		for i, col := range cols {
			if len(col) != 0 {
				t.Errorf("column %d has %d tasks, want 0", i, len(col))
			}
		}
	})

	t.Run("unapproved todos visible in list not board", func(t *testing.T) {
		t.Parallel()

		m := New([]*task.Task{{ID: ttid1, Status: task.StatusTodo, Approved: false}})

		if got := len(m.Items()); got != 1 {
			t.Errorf("list Items() = %d, want 1 (unapproved todo still in list)", got)
		}

		if got := len(m.boardCols[0].Items()); got != 0 {
			t.Errorf("board todo column = %d items, want 0 (unapproved todo filtered from board)", got)
		}
	})

	t.Run("unapproved todo is hidden from board", func(t *testing.T) {
		t.Parallel()

		tasks := []*task.Task{{ID: ttid1, Status: task.StatusTodo, Approved: false}}

		cols := groupByStatus(tasks)

		if got := len(cols[0]); got != 0 {
			t.Errorf("todo column has %d tasks, want 0 (unapproved todo must be hidden)", got)
		}
	})

	t.Run("approved todo appears in board", func(t *testing.T) {
		t.Parallel()

		tasks := []*task.Task{{ID: ttid1, Status: task.StatusTodo, Approved: true}}

		cols := groupByStatus(tasks)

		if got := len(cols[0]); got != 1 {
			t.Errorf("todo column has %d tasks, want 1 (approved todo must appear)", got)
		}
	})
}

func TestBoardColumns(t *testing.T) {
	t.Parallel()

	t.Run("from grouping", func(t *testing.T) {
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
	})
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
			cols: [3]list.Model{{}, newColumnList([]*task.Task{{ID: ttid1}, {ID: ttid1_1}}), {}},
			want: []boardEntry{
				{col: 1, pos: 0, t: &task.Task{ID: ttid1}},
				{col: 1, pos: 1, t: &task.Task{ID: ttid1_1}},
			},
		},
		{
			name: "all three columns concatenate in order",
			cols: [3]list.Model{
				newColumnList([]*task.Task{{ID: ttid2}}),
				newColumnList([]*task.Task{{ID: ttid1}, {ID: ttid1_1}}),
				newColumnList([]*task.Task{{ID: ttid3}}),
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
			cols: [3]list.Model{{}, newColumnList([]*task.Task{{ID: ttid1}}), {}},
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
