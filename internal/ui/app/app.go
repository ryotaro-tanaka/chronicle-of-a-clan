package app

import (
	"bytes"
	"fmt"
	"strings"

	"chronicle-of-a-clan/internal/core/equipment"
	"chronicle-of-a-clan/internal/core/members"
	"chronicle-of-a-clan/internal/core/monsters"
	"chronicle-of-a-clan/internal/core/save"
	"chronicle-of-a-clan/internal/ui/repl"
	"chronicle-of-a-clan/internal/ui/vfs"
	tea "github.com/charmbracelet/bubbletea"
)

type memberView struct {
	ID    string
	Name  string
	Level int
	Stats members.Stats
}

type assignment struct {
	WeaponID string
	ArmorID  string
}

type partySelection struct {
	MemberIDs   []string
	Assignments map[string]assignment
}

type screenMode int

const (
	modeNav screenMode = iota
	modeParty
	modeEquip
)

type model struct {
	mode           screenMode
	nav            navModel
	outBuf         *bytes.Buffer
	errBuf         *bytes.Buffer
	party          partySetupModel
	equip          equipMemberModel
	state          save.State
	members        []memberView
	partyByQuest   map[string]partySelection
	activeQuestKey string
}

type navModel struct {
	session *repl.Session
	input   string
	lines   []string
}

type partySetupModel struct {
	cursor     int
	selected   map[string]bool
	message    string
	memberList []memberView
}

type equipMemberModel struct {
	idx            int
	focusArmor     bool
	weaponCursor   int
	armorCursor    int
	weaponOptions  []equipment.Item
	armorOptions   []equipment.Item
	weaponSelected int
	armorSelected  int
	memberList     []memberView
	message        string
}

func Run(state save.State, root *vfs.Node, bossProfiles *monsters.BossProfilesConfig) int {
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	session := repl.NewSession(state, root, bossProfiles, outBuf, errBuf)

	m := model{
		mode: modeNav,
		nav: navModel{
			session: session,
			lines:   []string{"Bubble Tea Navigation", "Type ls, cd, status, exit."},
		},
		state:        state,
		outBuf:       outBuf,
		errBuf:       errBuf,
		members:      buildMembers(state),
		partyByQuest: map[string]partySelection{},
	}

	session.SetActionHooks(func(questPath string) {
		m.activeQuestKey = questPath
		m.startPartyScreen(questPath)
	}, func(questPath string) {
		delete(m.partyByQuest, questPath)
		m.nav.lines = append(m.nav.lines, "Party selection cleared.")
	})

	p := tea.NewProgram(&m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("application error: %v\n", err)
		return 1
	}
	return 0
}

func (m *model) flushSessionOutput() {
	if m.outBuf != nil && m.outBuf.Len() > 0 {
		for _, ln := range strings.Split(strings.TrimRight(m.outBuf.String(), "\n"), "\n") {
			if ln != "" {
				m.nav.lines = append(m.nav.lines, ln)
			}
		}
		m.outBuf.Reset()
	}
	if m.errBuf != nil && m.errBuf.Len() > 0 {
		for _, ln := range strings.Split(strings.TrimRight(m.errBuf.String(), "\n"), "\n") {
			if ln != "" {
				m.nav.lines = append(m.nav.lines, "ERR: "+ln)
			}
		}
		m.errBuf.Reset()
	}
}

func buildMembers(state save.State) []memberView {
	out := make([]memberView, 0, len(state.Members))
	for _, m := range state.Members {
		lvl := members.LevelFromXP(m.XP)
		st, err := members.StatsFor(m.GrowthTypeID, lvl)
		if err != nil {
			st = members.Stats{Might: 180, Mastery: 180, Tactics: 180, Survival: 180}
		}
		out = append(out, memberView{ID: m.ID, Name: m.Name, Level: lvl, Stats: st})
	}
	return out
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case modeParty:
		return m.updateParty(msg)
	case modeEquip:
		return m.updateEquip(msg)
	default:
		return m.updateNav(msg)
	}
}

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

func (m *model) View() string {
	switch m.mode {
	case modeParty:
		return m.partyView()
	case modeEquip:
		return m.equipView()
	default:
		return m.navView()
	}
}

func (m *model) navView() string {
	path := m.nav.session.CurrentPath()
	lines := append([]string{}, m.nav.lines...)
	if len(lines) > 20 {
		lines = lines[len(lines)-20:]
	}
	return strings.Join(lines, "\n") + "\n\n" + path + " > " + m.nav.input
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
