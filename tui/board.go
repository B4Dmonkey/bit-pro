package tui

import (
	"fmt"

	"github.com/B4Dmonkey/bit-pro/task"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	{title: "To Do", status: "todo"},
	{title: "Doing", status: "doing"},
	{title: "Done", status: "done"},
}

func groupByStatus(tasks []*task.Task) [3][]*task.Task {
	var cols [3][]*task.Task
	for _, t := range tasks {
		for i, col := range boardColumns {
			if t.Status == col.status {
				cols[i] = append(cols[i], t)
				break
			}
		}
	}
	return cols
}

func newColumnList(tasks []*task.Task) list.Model {
	items := make([]list.Item, len(tasks))
	for i, t := range tasks {
		items[i] = item{t: t}
	}
	l := list.New(items, delegate{}, 0, 0)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	return l
}

func (m model) boardSelected() *task.Task {
	it, ok := m.boardCols[m.activeCol].SelectedItem().(item)
	if !ok {
		return nil
	}
	return it.t
}

func (m model) updateBoard(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.modalOpen {
		switch msg.String() {
		case "q", "esc":
			m.modalOpen = false
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
		return m, nil
	}
	if msg.Code == tea.KeyEnter {
		if m.boardSelected() != nil {
			m.modalOpen = true
			m.refreshModal()
		}
		return m, nil
	}
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	case "right":
		if m.activeCol < len(boardColumns)-1 {
			m.activeCol++
		}
	case "left":
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

func boardView(m model) string {
	colW := m.winWidth / len(boardColumns)
	cols := make([]string, len(m.boardCols))
	for i := range m.boardCols {
		title := fmt.Sprintf("%s (%d)", boardColumns[i].title, len(m.boardCols[i].Items()))
		cols[i] = titledBorder(m.boardCols[i].View(), title, max(colW-2, 0), max(m.height-2, 0), i == m.activeCol)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}
