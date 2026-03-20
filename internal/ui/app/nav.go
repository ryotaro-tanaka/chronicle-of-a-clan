package app

import (
	"strings"

	tea "chronicle-of-a-clan/internal/ui/tea"
)

func (m *model) updateNav(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			line := strings.TrimSpace(m.nav.input)
			if line == "" {
				m.nav.input = ""
				return m, nil
			}
			m.nav.lines = append(m.nav.lines, "> "+line)
			m.nav.input = ""
			before := len(m.nav.lines)
			m.nav.session.ExecuteLine(line)
			m.flushSessionOutput()
			if m.nav.session.IsDone() {
				return m, tea.Quit
			}
			if len(m.nav.lines) == before && m.mode != modeNav {
				return m, nil
			}
		case "backspace":
			if len(m.nav.input) > 0 {
				m.nav.input = m.nav.input[:len(m.nav.input)-1]
			}
		default:
			if len(msg.String()) == 1 {
				m.nav.input += msg.String()
			}
		}
	}
	return m, nil
}

func (m *model) navView() string {
	path := m.nav.session.CurrentPath()
	lines := append([]string{}, m.nav.lines...)
	if len(lines) > 20 {
		lines = lines[len(lines)-20:]
	}
	return strings.Join(lines, "\n") + "\n\n" + path + " > " + m.nav.input
}
