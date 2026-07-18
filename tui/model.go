package tui

import (
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type item struct {
	t *task.Task
}

func (i item) FilterValue() string { return i.t.Title }
func (i item) Title() string       { return i.t.Title }
func (i item) Description() string { return i.t.ID + " · " + i.t.Status }

type model struct {
	list.Model
	detailWidth int
}

func New(tasks []*task.Task) model {
	items := make([]list.Item, len(tasks))
	for i, t := range tasks {
		items[i] = item{t: t}
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.SetFilteringEnabled(false)
	return model{Model: l}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		listW, detailW := splitWidth(msg.Width)
		m.SetSize(listW, msg.Height)
		m.detailWidth = detailW
		return m, nil
	}
	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.Model.View()
}

func splitWidth(total int) (listW, detailW int) {
	listW = total * 40 / 100
	detailW = max(total-listW-1, 0)
	return listW, detailW
}

func (m model) selected() *task.Task {
	it, ok := m.SelectedItem().(item)
	if !ok {
		return nil
	}
	return it.t
}
