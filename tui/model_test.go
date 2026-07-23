package tui

import (
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})

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

func TestNew_ListHelpDisabled(t *testing.T) {
	t.Parallel()

	m := New([]*task.Task{{ID: "BIT-1"}})

	if m.ShowHelp() {
		t.Error("New() left the list's built-in help on, want it disabled")
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

func TestUpdate_WindowSizeBuildsRenderer(t *testing.T) {
	t.Parallel()

	m := New([]*task.Task{{ID: "BIT-1", Body: "# Hi"}})

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if updated.(model).renderer == nil {
		t.Fatal("after WindowSizeMsg, renderer = nil, want it constructed in Update, not View")
	}
}

func TestView_FitsWindowHeight(t *testing.T) {
	t.Parallel()

	body := strings.Repeat("line\n", 500)
	m := New([]*task.Task{{ID: "BIT-1", Body: body}})

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if h := lipgloss.Height(updated.(model).View().Content); h > 24 {
		t.Fatalf("View height = %d, want <= 24 (detail must not overflow the screen)", h)
	}
}

func TestView_PaneTitles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		count int
		want  string
	}{
		{"many tasks", 29, "Tasks (29)"},
		{"one task", 1, "Tasks (1)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tasks := make([]*task.Task, tt.count)
			for i := range tasks {
				tasks[i] = &task.Task{ID: "BIT-1"}
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
}

func TestView_HelpBarPresentAndBounded(t *testing.T) {
	t.Parallel()

	body := strings.Repeat("line\n", 500)
	m := New([]*task.Task{{ID: "BIT-1", Body: body}})

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := updated.(model).View().Content
	if !strings.Contains(view, "focus") {
		t.Errorf("View() missing help text %q", "focus")
	}
	if h := lipgloss.Height(view); h > 24 {
		t.Fatalf("View height = %d, want <= 24 (help bar must fit the budget)", h)
	}
}

func TestUpdate_QuestionTogglesFullHelp(t *testing.T) {
	t.Parallel()

	body := strings.Repeat("line\n", 500)
	var mdl tea.Model = New([]*task.Task{{ID: "BIT-1", Body: body}})
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
}

func TestUpdate_WindowSizeSizesViewport(t *testing.T) {
	t.Parallel()

	m := New([]*task.Task{{ID: "BIT-1", Body: "# Hi"}})

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if h := updated.(model).viewport.Height(); h != 21 {
		t.Fatalf("viewport.Height = %d, want 21", h)
	}
}

func TestUpdate_CtrlDScrollsDetail(t *testing.T) {
	t.Parallel()

	body := strings.Repeat("line\n", 500)
	m := New([]*task.Task{{ID: "BIT-1", Body: body}})

	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	focused, _ := sized.(model).Update(tea.KeyPressMsg{Code: tea.KeyRight})
	scrolled, _ := focused.(model).Update(tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl})

	got := scrolled.(model)
	if off := got.viewport.YOffset(); off == 0 {
		t.Fatal("after ctrl+d, viewport.YOffset = 0, want > 0")
	}
}

func TestUpdate_NavigationResetsDetailScroll(t *testing.T) {
	t.Parallel()

	body := strings.Repeat("line\n", 500)
	m := New([]*task.Task{
		{ID: "BIT-2", Body: body},
		{ID: "BIT-1", Body: body},
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
}

func TestUpdate_Focus(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var mdl tea.Model = New([]*task.Task{{ID: "BIT-1", Body: "# Hi"}})
			mdl, _ = mdl.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			for _, k := range tt.keys {
				mdl, _ = mdl.Update(tea.KeyPressMsg{Code: k})
			}

			if got := mdl.(model).detailFocused; got != tt.wantFocused {
				t.Errorf("detailFocused = %v, want %v", got, tt.wantFocused)
			}
		})
	}
}

func TestUpdate_RightDoesNotPageList(t *testing.T) {
	t.Parallel()

	m := New([]*task.Task{{ID: "BIT-1"}, {ID: "BIT-2"}, {ID: "BIT-3"}})
	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	moved, _ := sized.(model).Update(tea.KeyPressMsg{Code: tea.KeyRight})

	got := moved.(model)
	if idx := got.Index(); idx != 0 {
		t.Errorf("after KeyRight, Index() = %d, want 0", idx)
	}
	if page := got.Paginator.Page; page != 0 {
		t.Errorf("after KeyRight, Paginator.Page = %d, want 0", page)
	}
}

func TestUpdate_FocusRoutesArrows(t *testing.T) {
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
				{ID: "BIT-2", Body: body},
				{ID: "BIT-1", Body: body},
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
}

func TestIsBar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"track", "BIT-2", false},
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
		{"phased bar", &task.Task{ID: "BIT-2.1", Phase: 2, PhaseLabel: "List & read"}, "phase 2 — List & read"},
		{"unphased bar", &task.Task{ID: "BIT-2.1", Phase: 0}, ""},
		{"track", &task.Task{ID: "BIT-2", Phase: 2, PhaseLabel: "List & read"}, ""},
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

func TestUpdate_TabTogglesMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		presses int
		want    viewMode
	}{
		{"default is list", 0, modeList},
		{"one tab to board", 1, modeBoard},
		{"two tabs back to list", 2, modeList},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var mdl tea.Model = New([]*task.Task{{ID: "BIT-1"}})
			for range tt.presses {
				mdl, _ = mdl.Update(tea.KeyPressMsg{Code: tea.KeyTab})
			}

			if got := mdl.(model).mode; got != tt.want {
				t.Errorf("after %d tab(s), mode = %v, want %v", tt.presses, got, tt.want)
			}
		})
	}
}

func TestUpdate_QuitsFromDetail(t *testing.T) {
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

			var mdl tea.Model = New([]*task.Task{{ID: "BIT-1"}})
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
}

func TestUpdate_EscQuitsFromList(t *testing.T) {
	t.Parallel()

	m := New([]*task.Task{{ID: "BIT-1"}})

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("after KeyEsc in list, cmd = nil, want a quit cmd")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("after KeyEsc in list, cmd() = %T, want tea.QuitMsg", cmd())
	}
}
