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

	t.Run("forwards navigation to list", func(t *testing.T) {
		t.Parallel()

		tasks := []*task.Task{
			{ID: ttid2},
			{ID: ttid2_1},
			{ID: ttid1},
		}

		m := New(tasks)

		updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})

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
		focused, _ := sized.(model).Update(tea.KeyPressMsg{Code: tea.KeyRight})
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
		focused, _ := sized.(model).Update(tea.KeyPressMsg{Code: tea.KeyRight})
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

		moved, _ := sized.(model).Update(tea.KeyPressMsg{Code: tea.KeyRight})

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
		_, _ = m.Update(tea.KeyPressMsg{Code: ' '})

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

	t.Run("enter opens modal", func(t *testing.T) {
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
					ID: ttid1, Status: task.StatusTodo, Body: ttBody,
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
			check func(t *testing.T, mdl tea.Model, cmd tea.Cmd, approves int)
		}{
			{
				"ctrl+c quits",
				tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl},
				func(t *testing.T, _ tea.Model, cmd tea.Cmd, _ int) {
					if cmd == nil {
						t.Fatalf("cmd = nil, want quit")
					}

					if _, ok := cmd().(tea.QuitMsg); !ok {
						t.Errorf("cmd() = %T, want tea.QuitMsg", cmd())
					}
				},
			},
			{
				"space swallowed",
				tea.KeyPressMsg{Code: ' '},
				func(t *testing.T, mdl tea.Model, _ tea.Cmd, approves int) {
					if approves != 0 {
						t.Errorf("approve called %d times, want 0", approves)
					}

					if !mdl.(model).modalOpen {
						t.Errorf("modalOpen = false, want true")
					}
				},
			},
			{
				"right pages to next row",
				tea.KeyPressMsg{Code: tea.KeyRight},
				func(t *testing.T, mdl tea.Model, _ tea.Cmd, _ int) {
					if got := mdl.(model).selected().ID; got != ttid2 {
						t.Errorf("selected().ID = %q, want %q", got, ttid2)
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

				approves := 0

				var mdl tea.Model = New([]*task.Task{
					{ID: ttid1, Status: task.StatusTodo, Body: ttBody},
					{ID: ttid2, Status: task.StatusDoing, Body: ttBody},
					{ID: ttid3, Status: task.StatusDone, Body: ttBody},
				}).WithApprove(func(string, bool) error {
					approves++

					return nil
				})

				mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
				mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

				mdl, cmd := mdl.Update(tt.key)

				tt.check(t, mdl, cmd, approves)
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

	t.Run("modal pages list rows", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name   string
			key    tea.KeyPressMsg
			wantID string
		}{
			{"right", tea.KeyPressMsg{Code: tea.KeyRight}, ttid1_1},
			{"l", tea.KeyPressMsg{Code: 'l', Text: "l"}, ttid1_1},
			{"left", tea.KeyPressMsg{Code: tea.KeyLeft}, ttid2},
			{"h", tea.KeyPressMsg{Code: 'h', Text: "h"}, ttid2},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var mdl tea.Model = New([]*task.Task{
					{ID: ttid2, Status: task.StatusTodo},
					{ID: ttid1, Status: task.StatusDoing},
					{ID: ttid1_1, Status: task.StatusDoing},
				})

				mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
				mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyDown})
				mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

				mdl, _ = mdl.Update(tt.key)

				opened := mdl.(model)
				if got := opened.selected().ID; got != tt.wantID {
					t.Errorf("selected().ID = %q, want %q", got, tt.wantID)
				}

				if !opened.modalOpen {
					t.Errorf("modalOpen = false, want true")
				}
			})
		}
	})

	t.Run("modal paging clamps at ends", func(t *testing.T) {
		t.Parallel()

		var mdl tea.Model = New([]*task.Task{
			{ID: ttid2, Status: task.StatusTodo},
			{ID: ttid1, Status: task.StatusDoing},
			{ID: ttid1_1, Status: task.StatusDoing},
		})

		mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
		if got := mdl.(model).selected().ID; got != ttid2 {
			t.Fatalf("selected().ID = %q, want %q (clamp at start)", got, ttid2)
		}

		for range 3 {
			mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyRight})
		}

		if got := mdl.(model).selected().ID; got != ttid1_1 {
			t.Fatalf("selected().ID = %q, want %q (clamp at end)", got, ttid1_1)
		}
	})

	t.Run("modal paging single task noop", func(t *testing.T) {
		t.Parallel()

		var mdl tea.Model = New([]*task.Task{{ID: ttid1, Status: task.StatusDoing}})

		mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		for _, key := range []tea.KeyPressMsg{{Code: tea.KeyLeft}, {Code: tea.KeyRight}} {
			mdl, _ = mdl.Update(key)

			got := mdl.(model)
			if got.selected().ID != ttid1 {
				t.Errorf("selected().ID = %q, want %q", got.selected().ID, ttid1)
			}

			if !got.modalOpen {
				t.Errorf("modalOpen = false, want true")
			}
		}
	})

	t.Run("enter on empty list noop", func(t *testing.T) {
		t.Parallel()

		var mdl tea.Model = New(nil)

		mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		if got := mdl.(model).modalOpen; got {
			t.Errorf("modalOpen = %v, want false", got)
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

		view := mdl.(model).View().Content
		if strings.Contains(view, "3 items") {
			t.Errorf("View() = %q, still renders the list item-count status bar", view)
		}
	})

	t.Run("empty list single empty state", func(t *testing.T) {
		t.Parallel()

		var mdl tea.Model = New(nil)

		mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 60, Height: 16})

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

		view := updated.(model).View().Content
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
			open    bool
			want    string
			notWant string
		}{
			{"collapsed", false, ttFocus, "page"},
			{"modal open", true, "page", ttFocus},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var mdl tea.Model = New([]*task.Task{{ID: ttid2}, {ID: ttid1}})

				mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

				if tt.open {
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

	t.Run("modal follows paged row", func(t *testing.T) {
		t.Parallel()

		var mdl tea.Model = New([]*task.Task{
			{ID: ttid1, Title: "One", Body: "ONETOKEN"},
			{ID: ttid2, Title: "Two", Body: "TWOTOKEN"},
		})

		mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyRight})

		if view := mdl.(model).View().Content; !strings.Contains(view, "BIT-2 — Two") {
			t.Errorf("View() = %q, missing modal title %q", view, "BIT-2 — Two")
		}

		mdl, _ = mdl.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})

		closed := mdl.(model)
		if view := closed.View().Content; !strings.Contains(view, "TWOTOKEN") {
			t.Errorf("View() = %q, missing detail body %q", view, "TWOTOKEN")
		}

		if closed.modalOpen {
			t.Errorf("modalOpen = true, want false")
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

	t.Run("modal shows body", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			enter bool
			want  bool
		}{
			{"closed hides modal title", false, false},
			{"open shows modal title", true, true},
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

				if got := strings.Contains(mdl.(model).View().Content, "BIT-1 — T"); got != tt.want {
					t.Errorf("View contains modal title = %v, want %v", got, tt.want)
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
