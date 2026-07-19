package tui

import (
	"strings"

	"github.com/B4Dmonkey/bit-pro/task"
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

func boardView(m model) string {
	colW := m.winWidth / len(boardColumns)
	cols := make([]string, len(m.columns))
	for i, cards := range m.columns {
		lines := make([]string, len(cards))
		for j, t := range cards {
			lines[j] = t.ID + "  " + t.Title
		}
		box := lipgloss.NewStyle().Width(colW).Height(m.height)
		cols[i] = box.Render(strings.Join(lines, "\n"))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}
