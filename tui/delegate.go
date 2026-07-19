package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	trackStyle    = lipgloss.NewStyle().Bold(true)
	barStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	verseStyle    = lipgloss.NewStyle().Faint(true).Italic(true)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true)
)

type delegate struct{}

func (delegate) Height() int                         { return 1 }
func (delegate) Spacing() int                        { return 0 }
func (delegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (delegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
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
	}

	cursor := "  "
	if selected {
		cursor = selectedStyle.Render("▎ ")
	}

	row := cursor + main.Render(indent+t.ID+"  "+t.Title)
	if v := verse(t); v != "" {
		row += "  " + verseStyle.Render(v)
	}

	fmt.Fprint(w, lipgloss.NewStyle().MaxWidth(m.Width()).Render(row))
}
