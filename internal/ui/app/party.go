package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) updateParty(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch k.String() {
	case "esc":
		m.mode = modeNav
	case "up":
		if m.party.cursor > 0 {
			m.party.cursor--
		}
	case "down":
		if m.party.cursor < len(m.party.memberList)-1 {
			m.party.cursor++
		}
	case "enter":
		if len(m.party.memberList) == 0 {
			return m, nil
		}
		id := m.party.memberList[m.party.cursor].ID
		if m.party.selected[id] {
			delete(m.party.selected, id)
		} else if len(m.party.selected) < 4 {
			m.party.selected[id] = true
		}
	case "f":
		ids := m.selectedMemberIDs()
		if len(ids) == 0 {
			m.party.message = "Select at least 1 member."
			return m, nil
		}
		m.partyByQuest[m.activeQuestKey] = partySelection{MemberIDs: ids, Assignments: m.currentAssignments()}
		m.startEquipScreen()
	}
	return m, nil
}

func (m *model) partyView() string {
	var b strings.Builder
	b.WriteString("Party Setup — Select Members\n\n")
	for i, mem := range m.party.memberList {
		cursor := " "
		if i == m.party.cursor {
			cursor = ">"
		}
		check := " "
		if m.party.selected[mem.ID] {
			check = "x"
		}
		b.WriteString(fmt.Sprintf("%s [%s] %-10s Lv%d\n", cursor, check, mem.Name, mem.Level))
	}
	b.WriteString(fmt.Sprintf("\nSelected: %d/4\n", len(m.party.selected)))
	if m.party.message != "" {
		b.WriteString(m.party.message + "\n")
	}
	b.WriteString("Enter: Toggle  F: Confirm Members  Esc: Cancel")
	return b.String()
}

func (m *model) startPartyScreen(questPath string) {
	m.mode = modeParty
	current := m.partyByQuest[questPath]
	selected := map[string]bool{}
	for _, id := range current.MemberIDs {
		selected[id] = true
	}
	m.party = partySetupModel{memberList: m.members, selected: selected}
}

func (m *model) selectedMemberIDs() []string {
	ids := make([]string, 0, len(m.party.selected))
	for _, mem := range m.party.memberList {
		if m.party.selected[mem.ID] {
			ids = append(ids, mem.ID)
		}
	}
	return ids
}

func (m *model) currentAssignments() map[string]assignment {
	sel, ok := m.partyByQuest[m.activeQuestKey]
	if !ok || sel.Assignments == nil {
		return map[string]assignment{}
	}
	out := map[string]assignment{}
	for k, v := range sel.Assignments {
		out[k] = v
	}
	return out
}
