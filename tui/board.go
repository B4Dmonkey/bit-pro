package tui

import (
	"fmt"
	"slices"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/B4Dmonkey/bit-pro/task"
)

const (
	keyCtrlC = "ctrl+c"
	keyEsc   = "esc"
	keyLeft  = "left"
	keyRight = "right"
)

type boardKeyMap struct {
	column key.Binding
	card   key.Binding
	list   key.Binding
	help   key.Binding
	quit   key.Binding
}

func newBoardKeyMap() boardKeyMap {
	return boardKeyMap{
		column: key.NewBinding(key.WithKeys("left", "right"), key.WithHelp("←/→", "column")),
		card:   key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "card")),
		list:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "list")),
		help:   key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		quit:   key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}
}

func (k boardKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.column, k.card, k.list, k.help, k.quit}
}

func (k boardKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.column, k.card}, {k.list, k.help, k.quit}}
}

type boardColumn struct {
	title  string
	status string
}

var boardColumns = [3]boardColumn{
	{title: "To Do", status: task.StatusTodo},
	{title: "Doing", status: task.StatusDoing},
	{title: "Done", status: task.StatusDone},
}

func groupByStatus(tasks []*task.Task) [3][]*task.Task {
	var cols [3][]*task.Task

	for _, t := range tasks {
		for i, col := range boardColumns {
			if t.Status == col.status {
				if col.status == task.StatusTodo && !t.Approved {
					break
				}

				cols[i] = append(cols[i], t)

				break
			}
		}
	}

	return cols
}

func newColumnList(tasks []*task.Task, queued map[string]bool) list.Model {
	items := make([]list.Item, len(tasks))
	for i, t := range tasks {
		items[i] = item{t: t}
	}

	l := list.New(items, delegate{board: true, queued: queued}, 0, 0)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)

	return l
}

func defaultColumn(cols [3][]*task.Task) int {
	if len(cols[1]) > 0 {
		return 1
	}

	for i, c := range cols {
		if len(c) > 0 {
			return i
		}
	}

	return 0
}

type boardEntry struct {
	col, pos int
	t        *task.Task
}

func flattenBoard(cols [3]list.Model) []boardEntry {
	var entries []boardEntry

	for col := range cols {
		for pos, it := range cols[col].Items() {
			if bi, ok := it.(item); ok {
				entries = append(entries, boardEntry{col: col, pos: pos, t: bi.t})
			}
		}
	}

	return entries
}

func firstBarIndex(items []list.Item) int {
	for i, it := range items {
		if bi, ok := it.(item); ok && isBar(bi.t.ID) {
			return i
		}
	}

	return 0
}

func (m *model) pageModal(delta int) {
	order := flattenBoard(m.boardCols)
	if len(order) == 0 {
		return
	}

	idx := 0

	if cur := m.boardSelected(); cur != nil {
		if i := slices.IndexFunc(order, func(e boardEntry) bool { return e.t.ID == cur.ID }); i >= 0 {
			idx = i
		}
	}

	idx = min(max(idx+delta, 0), len(order)-1)
	m.activeCol = order[idx].col
	m.boardCols[m.activeCol].Select(order[idx].pos)
	m.refreshModal()
}

func (m model) boardSelected() *task.Task {
	it, ok := m.boardCols[m.activeCol].SelectedItem().(item)
	if !ok {
		return nil
	}

	return it.t
}

func (m model) updateModalBoard(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", keyEsc:
		m.modalOpen = false
		return m, nil
	case keyCtrlC:
		return m, tea.Quit
	case "up", "down", "j", "k":
		var cmd tea.Cmd

		m.modalViewport, cmd = m.modalViewport.Update(msg)

		return m, cmd
	case keyLeft, "h":
		m.pageModal(-1)
		return m, nil
	case keyRight, "l":
		m.pageModal(1)
		return m, nil
	}

	return m, nil
}

func (m model) handleBoardApprove() (tea.Model, tea.Cmd) {
	if m.approve != nil {
		if t := m.boardSelected(); t != nil {
			if !t.Approved && isBar(t.ID) {
				m.pendingApprovalID = t.ID
			}

			_ = m.approve(t.ID, !t.Approved)

			return m, m.reloadCmd()
		}
	}

	return m, nil
}

func (m model) updateBoard(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.modalOpen {
		return m.updateModalBoard(msg)
	}

	if msg.Code == tea.KeyEnter {
		if m.boardSelected() != nil {
			m.modalOpen = true
			m.refreshModal()
		}

		return m, nil
	}

	if msg.Code == ' ' {
		return m.handleBoardApprove()
	}

	switch msg.String() {
	case "q", keyEsc, keyCtrlC:
		return m, tea.Quit
	case "e":
		m.enqueueSelected()
	case keyRight:
		if m.activeCol < len(boardColumns)-1 {
			m.activeCol++
		}
	case keyLeft:
		if m.activeCol > 0 {
			m.activeCol--
		}
	case "up", "down":
		var cmd tea.Cmd

		m.boardCols[m.activeCol], cmd = m.boardCols[m.activeCol].Update(msg)

		return m, cmd
	}

	return m, nil
}

func modalInner(winWidth, winHeight int) (innerW, innerH int) {
	modalW := 2 * winWidth / 3
	modalH := winHeight - 6

	return max(modalW-4, 1), max(modalH-3, 1)
}

func modalView(m model, board string) string {
	t := m.boardSelected()
	if t == nil {
		return board
	}

	innerW, innerH := modalInner(m.winWidth, m.winHeight)
	title := t.ID + " — " + t.Title
	box := titledBorder(m.modalViewport.View(), title, innerW, innerH, true)
	cx := max((m.winWidth-lipgloss.Width(box))/2, 0)
	cy := max((lipgloss.Height(board)-lipgloss.Height(box))/2, 0)
	base := lipgloss.NewLayer(board)
	modal := lipgloss.NewLayer(box).X(cx).Y(cy).Z(1)

	return lipgloss.NewCompositor(base, modal).Render()
}

func playPromptView(m model, board string) string {
	prompt := "Play " + m.playPromptTitle + "? (y / n)"
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		BorderForeground(lipgloss.Color("4")).
		Background(lipgloss.Color("4")).
		Foreground(lipgloss.Color("0")).
		Render(prompt)
	cx := max((lipgloss.Width(board)-lipgloss.Width(box))/2, 0)
	cy := max((lipgloss.Height(board)-lipgloss.Height(box))/2, 0)
	base := lipgloss.NewLayer(board)
	overlay := lipgloss.NewLayer(box).X(cx).Y(cy).Z(1)

	return lipgloss.NewCompositor(base, overlay).Render()
}

func boardView(m model) string {
	colW := m.winWidth / len(boardColumns)

	cols := make([]string, len(m.boardCols))
	for i := range m.boardCols {
		title := fmt.Sprintf("%s (%d)", boardColumns[i].title, len(m.boardCols[i].Items()))
		cols[i] = titledBorder(m.boardCols[i].View(), title, max(colW-2, 0), max(m.height-2, 0), i == m.activeCol)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}
