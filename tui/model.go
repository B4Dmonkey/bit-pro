package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/B4Dmonkey/bit-pro/task"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/glamour/v2/styles"
	"charm.land/lipgloss/v2"
)

type viewMode int

const (
	modeList viewMode = iota
	modeBoard
)

type item struct {
	t *task.Task
}

type reloadedMsg struct {
	tasks []*task.Task
	err   error
}

type tickMsg struct{}

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
	modalViewport viewport.Model
	help          help.Model
	keys          keyMap
	boardKeys     boardKeyMap
	mode          viewMode
	boardCols     [3]list.Model
	activeCol     int
	detailWidth   int
	listWidth     int
	winWidth      int
	winHeight     int
	height        int
	style         string
	renderer      *glamour.TermRenderer
	reload        func() ([]*task.Task, error)
	detailFocused bool
	modalOpen     bool
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
	if lipgloss.HasDarkBackground(os.Stdin, os.Stdout) {
		style = styles.DarkStyle
	}
	vp := viewport.New()
	mvp := viewport.New()
	var boardCols [3]list.Model
	for i, cards := range groupByStatus(tasks) {
		boardCols[i] = newColumnList(cards)
	}
	return model{Model: l, viewport: vp, modalViewport: mvp, help: help.New(), keys: newKeyMap(), boardKeys: newBoardKeyMap(), style: style, boardCols: boardCols}
}

const pollInterval = time.Second

func tick() tea.Cmd {
	return tea.Tick(pollInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m model) WithReload(r func() ([]*task.Task, error)) model {
	m.reload = r
	return m
}

func (m model) Init() tea.Cmd {
	if m.reload != nil {
		return tick()
	}
	return nil
}

func (m model) reloadCmd() tea.Cmd {
	if m.reload == nil {
		return nil
	}
	reload := m.reload
	return func() tea.Msg {
		tasks, err := reload()
		return reloadedMsg{tasks: tasks, err: err}
	}
}

func (m *model) setTasks(tasks []*task.Task) {
	items := make([]list.Item, len(tasks))
	for i, t := range tasks {
		items[i] = item{t: t}
	}
	m.SetItems(items)
	for i, cards := range groupByStatus(tasks) {
		m.boardCols[i] = newColumnList(cards)
	}
	m.layout()
	m.refreshDetail()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		return m, m.reloadCmd()
	case reloadedMsg:
		m.setTasks(msg.tasks)
		return m, tick()
	case tea.WindowSizeMsg:
		m.winWidth, m.winHeight = msg.Width, msg.Height
		m.layout()
		m.renderer = newRenderer(m.style, max(m.detailWidth-2, 1))
		m.refreshDetail()
		return m, nil
	case tea.KeyPressMsg:
		if m.mode == modeBoard && m.modalOpen {
			return m.updateBoard(msg)
		}
		if key.Matches(msg, m.keys.help) {
			m.help.ShowAll = !m.help.ShowAll
			m.layout()
			return m, nil
		}
		if msg.Code == tea.KeyTab {
			if m.mode == modeList {
				m.mode = modeBoard
			} else {
				m.mode = modeList
			}
			return m, nil
		}
		if m.mode == modeBoard {
			return m.updateBoard(msg)
		}
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
		switch msg.Code {
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

func (m *model) refreshModal() {
	t := m.boardSelected()
	if t == nil {
		m.modalViewport.SetContent("")
		return
	}
	innerW, innerH := modalInner(m.winWidth, m.winHeight)
	m.modalViewport.SetWidth(innerW)
	m.modalViewport.SetHeight(innerH)
	body := t.Body
	if r := newRenderer(m.style, max(innerW, 1)); r != nil {
		if out, err := r.Render(t.Body); err == nil {
			body = out
		}
	}
	m.modalViewport.SetContent(body)
	m.modalViewport.GotoTop()
}

func (m model) helpKeys() help.KeyMap {
	if m.mode == modeBoard {
		return m.boardKeys
	}
	return m.keys
}

func (m *model) layout() {
	listW, detailW := splitWidth(m.winWidth)
	m.help.SetWidth(m.winWidth)
	helpHeight := lipgloss.Height(m.help.View(m.helpKeys()))
	paneHeight := max(m.winHeight-helpHeight, 0)
	m.listWidth, m.detailWidth, m.height = listW, detailW, paneHeight
	m.SetSize(max(listW-2, 0), max(paneHeight-2, 0))
	m.viewport.SetWidth(max(detailW-2, 0))
	m.viewport.SetHeight(max(paneHeight-2, 0))
	colW := m.winWidth / len(boardColumns)
	for i := range m.boardCols {
		m.boardCols[i].SetSize(max(colW-2, 0), max(paneHeight-2, 0))
	}
}

func (m model) View() tea.View {
	v := tea.NewView(m.content())
	v.AltScreen = true
	return v
}

func (m model) content() string {
	if m.mode == modeBoard {
		board := boardView(m)
		if m.modalOpen {
			board = modalView(m, board)
		}
		return lipgloss.JoinVertical(lipgloss.Left, board, m.help.View(m.helpKeys()))
	}
	listTitle := fmt.Sprintf("Tasks (%d)", len(m.Items()))
	listPane := titledBorder(m.Model.View(), listTitle, max(m.listWidth-2, 0), max(m.height-2, 0), !m.detailFocused)
	detailPane := titledBorder(m.viewport.View(), "Details", max(m.detailWidth-2, 0), max(m.height-2, 0), m.detailFocused)
	panes := lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane)
	return lipgloss.JoinVertical(lipgloss.Left, panes, m.help.View(m.helpKeys()))
}

func titledBorder(content, title string, width, height int, active bool) string {
	border := lipgloss.RoundedBorder()

	boxStyle := lipgloss.NewStyle().
		Border(border).
		BorderTop(false).
		Width(width + 2).
		Height(height + 1)
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
