package tui

import (
	"fmt"

	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

func (m model) updateBoard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

func boardView(m model) string {
	colW := m.winWidth / len(boardColumns)
	cols := make([]string, len(m.boardCols))
	for i := range m.boardCols {
		title := fmt.Sprintf("%s (%d)", boardColumns[i].title, len(m.boardCols[i].Items()))
		cols[i] = titledBorder(m.boardCols[i].View(), title, max(colW-2, 0), max(m.height-2, 0), i == m.activeCol)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}
