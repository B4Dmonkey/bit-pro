package tui

import (
	"fmt"
	"strings"

	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type viewMode int

const (
	modeList viewMode = iota
	modeBoard
)

type item struct {
	t *task.Task
}

func (i item) FilterValue() string { return i.t.Title }
func (i item) Title() string       { return i.t.Title }
func (i item) Description() string { return i.t.ID + " · " + i.t.Status }

type keyMap struct {
	focus key.Binding
	move  key.Binding
	help  key.Binding
	quit  key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		focus: key.NewBinding(key.WithKeys("left", "right"), key.WithHelp("←/→", "focus")),
		move:  key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "move·scroll")),
		help:  key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		quit:  key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.focus, k.move, k.help, k.quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.focus, k.move}, {k.help, k.quit}}
}

type model struct {
	list.Model
	viewport      viewport.Model
	help          help.Model
	keys          keyMap
	mode          viewMode
	boardCols     [3]list.Model
	detailWidth   int
	listWidth     int
	winWidth      int
	winHeight     int
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
	l := list.New(items, delegate{}, 0, 0)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	style := styles.LightStyle
	if termenv.HasDarkBackground() {
		style = styles.DarkStyle
	}
	vp := viewport.New(0, 0)
	var boardCols [3]list.Model
	for i, cards := range groupByStatus(tasks) {
		boardCols[i] = newColumnList(cards)
	}
	return model{Model: l, viewport: vp, help: help.New(), keys: newKeyMap(), style: style, boardCols: boardCols}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.winWidth, m.winHeight = msg.Width, msg.Height
		m.layout()
		m.renderer = newRenderer(m.style, max(m.detailWidth-2, 1))
		m.refreshDetail()
		return m, nil
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.help) {
			m.help.ShowAll = !m.help.ShowAll
			m.layout()
			return m, nil
		}
		if msg.Type == tea.KeyTab {
			if m.mode == modeList {
				m.mode = modeBoard
			} else {
				m.mode = modeList
			}
			return m, nil
		}
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

func (m *model) layout() {
	listW, detailW := splitWidth(m.winWidth)
	m.help.Width = m.winWidth
	helpHeight := lipgloss.Height(m.help.View(m.keys))
	paneHeight := max(m.winHeight-helpHeight, 0)
	m.listWidth, m.detailWidth, m.height = listW, detailW, paneHeight
	m.SetSize(max(listW-2, 0), max(paneHeight-2, 0))
	m.viewport.Width, m.viewport.Height = max(detailW-2, 0), max(paneHeight-2, 0)
	colW := m.winWidth / len(boardColumns)
	for i := range m.boardCols {
		m.boardCols[i].SetSize(max(colW-2, 0), max(paneHeight-2, 0))
	}
}

func (m model) View() string {
	if m.mode == modeBoard {
		return lipgloss.JoinVertical(lipgloss.Left, boardView(m), m.help.View(m.keys))
	}
	listTitle := fmt.Sprintf("Tasks (%d)", len(m.Items()))
	listPane := titledBorder(m.Model.View(), listTitle, max(m.listWidth-2, 0), max(m.height-2, 0), !m.detailFocused)
	detailPane := titledBorder(m.viewport.View(), "Details", max(m.detailWidth-2, 0), max(m.height-2, 0), m.detailFocused)
	panes := lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane)
	return lipgloss.JoinVertical(lipgloss.Left, panes, m.help.View(m.keys))
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

func isBar(id string) bool {
	return strings.Contains(id, ".")
}

func verse(t *task.Task) string {
	if !isBar(t.ID) || t.Phase == 0 {
		return ""
	}
	return fmt.Sprintf("phase %d — %s", t.Phase, t.PhaseLabel)
}

func (m model) selected() *task.Task {
	it, ok := m.SelectedItem().(item)
	if !ok {
		return nil
	}
	return it.t
}
