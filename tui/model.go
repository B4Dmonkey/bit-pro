package tui

import (
	"fmt"
	"strings"

	"github.com/B4Dmonkey/bit-pro/task"
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
	viewport      viewport.Model
	detailWidth   int
	listWidth     int
	height        int
	style         string
	renderer      *glamour.TermRenderer
	detailFocused bool
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
		m.viewport.Width, m.viewport.Height = max(detailW-2, 0), max(msg.Height-2, 0)
		m.renderer = newRenderer(m.style, max(detailW-2, 1))
		m.refreshDetail()
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRight:
			m.detailFocused = true
			return m, nil
		case tea.KeyLeft:
			m.detailFocused = false
			return m, nil
		}
	}
	var cmd tea.Cmd
	if m.detailFocused {
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	prev := m.Index()
	m.Model, cmd = m.Model.Update(msg)
	if m.Index() != prev {
		m.refreshDetail()
	}
	return m, cmd
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
	listTitle := fmt.Sprintf("Tasks (%d)", len(m.Items()))
	listPane := titledBorder(m.Model.View(), listTitle, max(m.listWidth-2, 0), max(m.height-2, 0), !m.detailFocused)
	detailPane := titledBorder(m.viewport.View(), "Details", max(m.detailWidth-2, 0), max(m.height-2, 0), m.detailFocused)
	return lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane)
}

func titledBorder(content, title string, width, height int, active bool) string {
	border := lipgloss.RoundedBorder()

	boxStyle := lipgloss.NewStyle().
		Border(border).
		BorderTop(false).
		Width(width).
		Height(height)
	topStyle := lipgloss.NewStyle()
	if active {
		accent := lipgloss.Color("99")
		boxStyle = boxStyle.BorderForeground(accent)
		topStyle = topStyle.Foreground(accent)
	}

	fill := max(width-lipgloss.Width(title)-3, 0)
	top := border.TopLeft + border.Top + " " + title + " " + strings.Repeat(border.Top, fill) + border.TopRight
	return lipgloss.JoinVertical(lipgloss.Left, topStyle.Render(top), boxStyle.Render(content))
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
