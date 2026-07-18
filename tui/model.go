package tui

import (
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/charmbracelet/bubbles/list"
)

type item struct {
	t *task.Task
}

func (i item) FilterValue() string { return i.t.Title }
func (i item) Title() string       { return i.t.Title }
func (i item) Description() string { return i.t.ID + " · " + i.t.Status }

type model struct {
	list.Model
}

func New(tasks []*task.Task) model {
	items := make([]list.Item, len(tasks))
	for i, t := range tasks {
		items[i] = item{t: t}
	}
	return model{Model: list.New(items, list.NewDefaultDelegate(), 0, 0)}
}
