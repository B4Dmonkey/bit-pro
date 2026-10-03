package tui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/B4Dmonkey/bit-pro/task"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("preserves store order", func(t *testing.T) {
		t.Parallel()

		tasks := []*task.Task{
			{ID: ttid2},
			{ID: ttid2_1},
			{ID: ttid1},
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
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()

		m := New(nil)

		if got := len(m.Items()); got != 0 {
			t.Errorf("New(nil) produced %d items, want 0", got)
		}
	})

	t.Run("list help disabled", func(t *testing.T) {
		t.Parallel()

		m := New([]*task.Task{{ID: ttid1}})

		if m.ShowHelp() {
			t.Error("New() left the list's built-in help on, want it disabled")
		}
	})

	t.Run("defaults to board mode", func(t *testing.T) {
		t.Parallel()

		m := New([]*task.Task{{ID: ttid1}})

		if m.mode != modeBoard {
			t.Errorf("New() mode = %v, want modeBoard", m.mode)
		}
	})

	t.Run("lands on doings top bar", func(t *testing.T) {
		t.Parallel()

		tasks := []*task.Task{
			{ID: ttid2, Status: task.StatusTodo},
			{ID: ttid1, Status: task.StatusDoing},
			{ID: "BIT-1.1", Status: task.StatusDoing},
		}

		m := New(tasks)

		if m.activeCol != 1 {
			t.Errorf("New() activeCol = %d, want 1 (Doing)", m.activeCol)
		}

		if got := m.boardSelected(); got == nil || got.ID != "BIT-1.1" {
			t.Errorf("New() boardSelected() = %v, want BIT-1.1", got)
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	t.Run("reloaded msg rebuilds list", func(t *testing.T) {
		t.Parallel()

		m := New([]*task.Task{{ID: ttid1}})

		updated, _ := m.Update(reloadedMsg{tasks: []*task.Task{{ID: ttid1}, {ID: ttid2}}})

		items := updated.(model).Items()
		if len(items) != 2 {
			t.Fatalf("after reloadedMsg, len(Items()) = %d, want 2", len(items))
		}

		if got := items[0].(item).t.ID; got != ttid1 {
			t.Errorf("items[0].ID = %q, want %q", got, ttid1)
		}

		if got := items[1].(item).t.ID; got != ttid2 {
			t.Errorf("items[1].ID = %q, want %q", got, ttid2)
		}
	})

	t.Run("reloaded msg rebuilds board", func(t *testing.T) {
		t.Parallel()

		var mdl tea.Model = New([]*task.Task{{ID: ttid1, Status: task.StatusTodo, Approved: true}})

		mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

		mdl, _ = mdl.Update(reloadedMsg{tasks: []*task.Task{
			{ID: ttid1, Status: task.StatusTodo, Approved: true},
			{ID: ttid2, Status: task.StatusTodo, Approved: true},
		}})

		if got := len(mdl.(model).boardCols[0].Items()); got != 2 {
			t.Fatalf("after reloadedMsg, To Do column has %d items, want 2", got)
		}
	})

	t.Run("tick triggers reload", func(t *testing.T) {
		t.Parallel()

		m := New(nil)
		m.reload = func() ([]*task.Task, error) { return []*task.Task{{ID: ttid9}}, nil }

		_, cmd := m.Update(tickMsg{})

		if cmd == nil {
			t.Fatal("tickMsg produced cmd = nil, want a reload cmd")
		}

		rm, ok := cmd().(reloadedMsg)
		if !ok {
			t.Fatalf("tickMsg cmd() = %T, want reloadedMsg", cmd())
		}

		if len(rm.tasks) != 1 || rm.tasks[0].ID != ttid9 {
			t.Errorf("reloadedMsg tasks = %v, want one task BIT-9", rm.tasks)
		}

		if rm.err != nil {
			t.Errorf("reloadedMsg err = %v, want nil", rm.err)
		}
	})

	t.Run("reloaded msg reschedules", func(t *testing.T) {
		t.Parallel()

		m := New(nil).WithReload(func() ([]*task.Task, error) { return nil, nil })

		_, cmd := m.Update(reloadedMsg{tasks: nil})

		if cmd == nil {
			t.Error("reloadedMsg produced cmd = nil, want the next poll cmd")
		}
	})

	t.Run("reload error holds view", func(t *testing.T) {
		t.Parallel()

		m := New(nil).WithReload(func() ([]*task.Task, error) { return nil, nil })

		good, _ := m.Update(reloadedMsg{tasks: []*task.Task{{ID: ttid1}, {ID: ttid2}}})

		updated, cmd := good.(model).Update(reloadedMsg{tasks: nil, err: errors.New("mid-write")})

		if got := len(updated.(model).Items()); got != 2 {
			t.Fatalf("after errored reloadedMsg, len(Items()) = %d, want 2", got)
		}

		if cmd == nil {
			t.Error("errored reloadedMsg produced cmd = nil, want the next poll cmd")
		}
	})

	t.Run("reload preserves list selection", func(t *testing.T) {
		t.Parallel()

		m := New([]*task.Task{{ID: ttid2}, {ID: ttid2_1}, {ID: ttid1}})
		m.Select(2)

		updated, _ := m.Update(reloadedMsg{tasks: []*task.Task{
			{ID: ttid3},
			{ID: ttid2},
			{ID: ttid2_1},
			{ID: ttid1},
		}})

		if got := updated.(model).selected().ID; got != ttid1 {
			t.Errorf("after reload, selected().ID = %q, want %q", got, ttid1)
		}
	})

	t.Run("reload selection gone clamps", func(t *testing.T) {
		t.Parallel()

		m := New([]*task.Task{{ID: ttid5}, {ID: ttid4}, {ID: ttid3}, {ID: ttid2}, {ID: ttid1}})
		sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
		m = sized.(model)
		m.Select(4)

		updated, _ := m.Update(reloadedMsg{tasks: []*task.Task{{ID: ttid5}, {ID: ttid4}}})
		got := updated.(model)

		if got.Index() != 1 {
			t.Errorf("after reload dropping the selected task, Index() = %d, want 1", got.Index())
		}

		sel := got.selected()
		if sel == nil {
			t.Fatalf("after reload dropping the selected task, selected() = nil, want a valid item")
		}

		if sel.ID != ttid4 {
			t.Errorf("after reload dropping the selected task, selected().ID = %q, want %q", sel.ID, ttid4)
		}
	})

	t.Run("reload preserves board selection", func(t *testing.T) {
		t.Parallel()

		var mdl tea.Model = New([]*task.Task{
			{ID: ttid1, Status: task.StatusTodo},
			{ID: ttid2, Status: task.StatusDoing},
			{ID: ttid3, Status: task.StatusDoing},
		})

		mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyDown})

		mdl, _ = mdl.Update(reloadedMsg{tasks: []*task.Task{
			{ID: ttid1, Status: task.StatusTodo},
			{ID: ttid2, Status: task.StatusDoing},
			{ID: ttid3, Status: task.StatusDoing},
			{ID: ttid4, Status: task.StatusTodo},
		}})
		got := mdl.(model)

		if got.activeCol != 1 {
			t.Errorf("after reload, activeCol = %d, want 1", got.activeCol)
		}

		if got.mode != modeBoard {
			t.Errorf("after reload, mode = %v, want modeBoard", got.mode)
		}

		sel := got.boardSelected()
		if sel == nil {
			t.Fatalf("after reload, boardSelected() = nil, want a valid card")
		}

		if sel.ID != ttid3 {
			t.Errorf("after reload, boardSelected().ID = %q, want %q", sel.ID, ttid3)
		}
	})

	t.Run("forwards navigation to list", func(t *testing.T) {
		t.Parallel()

		tasks := []*task.Task{
			{ID: ttid2},
			{ID: ttid2_1},
			{ID: ttid1},
		}

		m := New(tasks)

		toList, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		updated, _ := toList.(model).Update(tea.KeyPressMsg{Code: tea.KeyDown})

		if got := updated.(model).Index(); got != 1 {
			t.Errorf("after KeyDown, Index() = %d, want 1", got)
		}
	})

	t.Run("window size builds renderer", func(t *testing.T) {
		t.Parallel()

		m := New([]*task.Task{{ID: ttid1, Body: ttBodyHi}})

		updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

		if updated.(model).renderer == nil {
			t.Fatal("after WindowSizeMsg, renderer = nil, want it constructed in Update, not View")
		}
	})

	t.Run("question toggles full help", func(t *testing.T) {
		t.Parallel()

		body := strings.Repeat("line\n", 500)

		var mdl tea.Model = New([]*task.Task{{ID: ttid1, Body: body}})

		mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

		if mdl.(model).help.ShowAll {
			t.Fatal("help starts expanded, want collapsed")
		}

		q := tea.KeyPressMsg{Code: '?', Text: "?"}
		mdl, _ = mdl.Update(q)

		if !mdl.(model).help.ShowAll {
			t.Error("after ?, help.ShowAll = false, want true (full menu)")
		}

		if h := lipgloss.Height(mdl.(model).View().Content); h > 24 {
			t.Fatalf("expanded help View height = %d, want <= 24", h)
		}

		mdl, _ = mdl.Update(q)
		if mdl.(model).help.ShowAll {
			t.Error("after second ?, help.ShowAll = true, want false (collapsed)")
		}
	})

	t.Run("window size sizes viewport", func(t *testing.T) {
		t.Parallel()

		m := New([]*task.Task{{ID: ttid1, Body: ttBodyHi}})

		updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

		if h := updated.(model).viewport.Height(); h != 21 {
			t.Fatalf("viewport.Height = %d, want 21", h)
		}
	})

	t.Run("ctrl d scrolls detail", func(t *testing.T) {
		t.Parallel()

		body := strings.Repeat("line\n", 500)
		m := New([]*task.Task{{ID: ttid1, Body: body}})

		sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		toList, _ := sized.(model).Update(tea.KeyPressMsg{Code: tea.KeyTab})
		focused, _ := toList.(model).Update(tea.KeyPressMsg{Code: tea.KeyRight})
		scrolled, _ := focused.(model).Update(tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl})

		got := scrolled.(model)
		if off := got.viewport.YOffset(); off == 0 {
			t.Fatal("after ctrl+d, viewport.YOffset = 0, want > 0")
		}
	})

	t.Run("navigation resets detail scroll", func(t *testing.T) {
		t.Parallel()

		body := strings.Repeat("line\n", 500)
		m := New([]*task.Task{
			{ID: ttid2, Body: body},
			{ID: ttid1, Body: body},
		})

		sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		toList, _ := sized.(model).Update(tea.KeyPressMsg{Code: tea.KeyTab})
		focused, _ := toList.(model).Update(tea.KeyPressMsg{Code: tea.KeyRight})
		scrolled, _ := focused.(model).Update(tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl})

		scrolledModel := scrolled.(model)
		if scrolledModel.viewport.YOffset() == 0 {
			t.Fatal("setup: ctrl+d did not scroll the detail")
		}

		listFocused, _ := scrolled.(model).Update(tea.KeyPressMsg{Code: tea.KeyLeft})
		moved, _ := listFocused.(model).Update(tea.KeyPressMsg{Code: tea.KeyDown})

		movedModel := moved.(model)
		if off := movedModel.viewport.YOffset(); off != 0 {
			t.Fatalf("after changing selection, viewport.YOffset = %d, want 0", off)
		}
	})

	t.Run("focus", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name        string
			keys        []rune
			wantFocused bool
		}{
			{"default is list", nil, false},
			{"right focuses detail", []rune{tea.KeyRight}, true},
			{"right then left returns to list", []rune{tea.KeyRight, tea.KeyLeft}, false},
			{"left on list clamps to list", []rune{tea.KeyLeft}, false},
			{"right twice clamps to detail", []rune{tea.KeyRight, tea.KeyRight}, true},
			{"h focuses list like left", []rune{tea.KeyRight, 'h'}, false},
			{"l focuses detail like right", []rune{'l'}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var mdl tea.Model = New([]*task.Task{{ID: ttid1, Body: ttBodyHi}})

				mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

				mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
				for _, k := range tt.keys {
					mdl, _ = mdl.Update(tea.KeyPressMsg{Code: k})
				}

				if got := mdl.(model).detailFocused; got != tt.wantFocused {
					t.Errorf("detailFocused = %v, want %v", got, tt.wantFocused)
				}
			})
		}
	})

	t.Run("right does not page list", func(t *testing.T) {
		t.Parallel()

		m := New([]*task.Task{{ID: ttid1}, {ID: ttid2}, {ID: ttid3}})
		sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		toList, _ := sized.(model).Update(tea.KeyPressMsg{Code: tea.KeyTab})

		moved, _ := toList.(model).Update(tea.KeyPressMsg{Code: tea.KeyRight})

		got := moved.(model)
		if idx := got.Index(); idx != 0 {
			t.Errorf("after KeyRight, Index() = %d, want 0", idx)
		}

		if page := got.Paginator.Page; page != 0 {
			t.Errorf("after KeyRight, Paginator.Page = %d, want 0", page)
		}
	})

	t.Run("focus routes arrows", func(t *testing.T) {
		t.Parallel()

		body := strings.Repeat("line\n", 500)
		tests := []struct {
			name         string
			keys         []rune
			wantIndex    int
			wantScrolled bool
		}{
			{"list focused: down moves selection, detail still", []rune{tea.KeyDown}, 1, false},
			{"detail focused: down scrolls body, list still", []rune{tea.KeyRight, tea.KeyDown}, 0, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var mdl tea.Model = New([]*task.Task{
					{ID: ttid2, Body: body},
					{ID: ttid1, Body: body},
				})

				mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

				mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
				for _, k := range tt.keys {
					mdl, _ = mdl.Update(tea.KeyPressMsg{Code: k})
				}

				got := mdl.(model)
				if got.Index() != tt.wantIndex {
					t.Errorf("Index() = %d, want %d", got.Index(), tt.wantIndex)
				}

				if scrolled := got.viewport.YOffset() > 0; scrolled != tt.wantScrolled {
					t.Errorf("viewport scrolled = %v (YOffset=%d), want %v", scrolled, got.viewport.YOffset(), tt.wantScrolled)
				}
			})
		}
	})

	t.Run("tab toggles mode", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			presses int
			want    viewMode
		}{
			{"default is board", 0, modeBoard},
			{"one tab to list", 1, modeList},
			{"two tabs back to board", 2, modeBoard},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var mdl tea.Model = New([]*task.Task{{ID: ttid1}})
				for range tt.presses {
					mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
				}

				if got := mdl.(model).mode; got != tt.want {
					t.Errorf("after %d tab(s), mode = %v, want %v", tt.presses, got, tt.want)
				}
			})
		}
	})

	t.Run("enter expands detail", func(t *testing.T) {
		t.Parallel()

		var mdl tea.Model = New([]*task.Task{{ID: ttid1}})

		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		if got := mdl.(model).detailExpanded; !got {
			t.Errorf("detailExpanded = %v, want true after Enter in list mode", got)
		}
	})

	t.Run("enter toggles detail back and forth", func(t *testing.T) {
		t.Parallel()

		var mdl tea.Model = New([]*task.Task{{ID: ttid1}})

		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		if got := mdl.(model).detailExpanded; got {
			t.Errorf("detailExpanded = %v, want false after a second Enter", got)
		}
	})

	t.Run("enter focuses detail for scrolling", func(t *testing.T) {
		t.Parallel()

		body := strings.Repeat("line\n", 500)

		var mdl tea.Model = New([]*task.Task{
			{ID: ttid2, Body: body},
			{ID: ttid1, Body: body},
		})

		mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyDown})

		got := mdl.(model)
		if off := got.viewport.YOffset(); off == 0 {
			t.Errorf("viewport.YOffset = %d, want > 0 — down should scroll the expanded body", off)
		}

		if idx := got.Index(); idx != 0 {
			t.Errorf("Index() = %d, want 0 — down should not move the list selection while expanded", idx)
		}
	})

	t.Run("enter again returns focus to list", func(t *testing.T) {
		t.Parallel()

		body := strings.Repeat("line\n", 500)

		var mdl tea.Model = New([]*task.Task{
			{ID: ttid2, Body: body},
			{ID: ttid1, Body: body},
		})

		mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyDown})

		got := mdl.(model)
		if idx := got.Index(); idx != 1 {
			t.Errorf("Index() = %d, want 1 — down should move the list selection once collapsed", idx)
		}

		if off := got.viewport.YOffset(); off != 0 {
			t.Errorf("viewport.YOffset = %d, want 0 — down should not scroll the body once collapsed", off)
		}
	})

	t.Run("pages list when expanded", func(t *testing.T) {
		t.Parallel()

		var mdl tea.Model = New([]*task.Task{{ID: ttid2}, {ID: ttid1}})

		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyRight})

		if got := mdl.(model).Index(); got != 1 {
			t.Errorf("Index() = %d, want 1 after right while expanded", got)
		}
	})

	t.Run("pages list left when expanded", func(t *testing.T) {
		t.Parallel()

		m := New([]*task.Task{{ID: ttid3}, {ID: ttid2}, {ID: ttid1}})
		toList, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		expanded, _ := toList.(model).Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		em := expanded.(model)
		em.Select(2)

		mdl, _ := em.Update(tea.KeyPressMsg{Code: tea.KeyLeft})

		if got := mdl.(model).Index(); got != 1 {
			t.Errorf("Index() = %d, want 1 after left while expanded", got)
		}
	})

	t.Run("paging clamps at list ends", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name      string
			startIdx  int
			key       tea.KeyPressMsg
			wantIndex int
		}{
			{"left at first item stays", 0, tea.KeyPressMsg{Code: tea.KeyLeft}, 0},
			{"right at last item stays", 1, tea.KeyPressMsg{Code: tea.KeyRight}, 1},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				m := New([]*task.Task{{ID: ttid2}, {ID: ttid1}})
				toList, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
				expanded, _ := toList.(model).Update(tea.KeyPressMsg{Code: tea.KeyEnter})
				em := expanded.(model)
				em.Select(tt.startIdx)

				mdl, _ := em.Update(tt.key)

				if got := mdl.(model).Index(); got != tt.wantIndex {
					t.Errorf("Index() = %d, want %d", got, tt.wantIndex)
				}
			})
		}
	})

	t.Run("enter relayouts pane widths", func(t *testing.T) {
		t.Parallel()

		var mdl tea.Model = New([]*task.Task{{ID: ttid1}})

		mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})

		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		if lw := mdl.(model).listWidth; lw != 10 {
			t.Errorf("listWidth = %d, want 10 (splitWidthExpanded) after Enter alone, without another resize", lw)
		}
	})

	t.Run("quits from detail", func(t *testing.T) {
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

				var mdl tea.Model = New([]*task.Task{{ID: ttid1}})

				mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
				mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
				mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyRight})

				_, cmd := mdl.Update(tt.key)

				if cmd == nil {
					t.Fatalf("%s from detail pane: cmd = nil, want a quit cmd", tt.name)
				}

				if _, ok := cmd().(tea.QuitMsg); !ok {
					t.Errorf("%s from detail pane: cmd() = %T, want tea.QuitMsg", tt.name, cmd())
				}
			})
		}
	})

	t.Run("esc quits from list", func(t *testing.T) {
		t.Parallel()

		m := New([]*task.Task{{ID: ttid1}})

		_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})

		if cmd == nil {
			t.Fatal("after KeyEsc in list, cmd = nil, want a quit cmd")
		}

		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Errorf("after KeyEsc in list, cmd() = %T, want tea.QuitMsg", cmd())
		}
	})

	t.Run("space toggles approval in list mode", func(t *testing.T) {
		t.Parallel()

		var called []struct {
			id       string
			approved bool
		}

		m := New([]*task.Task{{ID: ttid1, Status: task.StatusTodo, Approved: false}}).
			WithApprove(func(id string, a bool) error {
				called = append(called, struct {
					id       string
					approved bool
				}{id, a})

				return nil
			})
		mdl, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		_, _ = mdl.(model).Update(tea.KeyPressMsg{Code: ' '})

		if len(called) != 1 {
			t.Fatalf("approve called %d times, want 1", len(called))
		}

		if called[0].id != ttid1 {
			t.Errorf("approve id = %q, want %q", called[0].id, ttid1)
		}

		if !called[0].approved {
			t.Errorf("approve approved = false, want true (invert of Approved:false)")
		}
	})

	t.Run("space toggles approval in board mode", func(t *testing.T) {
		t.Parallel()

		var called []struct {
			id       string
			approved bool
		}

		m := New([]*task.Task{{ID: ttid1, Status: task.StatusDoing, Approved: false}}).
			WithApprove(func(id string, a bool) error {
				called = append(called, struct {
					id       string
					approved bool
				}{id, a})

				return nil
			})
		_, _ = m.Update(tea.KeyPressMsg{Code: ' '})

		if len(called) != 1 {
			t.Fatalf("approve called %d times, want 1", len(called))
		}

		if called[0].id != ttid1 {
			t.Errorf("approve id = %q, want %q", called[0].id, ttid1)
		}
	})

	t.Run("space on approved item sends unapproved", func(t *testing.T) {
		t.Parallel()

		var called []struct {
			id       string
			approved bool
		}

		m := New([]*task.Task{{ID: ttid1, Status: task.StatusTodo, Approved: true}}).
			WithApprove(func(id string, a bool) error {
				called = append(called, struct {
					id       string
					approved bool
				}{id, a})

				return nil
			})
		_, _ = m.Update(tea.KeyPressMsg{Code: ' '})

		if len(called) != 1 {
			t.Fatalf("approve called %d times, want 1", len(called))
		}

		if called[0].approved {
			t.Errorf("approve approved = true, want false (invert of Approved:true)")
		}
	})

	t.Run("space with no callback is noop", func(t *testing.T) {
		t.Parallel()

		tasks := []*task.Task{{ID: ttid1, Status: task.StatusTodo, Approved: false}}
		m := New(tasks)

		updated, _ := m.Update(tea.KeyPressMsg{Code: ' '})

		got := updated.(model)
		if got.selected() == nil {
			t.Fatal("selected() = nil after space noop, want unchanged model")
		}

		if got.selected().ID != ttid1 {
			t.Errorf("selected().ID = %q after space noop, want %q", got.selected().ID, ttid1)
		}
	})

	t.Run("board active column", func(t *testing.T) {
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
	})

	t.Run("board card selection", func(t *testing.T) {
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
	})

	t.Run("board enter opens modal", func(t *testing.T) {
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
	})

	t.Run("modal closes", func(t *testing.T) {
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
	})

	t.Run("modal captures input", func(t *testing.T) {
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
	})

	t.Run("modal scrolls long body", func(t *testing.T) {
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
	})

	t.Run("modal pages within column", func(t *testing.T) {
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
	})

	t.Run("modal pages across columns", func(t *testing.T) {
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
	})

	t.Run("modal paging clamps at ends", func(t *testing.T) {
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
	})

	t.Run("modal paging single task noop", func(t *testing.T) {
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
	})

	t.Run("board enter empty column noop", func(t *testing.T) {
		t.Parallel()

		var mdl tea.Model = New(nil)

		mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		if got := mdl.(model).modalOpen; got {
			t.Errorf("modalOpen = %v, want false", got)
		}
	})

	t.Run("board quits", func(t *testing.T) {
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
	})
}

func TestInit(t *testing.T) {
	t.Parallel()

	t.Run("starts polling when reload set", func(t *testing.T) {
		t.Parallel()

		set := New(nil).WithReload(func() ([]*task.Task, error) { return nil, nil })
		none := New(nil)

		if set.Init() == nil {
			t.Error("Init() with reload set = nil, want a poll cmd")
		}

		if none.Init() != nil {
			t.Error("Init() with no reload = non-nil, want nil")
		}
	})
}

func TestSameTasks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    []*task.Task
		b    []*task.Task
		want bool
	}{
		{
			name: "identical single task",
			a:    []*task.Task{{ID: ttid1, Status: task.StatusTodo, Title: "one", Body: "b"}},
			b:    []*task.Task{{ID: ttid1, Status: task.StatusTodo, Title: "one", Body: "b"}},
			want: true,
		},
		{
			name: "different length",
			a:    []*task.Task{{ID: ttid1}},
			b:    []*task.Task{{ID: ttid1}, {ID: ttid2}},
			want: false,
		},
		{
			name: "same length different ID",
			a:    []*task.Task{{ID: ttid1}},
			b:    []*task.Task{{ID: ttid2}},
			want: false,
		},
		{
			name: "same ID different Status",
			a:    []*task.Task{{ID: ttid1, Status: task.StatusTodo}},
			b:    []*task.Task{{ID: ttid1, Status: task.StatusDoing}},
			want: false,
		},
		{
			name: "same ID different Body",
			a:    []*task.Task{{ID: ttid1, Body: "before"}},
			b:    []*task.Task{{ID: ttid1, Body: "after"}},
			want: false,
		},
		{
			name: "two tasks reordered",
			a:    []*task.Task{{ID: ttid1}, {ID: ttid2}},
			b:    []*task.Task{{ID: ttid2}, {ID: ttid1}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := sameTasks(tt.a, tt.b); got != tt.want {
				t.Errorf("sameTasks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelected(t *testing.T) {
	t.Parallel()

	t.Run("tracks cursor", func(t *testing.T) {
		t.Parallel()

		tasks := []*task.Task{
			{ID: ttid2},
			{ID: ttid2_1},
			{ID: ttid1},
		}

		m := New(tasks)

		if got := m.selected().ID; got != tasks[0].ID {
			t.Errorf("default selected().ID = %q, want %q", got, tasks[0].ID)
		}

		m.Select(2)

		if got := m.selected().ID; got != tasks[2].ID {
			t.Errorf("after Select(2), selected().ID = %q, want %q", got, tasks[2].ID)
		}
	})

	t.Run("empty list nil", func(t *testing.T) {
		t.Parallel()

		m := New(nil)

		if got := m.selected(); got != nil {
			t.Errorf("selected() on empty list = %v, want nil", got)
		}
	})
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

func TestSplitWidthExpanded(t *testing.T) {
	t.Parallel()

	t.Run("typical terminal", func(t *testing.T) {
		t.Parallel()

		listW, detailW := splitWidthExpanded(100)

		if listW != 10 {
			t.Errorf("splitWidthExpanded(100) listW = %d, want 10", listW)
		}

		if detailW != 89 {
			t.Errorf("splitWidthExpanded(100) detailW = %d, want 89", detailW)
		}
	})

	t.Run("scales with width", func(t *testing.T) {
		t.Parallel()

		listW, detailW := splitWidthExpanded(200)

		if listW != 20 {
			t.Errorf("splitWidthExpanded(200) listW = %d, want 20", listW)
		}

		if detailW != 179 {
			t.Errorf("splitWidthExpanded(200) detailW = %d, want 179", detailW)
		}
	})

	t.Run("detail wider than list", func(t *testing.T) {
		t.Parallel()

		listW, detailW := splitWidthExpanded(120)

		if detailW <= listW {
			t.Errorf("splitWidthExpanded(120) detailW = %d, listW = %d, want detailW > listW", detailW, listW)
		}
	})

	t.Run("zero and one width", func(t *testing.T) {
		t.Parallel()

		for _, total := range []int{0, 1} {
			listW, detailW := splitWidthExpanded(total)

			if listW < 0 {
				t.Errorf("splitWidthExpanded(%d) listW = %d, want >= 0", total, listW)
			}

			if detailW < 0 {
				t.Errorf("splitWidthExpanded(%d) detailW = %d, want >= 0", total, detailW)
			}

			if listW+detailW > total {
				t.Errorf("splitWidthExpanded(%d) listW+detailW = %d, want <= %d", total, listW+detailW, total)
			}
		}
	})
}

func TestLayout(t *testing.T) {
	t.Parallel()

	t.Run("expanded uses wider split", func(t *testing.T) {
		t.Parallel()

		m := New([]*task.Task{{ID: ttid1}})
		m.detailExpanded = true

		updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})

		if lw := updated.(model).listWidth; lw != 10 {
			t.Errorf("listWidth = %d, want 10 (splitWidthExpanded, not splitWidth)", lw)
		}
	})
}

func TestView(t *testing.T) {
	t.Parallel()

	t.Run("fits window height", func(t *testing.T) {
		t.Parallel()

		body := strings.Repeat("line\n", 500)
		m := New([]*task.Task{{ID: ttid1, Body: body}})

		updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

		if h := lipgloss.Height(updated.(model).View().Content); h > 24 {
			t.Fatalf("View height = %d, want <= 24 (detail must not overflow the screen)", h)
		}
	})

	t.Run("pane titles", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			total int
			done  int
			want  string
		}{
			{"none done", 7, 0, "Tasks (0/7)"},
			{"some done", 7, 3, "Tasks (3/7)"},
			{"all done", 3, 3, "Tasks (3/3)"},
			{"empty", 0, 0, "Tasks (0/0)"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				tasks := make([]*task.Task, tt.total)
				for i := range tasks {
					tasks[i] = &task.Task{ID: ttid1}
					if i < tt.done {
						tasks[i].Status = "done"
					}
				}

				var mdl tea.Model = New(tasks)

				mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
				mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})

				view := mdl.(model).View().Content
				if !strings.Contains(view, tt.want) {
					t.Errorf("View() missing %q", tt.want)
				}

				if !strings.Contains(view, "Details") {
					t.Errorf("View() missing \"Details\"")
				}
			})
		}
	})

	t.Run("list hides title heading", func(t *testing.T) {
		t.Parallel()

		var mdl tea.Model = New(nil)

		mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 60, Height: 16})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})

		view := mdl.(model).View().Content
		if strings.Contains(view, "List") {
			t.Errorf("View() = %q, still renders the List title heading", view)
		}
	})

	t.Run("list hides item count", func(t *testing.T) {
		t.Parallel()

		tasks := make([]*task.Task, 3)
		for i := range tasks {
			tasks[i] = &task.Task{ID: ttid1}
		}

		var mdl tea.Model = New(tasks)

		mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 60, Height: 16})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})

		view := mdl.(model).View().Content
		if strings.Contains(view, "3 items") {
			t.Errorf("View() = %q, still renders the list item-count status bar", view)
		}
	})

	t.Run("empty list single empty state", func(t *testing.T) {
		t.Parallel()

		var mdl tea.Model = New(nil)

		mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 60, Height: 16})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})

		view := mdl.(model).View().Content
		if got := strings.Count(view, "No items"); got != 1 {
			t.Errorf("View() = %q, %d %q lines, want exactly 1", view, got, "No items")
		}
	})

	t.Run("help bar present and bounded", func(t *testing.T) {
		t.Parallel()

		body := strings.Repeat("line\n", 500)
		m := New([]*task.Task{{ID: ttid1, Body: body}})

		updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		withTab, _ := updated.(model).Update(tea.KeyPressMsg{Code: tea.KeyTab})

		view := withTab.(model).View().Content
		if !strings.Contains(view, ttFocus) {
			t.Errorf("View() missing help text %q", ttFocus)
		}

		if h := lipgloss.Height(view); h > 24 {
			t.Fatalf("View height = %d, want <= 24 (help bar must fit the budget)", h)
		}
	})

	t.Run("footer labels arrows for current state", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			expand  bool
			want    string
			notWant string
		}{
			{"collapsed", false, ttFocus, "page"},
			{"expanded", true, "page", ttFocus},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var mdl tea.Model = New([]*task.Task{{ID: ttid2}, {ID: ttid1}})

				mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

				mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
				if tt.expand {
					mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
				}

				view := mdl.(model).View().Content
				if !strings.Contains(view, tt.want) {
					t.Errorf("View() missing help text %q", tt.want)
				}

				if strings.Contains(view, tt.notWant) {
					t.Errorf("View() contains help text %q, want it absent", tt.notWant)
				}
			})
		}
	})

	t.Run("modal title inverted", func(t *testing.T) {
		t.Parallel()

		m := New([]*task.Task{{ID: ttid1, Status: task.StatusTodo, Approved: true, Title: "T", Body: "b"}})
		mdl, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		mdl, _ = mdl.(model).Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		view := mdl.(model).View().Content
		if !strings.Contains(view, "\x1b[7m BIT-1 — T \x1b[27m") {
			t.Errorf("modal view = %q, want reverse-video title span \\x1b[7m BIT-1 — T \\x1b[27m", view)
		}

		if !strings.Contains(view, "\x1b[32m") {
			t.Errorf("modal view = %q, want green border SGR \\x1b[32m", view)
		}
	})

	t.Run("board column counts", func(t *testing.T) {
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
	})

	t.Run("board help", func(t *testing.T) {
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
	})

	t.Run("modal shows body", func(t *testing.T) {
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
	})
}

func TestIsBar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"track", ttid2, false},
		{"bar", "BIT-2.5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isBar(tt.id); got != tt.want {
				t.Errorf("isBar(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestVerse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		task *task.Task
		want string
	}{
		{"phased bar", &task.Task{ID: ttid2_1, Phase: 2, PhaseLabel: "List & read"}, "phase 2 — List & read"},
		{"unphased bar", &task.Task{ID: ttid2_1, Phase: 0}, ""},
		{"track", &task.Task{ID: ttid2, Phase: 2, PhaseLabel: "List & read"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := verse(tt.task); got != tt.want {
				t.Errorf("verse(%+v) = %q, want %q", tt.task, got, tt.want)
			}
		})
	}
}

func TestTitledBorder(t *testing.T) {
	t.Parallel()

	t.Run("active uses terminal green", func(t *testing.T) {
		t.Parallel()

		got := titledBorder(ttBody, "Tasks (0)", 20, 3, true)

		if !strings.Contains(got, "\x1b[32m") {
			t.Errorf("titledBorder active = %q, want terminal green SGR \\x1b[32m", got)
		}

		if strings.Contains(got, "38;5;99") {
			t.Errorf("titledBorder active = %q, still contains 256-purple 38;5;99", got)
		}
	})

	t.Run("active title inverted", func(t *testing.T) {
		t.Parallel()

		got := titledBorder(ttBody, "Tasks (0)", 20, 3, true)

		if !strings.Contains(got, "\x1b[7;32m") {
			t.Errorf("titledBorder active = %q, want reverse-green title SGR \\x1b[7;32m", got)
		}

		if !strings.Contains(got, "\x1b[32m") {
			t.Errorf("titledBorder active = %q, want green border SGR \\x1b[32m", got)
		}
	})

	t.Run("inactive title framed", func(t *testing.T) {
		t.Parallel()

		got := titledBorder(ttBody, "Doing (0)", 20, 3, false)

		if !strings.Contains(got, "| Doing (0) |") {
			t.Errorf("titledBorder inactive = %q, want framed title | Doing (0) |", got)
		}
	})

	t.Run("active title not framed", func(t *testing.T) {
		t.Parallel()

		got := titledBorder(ttBody, "Tasks (0)", 20, 3, true)

		if strings.Contains(got, "| Tasks (0) |") {
			t.Errorf("titledBorder active = %q, should not frame title with pipes", got)
		}
	})
}
