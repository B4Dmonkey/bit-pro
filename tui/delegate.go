package tui

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/B4Dmonkey/bit-pro/task"
)

var (
	trackStyle         = lipgloss.NewStyle().Bold(true)
	barStyle           = lipgloss.NewStyle()
	verseStyle         = lipgloss.NewStyle().Faint(true).Italic(true)
	selectedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	selectedBoardStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Reverse(true)
	queuedColor        = lipgloss.Color("6")
)

type delegate struct {
	board  bool
	queued map[string]bool
}

func (delegate) Height() int                         { return 1 }
func (delegate) Spacing() int                        { return 0 }
func (delegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (d delegate) resolveStyle(main lipgloss.Style, t *task.Task, selected bool) lipgloss.Style {
	if !selected {
		if d.queued[t.ID] {
			return main.Foreground(queuedColor)
		}

		if !t.Approved {
			return main.Foreground(lipgloss.Color("3"))
		}

		return main
	}

	if d.board {
		return selectedBoardStyle
	}

	main = main.Bold(true)

	if !t.Approved {
		return main.Foreground(lipgloss.Color("3"))
	}

	return main
}

func (d delegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	it, ok := listItem.(item)
	if !ok {
		return
	}

	t := it.t
	selected := index == m.Index()
	indent := ""
	main := trackStyle

	if isBar(t.ID) {
		indent = "  "
		main = barStyle
	}

	main = d.resolveStyle(main, t, selected)

	cursor := "  "
	if selected {
		cursor = selectedStyle.Render("▎ ")
	}

	mark := "  "

	switch t.Status {
	case task.StatusDone:
		mark = "✓ "
	case task.StatusDoing:
		if !d.board {
			mark = "→ "
		}
	}

	row := cursor + mark + main.Render(indent+t.ID+"  "+t.Title)
	if v := verse(t); v != "" {
		row += "  " + verseStyle.Render(v)
	}

	fmt.Fprint(w, lipgloss.NewStyle().MaxWidth(m.Width()).Render(row))
}
