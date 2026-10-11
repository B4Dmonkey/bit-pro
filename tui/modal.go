package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	keyCtrlC = "ctrl+c"
	keyEsc   = "esc"
	keyLeft  = "left"
	keyRight = "right"
)

func (m *model) pageModal(delta int) {
	m.Select(min(max(m.Index()+delta, 0), len(m.Items())-1))
	m.refreshDetail()
	m.refreshModal()
}

func (m model) updateModal(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
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

func modalInner(winWidth, winHeight int) (innerW, innerH int) {
	modalW := 2 * winWidth / 3
	modalH := winHeight - 6

	return max(modalW-4, 1), max(modalH-3, 1)
}

func modalView(m model, base string) string {
	t := m.selected()
	if t == nil {
		return base
	}

	innerW, innerH := modalInner(m.winWidth, m.winHeight)
	title := t.ID + " — " + t.Title
	box := titledBorder(m.modalViewport.View(), title, innerW, innerH, true)
	cx := max((m.winWidth-lipgloss.Width(box))/2, 0)
	cy := max((lipgloss.Height(base)-lipgloss.Height(box))/2, 0)
	baseLayer := lipgloss.NewLayer(base)
	modal := lipgloss.NewLayer(box).X(cx).Y(cy).Z(1)

	return lipgloss.NewCompositor(baseLayer, modal).Render()
}
