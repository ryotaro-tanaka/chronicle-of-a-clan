package app

import (
	"fmt"
	"strings"

	"chronicle-of-a-clan/internal/core/equipment"
	tea "chronicle-of-a-clan/internal/ui/tea"
)

func (m *model) updateEquip(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch k.String() {
	case "esc":
		m.mode = modeParty
	case "tab":
		m.equip.focusArmor = !m.equip.focusArmor
	case "up":
		if m.equip.focusArmor {
			if m.equip.armorCursor > 0 {
				m.equip.armorCursor--
			}
		} else if m.equip.weaponCursor > 0 {
			m.equip.weaponCursor--
		}
	case "down":
		if m.equip.focusArmor {
			if m.equip.armorCursor < len(m.equip.armorOptions)-1 {
				m.equip.armorCursor++
			}
		} else if m.equip.weaponCursor < len(m.equip.weaponOptions)-1 {
			m.equip.weaponCursor++
		}
	case "enter":
		if m.equip.focusArmor {
			m.equip.armorSelected = m.equip.armorCursor
		} else {
			m.equip.weaponSelected = m.equip.weaponCursor
		}
		m.persistEquipmentSelection()
	case "n":
		m.persistEquipmentSelection()
		if m.equip.idx < len(m.equip.memberList)-1 {
			m.equip.idx++
			m.prepareCurrentEquipOptions()
		} else {
			m.mode = modeNav
			m.nav.lines = append(m.nav.lines, "Party setup completed.")
		}
	}
	return m, nil
}

func (m *model) equipView() string {
	if len(m.equip.memberList) == 0 {
		return "No members selected."
	}
	mem := m.equip.memberList[m.equip.idx]
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Equip Member — %s\n\n", mem.Name))
	b.WriteString("Weapon\n")
	for i, w := range m.equip.weaponOptions {
		cursor := " "
		if !m.equip.focusArmor && i == m.equip.weaponCursor {
			cursor = ">"
		}
		b.WriteString(fmt.Sprintf("%s %s  %s\n", cursor, w.ID, w.Name))
	}
	b.WriteString("\nArmor\n")
	for i, a := range m.equip.armorOptions {
		cursor := " "
		if m.equip.focusArmor && i == m.equip.armorCursor {
			cursor = ">"
		}
		b.WriteString(fmt.Sprintf("%s %s  %s\n", cursor, a.ID, a.Name))
	}
	b.WriteString(fmt.Sprintf("\nCurrent Stats\nMight: %d\nMastery: %d\nTactics: %d\nSurvival: %d\n", mem.Stats.Might, mem.Stats.Mastery, mem.Stats.Tactics, mem.Stats.Survival))
	if len(m.equip.weaponOptions) > 0 {
		w := m.equip.weaponOptions[m.equip.weaponCursor]
		b.WriteString(fmt.Sprintf("\nPreview\nWeapon: +Might %d +Mastery %d +Tactics %d +Survival %d\n", w.Modifiers.Might, w.Modifiers.Mastery, w.Modifiers.Tactics, w.Modifiers.Survival))
	}
	if len(m.equip.armorOptions) > 0 {
		a := m.equip.armorOptions[m.equip.armorCursor]
		b.WriteString(fmt.Sprintf("Armor: PROT %d\n", a.Prot))
	}
	if m.equip.message != "" {
		b.WriteString(m.equip.message + "\n")
	}
	b.WriteString("\nEnter: Confirm  Tab: Switch weapon/armor  N: Next member  Esc: Back")
	return b.String()
}

func (m *model) startEquipScreen() {
	sel := m.partyByQuest[m.activeQuestKey]
	list := make([]memberView, 0, len(sel.MemberIDs))
	for _, id := range sel.MemberIDs {
		for _, mem := range m.members {
			if mem.ID == id {
				list = append(list, mem)
			}
		}
	}
	m.mode = modeEquip
	m.equip = equipMemberModel{memberList: list, weaponSelected: -1, armorSelected: -1}
	m.prepareCurrentEquipOptions()
}

func (m *model) prepareCurrentEquipOptions() {
	if len(m.equip.memberList) == 0 {
		return
	}
	mem := m.equip.memberList[m.equip.idx]
	m.equip.weaponOptions = equipment.EligibleWeapons(m.state.Inventory, mem.Level, mem.Stats)
	m.equip.armorOptions = equipment.EligibleArmor(m.state.Inventory, mem.Level, mem.Stats)
	m.equip.weaponCursor = 0
	m.equip.armorCursor = 0
	m.equip.focusArmor = false

	sel := m.partyByQuest[m.activeQuestKey]
	assign := sel.Assignments[mem.ID]
	m.equip.weaponSelected = 0
	for i, w := range m.equip.weaponOptions {
		if w.ID == assign.WeaponID {
			m.equip.weaponSelected = i
			m.equip.weaponCursor = i
		}
	}
	m.equip.armorSelected = 0
	for i, a := range m.equip.armorOptions {
		if a.ID == assign.ArmorID {
			m.equip.armorSelected = i
			m.equip.armorCursor = i
		}
	}
}

func (m *model) persistEquipmentSelection() {
	if len(m.equip.memberList) == 0 {
		return
	}
	mem := m.equip.memberList[m.equip.idx]
	sel := m.partyByQuest[m.activeQuestKey]
	if sel.Assignments == nil {
		sel.Assignments = map[string]assignment{}
	}
	as := sel.Assignments[mem.ID]
	if len(m.equip.weaponOptions) > 0 {
		as.WeaponID = m.equip.weaponOptions[m.equip.weaponSelected].ID
	}
	if len(m.equip.armorOptions) > 0 {
		as.ArmorID = m.equip.armorOptions[m.equip.armorSelected].ID
	}
	sel.Assignments[mem.ID] = as
	m.partyByQuest[m.activeQuestKey] = sel
}
