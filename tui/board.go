package tui

import "github.com/B4Dmonkey/bit-pro/task"

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
