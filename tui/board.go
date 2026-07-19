package tui

import (
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/charmbracelet/bubbles/list"
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

func boardView(m model) string {
	colW := m.winWidth / len(boardColumns)
	cols := make([]string, len(m.boardCols))
	for i := range m.boardCols {
		box := lipgloss.NewStyle().Width(colW).Height(m.height)
		cols[i] = box.Render(m.boardCols[i].View())
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}
