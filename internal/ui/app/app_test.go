package app

import (
	"fmt"
	"strings"
	"testing"

	"chronicle-of-a-clan/internal/core/equipment"
	"chronicle-of-a-clan/internal/core/members"
	"chronicle-of-a-clan/internal/core/monsters"
	"chronicle-of-a-clan/internal/core/quests"
	"chronicle-of-a-clan/internal/core/save"
	"chronicle-of-a-clan/internal/ui/vfs"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNavPartyAndClearFlow(t *testing.T) {
	model := newTestModel(t)
	questPath := "/quests/keys/hunt_ambushjaw_gator"
	model.nav.cwd = mustResolve(t, model.root, questPath)

	if quit := model.executeCommand("party"); quit {
		t.Fatal("party should not quit")
	}
	if model.activeScreen != screenParty {
		t.Fatalf("expected party screen, got %v", model.activeScreen)
	}

	model.pendingQuest = questPath
	model.partyByQuest[questPath] = PartySelection{MemberIDs: []string{"member_0001"}}
	model.activeScreen = screenNav
	if quit := model.executeCommand("clear"); quit {
		t.Fatal("clear should not quit")
	}
	if _, ok := model.partyByQuest[questPath]; ok {
		t.Fatal("expected quest party selection to be cleared")
	}
	if !strings.Contains(strings.Join(model.nav.lines, "\n"), "Party selection cleared.") {
		t.Fatalf("expected clear message, got %v", model.nav.lines)
	}
}

func TestPartySetupModelLimitsSelectionToFour(t *testing.T) {
	roster := makeRoster(t, 5)
	model := NewPartySetupModel("/quests/keys/hunt_ambushjaw_gator", roster, nil)
	for range 4 {
		model.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
		model.HandleKey(tea.KeyMsg{Type: tea.KeyDown})
	}
	model.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if model.SelectedCount() != 4 {
		t.Fatalf("expected 4 selected members, got %d", model.SelectedCount())
	}
	if !strings.Contains(model.message, "up to 4") {
		t.Fatalf("expected selection limit message, got %q", model.message)
	}
}

func TestEquipMemberModelProgression(t *testing.T) {
	baseStats, err := members.BaseStats("GROWTH_005", 1)
	if err != nil {
		t.Fatalf("BaseStats: %v", err)
	}
	membersList := []memberRosterEntry{
		{Member: save.Member{ID: "member_0001", Name: "Ralf"}, Level: 1, BaseStats: baseStats},
		{Member: save.Member{ID: "member_0002", Name: "Mei"}, Level: 1, BaseStats: baseStats},
	}
	catalog := &equipment.Catalog{
		Weapons: []equipment.EquipmentOption{{ID: "R_W_001", Name: "Rental Iron Blade 01", Kind: equipment.KindWeapon, Source: equipment.SourceRental, LevelMin: 1, LevelMax: 5}},
		Armor:   []equipment.EquipmentOption{{ID: "R_A_001", Name: "Rental Vest 01", Kind: equipment.KindArmor, Source: equipment.SourceRental, LevelMin: 1, LevelMax: 5, Prot: 18}},
	}
	model := NewEquipMemberModel("/quests/keys/hunt_ambushjaw_gator", PartySelection{
		MemberIDs:         []string{"member_0001", "member_0002"},
		EquipmentByMember: map[string]MemberEquipment{},
	}, membersList, catalog, save.Inventory{})

	model.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	model.HandleKey(tea.KeyMsg{Type: tea.KeyTab})
	model.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if action := model.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}); action != equipActionNone {
		t.Fatalf("expected first n to move to next member, got %v", action)
	}
	if model.currentIndex != 1 {
		t.Fatalf("expected second member, got index %d", model.currentIndex)
	}
	if action := model.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}); action != equipActionDone {
		t.Fatalf("expected final n to finish, got %v", action)
	}

	selection := model.Selection()
	if selection.EquipmentByMember["member_0001"].WeaponID != "R_W_001" {
		t.Fatalf("expected weapon selection to persist, got %+v", selection)
	}
	if selection.EquipmentByMember["member_0001"].ArmorID != "R_A_001" {
		t.Fatalf("expected armor selection to persist, got %+v", selection)
	}
}

func TestNavModelAllowsSpaceInput(t *testing.T) {
	nav := NewNavModel(vfs.NewTree())
	nav.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("cd")})
	nav.HandleKey(tea.KeyMsg{Type: tea.KeySpace})
	nav.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("clan")})

	if got := string(nav.input); got != "cd clan" {
		t.Fatalf("expected spaced input, got %q", got)
	}
}

func TestNavModelTabCompletesCommandsAndPaths(t *testing.T) {
	nav := NewNavModel(vfs.NewTree())

	nav.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("st")})
	nav.HandleKey(tea.KeyMsg{Type: tea.KeyTab})
	if got := string(nav.input); got != "status" {
		t.Fatalf("expected command completion to status, got %q", got)
	}

	nav.input = nil
	nav.cursor = 0
	nav.suggestions = nil
	nav.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("cd c")})
	nav.HandleKey(tea.KeyMsg{Type: tea.KeyTab})
	if got := string(nav.input); got != "cd clan/" {
		t.Fatalf("expected path completion to clan/, got %q", got)
	}
}

func newTestModel(t *testing.T) *Model {
	t.Helper()
	keyQuests, err := quests.LoadKeyQuests()
	if err != nil {
		t.Skipf("LoadKeyQuests: %v", err)
	}
	bossProfiles, err := monsters.LoadBossProfiles()
	if err != nil {
		t.Skipf("LoadBossProfiles: %v", err)
	}
	root := vfs.NewTree()
	vfs.AttachQuests(root, save.State{KeyQuestCurrentOrder: 1}, keyQuests, bossProfiles)

	memberList := makeRoster(t, 2)
	memberMap := make(map[string]memberRosterEntry, len(memberList))
	for _, member := range memberList {
		memberMap[member.Member.ID] = member
	}

	catalog, err := equipment.LoadCatalog()
	if err != nil {
		t.Skipf("LoadCatalog: %v", err)
	}

	return &Model{
		state:        save.State{ClanName: "Chronicle Clan", CurrentDay: 1, Gold: 120, Fame: 0},
		root:         root,
		bossProfiles: bossProfiles,
		catalog:      catalog,
		partyByQuest: map[string]PartySelection{},
		activeScreen: screenNav,
		nav:          NewNavModel(root),
		memberList:   memberList,
		membersByID:  memberMap,
		detailed: save.DetailedState{
			Inventory: save.Inventory{
				Weapons: []save.InventoryEntry{{ID: "CW_001", Count: 1}},
				Armor:   []save.InventoryEntry{{ID: "CA_001", Count: 1}},
			},
		},
	}
}

func makeRoster(t *testing.T, count int) []memberRosterEntry {
	t.Helper()
	list := make([]memberRosterEntry, 0, count)
	for i := 0; i < count; i++ {
		level := 1
		stats, err := members.BaseStats("GROWTH_005", level)
		if err != nil {
			t.Fatalf("BaseStats: %v", err)
		}
		list = append(list, memberRosterEntry{
			Member: save.Member{
				ID:           fmt.Sprintf("member_%04d", i+1),
				Name:         []string{"Ralf", "Mei", "Doran", "Sia", "Toma"}[i],
				GrowthTypeID: "GROWTH_005",
			},
			Level:     level,
			BaseStats: stats,
		})
	}
	return list
}

func mustResolve(t *testing.T, root *vfs.Node, path string) *vfs.Node {
	t.Helper()
	node, err := vfs.Resolve(root, root, path)
	if err != nil {
		t.Fatalf("Resolve(%s): %v", path, err)
	}
	return node
}
