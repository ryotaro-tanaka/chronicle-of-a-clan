package app

import (
	"bytes"
	"fmt"
	"io"
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
	errOut         io.Writer
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

func Run(state save.State, root *vfs.Node, bossProfiles *monsters.BossProfilesConfig, errOut io.Writer) int {
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
		errOut:       errOut,
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
		if m.errOut != nil {
			fmt.Fprintf(m.errOut, "application error: %v\n", err)
		}
		return 1
	}
	return 0
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
