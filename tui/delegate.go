package tui

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	trackStyle    = lipgloss.NewStyle().Bold(true)
	barStyle      = lipgloss.NewStyle()
	verseStyle         = lipgloss.NewStyle().Faint(true).Italic(true)
	selectedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	selectedBoardStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Reverse(true)
)

type delegate struct {
	board bool
}

func (delegate) Height() int                         { return 1 }
func (delegate) Spacing() int                        { return 0 }
func (delegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

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
	if selected {
		main = selectedStyle
		if d.board {
			main = selectedBoardStyle
		}
	}

	cursor := "  "
	if selected {
		cursor = selectedStyle.Render("▎ ")
	}

	mark := "  "
	if t.Status == "done" {
		mark = "✓ "
	}

	row := cursor + mark + main.Render(indent+t.ID+"  "+t.Title)
	if v := verse(t); v != "" {
		row += "  " + verseStyle.Render(v)
	}

	fmt.Fprint(w, lipgloss.NewStyle().MaxWidth(m.Width()).Render(row))
}
