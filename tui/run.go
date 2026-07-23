package tui

import tea "charm.land/bubbletea/v2"

func Run(m model) error {
	_, err := tea.NewProgram(m).Run()
	return err
}
