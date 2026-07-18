package tui

import tea "github.com/charmbracelet/bubbletea"

func Run(m model) error {
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}
