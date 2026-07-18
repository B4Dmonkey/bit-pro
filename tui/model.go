package tui

import (
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type item struct {
	t *task.Task
}

func (i item) FilterValue() string { return i.t.Title }
func (i item) Title() string       { return i.t.Title }
func (i item) Description() string { return i.t.ID + " · " + i.t.Status }

type model struct {
	list.Model
	viewport    viewport.Model
	detailWidth int
	listWidth   int
	height      int
	style       string
	renderer    *glamour.TermRenderer
}

func New(tasks []*task.Task) model {
	items := make([]list.Item, len(tasks))
	for i, t := range tasks {
		items[i] = item{t: t}
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.SetFilteringEnabled(false)
	style := styles.LightStyle
	if termenv.HasDarkBackground() {
		style = styles.DarkStyle
	}
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	vp.KeyMap = viewport.KeyMap{
		HalfPageUp:   key.NewBinding(key.WithKeys("ctrl+u")),
		HalfPageDown: key.NewBinding(key.WithKeys("ctrl+d")),
	}
	return model{Model: l, viewport: vp, style: style}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		listW, detailW := splitWidth(msg.Width)
		m.listWidth, m.height = listW, msg.Height
		m.SetSize(max(listW-2, 0), max(msg.Height-2, 0))
		m.detailWidth = detailW
		m.viewport.Width, m.viewport.Height = detailW, msg.Height
		m.renderer = newRenderer(m.style, max(detailW-2, 1))
		m.refreshDetail()
		return m, nil
	}
	prev := m.Index()
	var cmd, vpCmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)
	if m.Index() != prev {
		m.refreshDetail()
	}
	return m, tea.Batch(cmd, vpCmd)
}

func newRenderer(style string, width int) *glamour.TermRenderer {
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}
	return r
}

func (m *model) refreshDetail() {
	t := m.selected()
	if t == nil {
		m.viewport.SetContent("")
		return
	}
	body := t.Body
	if m.renderer != nil {
		if out, err := m.renderer.Render(t.Body); err == nil {
			body = out
		}
	}
	m.viewport.SetContent(body)
	m.viewport.GotoTop()
}

func (m model) View() string {
	listBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(max(m.listWidth-2, 0)).
		Height(max(m.height-2, 0)).
		Render(m.Model.View())
	return lipgloss.JoinHorizontal(lipgloss.Top, listBox, m.viewport.View())
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
