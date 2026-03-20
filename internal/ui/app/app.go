package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"chronicle-of-a-clan/internal/core/equipment"
	"chronicle-of-a-clan/internal/core/members"
	"chronicle-of-a-clan/internal/core/monsters"
	"chronicle-of-a-clan/internal/core/save"
	"chronicle-of-a-clan/internal/core/status"
	"chronicle-of-a-clan/internal/ui/format"
	"chronicle-of-a-clan/internal/ui/vfs"
	tea "github.com/charmbracelet/bubbletea"
)

type screenMode int

const (
	screenNav screenMode = iota
	screenParty
	screenEquip
)

type MemberEquipment struct {
	WeaponID string
	ArmorID  string
}

type PartySelection struct {
	MemberIDs         []string
	EquipmentByMember map[string]MemberEquipment
}

type memberRosterEntry struct {
	Member    save.Member
	Level     int
	BaseStats members.Stats
}

type Model struct {
	state        save.State
	detailed     save.DetailedState
	root         *vfs.Node
	bossProfiles *monsters.BossProfilesConfig
	catalog      *equipment.Catalog
	partyByQuest map[string]PartySelection
	activeScreen screenMode
	nav          NavModel
	party        PartySetupModel
	equip        EquipMemberModel
	pendingQuest string
	pendingParty PartySelection
	membersByID  map[string]memberRosterEntry
	memberList   []memberRosterEntry
}

var newProgram = func(m tea.Model, opts ...tea.ProgramOption) *tea.Program {
	return tea.NewProgram(m, opts...)
}

func Run(state save.State, root *vfs.Node, bossProfiles *monsters.BossProfilesConfig) int {
	model, err := NewModel(state, root, bossProfiles)
	if err != nil {
		fmt.Printf("Failed to initialize UI: %v\n", err)
		return 1
	}

	if _, err := newProgram(model).Run(); err != nil {
		fmt.Printf("Failed to run UI: %v\n", err)
		return 1
	}
	return 0
}

func NewModel(state save.State, root *vfs.Node, bossProfiles *monsters.BossProfilesConfig) (*Model, error) {
	detailed, err := save.LoadDetailed(filepath.Base(state.SaveDir))
	if err != nil {
		return nil, err
	}
	catalog, err := equipment.LoadCatalog()
	if err != nil {
		return nil, err
	}

	memberList := make([]memberRosterEntry, 0, len(detailed.Members))
	memberMap := make(map[string]memberRosterEntry, len(detailed.Members))
	for _, member := range detailed.Members {
		level := members.LevelFromXP(member.XP)
		baseStats, err := members.BaseStats(member.GrowthTypeID, level)
		if err != nil {
			return nil, err
		}

		entry := memberRosterEntry{
			Member:    member,
			Level:     level,
			BaseStats: baseStats,
		}
		memberList = append(memberList, entry)
		memberMap[member.ID] = entry
	}

	model := &Model{
		state:        state,
		detailed:     detailed,
		root:         root,
		bossProfiles: bossProfiles,
		catalog:      catalog,
		partyByQuest: map[string]PartySelection{},
		activeScreen: screenNav,
		nav:          NewNavModel(root),
		membersByID:  memberMap,
		memberList:   memberList,
	}
	return model, nil
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.activeScreen {
	case screenParty:
		return m.updateParty(msg)
	case screenEquip:
		return m.updateEquip(msg)
	default:
		return m.updateNav(msg)
	}
}

func (m *Model) View() string {
	switch m.activeScreen {
	case screenParty:
		return m.party.View()
	case screenEquip:
		return m.equip.View()
	default:
		return m.nav.View()
	}
}

func (m *Model) updateNav(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.nav.width = typed.Width
		m.nav.height = typed.Height
		return m, nil
	case tea.KeyMsg:
		if typed.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		submitted, quit := m.nav.HandleKey(typed)
		if quit {
			return m, tea.Quit
		}
		if submitted == "" {
			return m, nil
		}
		if quit := m.executeCommand(submitted); quit {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *Model) updateParty(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	action := m.party.HandleKey(key)
	switch action {
	case partyActionCancel:
		m.activeScreen = screenNav
		m.pendingQuest = ""
		m.pendingParty = PartySelection{}
		m.nav.AppendLine("Party setup canceled.")
	case partyActionConfirm:
		m.pendingParty.MemberIDs = append([]string(nil), m.party.SelectedIDs()...)
		if m.pendingParty.EquipmentByMember == nil {
			m.pendingParty.EquipmentByMember = map[string]MemberEquipment{}
		}
		for memberID := range m.pendingParty.EquipmentByMember {
			if !contains(m.pendingParty.MemberIDs, memberID) {
				delete(m.pendingParty.EquipmentByMember, memberID)
			}
		}
		m.equip = NewEquipMemberModel(m.pendingQuest, m.pendingParty, m.memberListFromIDs(m.pendingParty.MemberIDs), m.catalog, m.detailed.Inventory)
		m.activeScreen = screenEquip
	}
	return m, nil
}

func (m *Model) updateEquip(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	action := m.equip.HandleKey(key)
	switch action {
	case equipActionBack:
		m.pendingParty = m.equip.Selection()
		m.party = NewPartySetupModel(m.pendingQuest, m.memberList, m.pendingParty.MemberIDs)
		m.activeScreen = screenParty
	case equipActionDone:
		m.pendingParty = m.equip.Selection()
		m.partyByQuest[m.pendingQuest] = clonePartySelection(m.pendingParty)
		m.nav.AppendLine(fmt.Sprintf("Party saved for %s.", m.pendingQuest))
		m.activeScreen = screenNav
		m.pendingQuest = ""
		m.pendingParty = PartySelection{}
	}
	return m, nil
}

func (m *Model) executeCommand(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}

	m.nav.AppendLine("> " + line)
	fields := strings.Fields(line)
	cmd := fields[0]

	switch cmd {
	case "ls":
		target := m.nav.cwd
		if len(fields) == 2 {
			next, err := vfs.Resolve(m.nav.cwd, m.root, fields[1])
			if err != nil {
				m.nav.AppendLine("ls failed: " + err.Error())
				return false
			}
			if next.Type() != vfs.NodeDir {
				m.nav.AppendLine("ls failed: target is not a directory")
				return false
			}
			target = next
		} else if len(fields) > 2 {
			m.nav.AppendLine("ls accepts zero or one path argument")
			return false
		}
		for _, row := range vfs.List(target) {
			m.nav.AppendLine(row)
		}
	case "cd":
		if len(fields) != 2 {
			m.nav.AppendLine("cd requires exactly one path argument")
			return false
		}
		next, err := vfs.Resolve(m.nav.cwd, m.root, fields[1])
		if err != nil {
			m.nav.AppendLine("cd failed: " + err.Error())
			return false
		}
		if next.Type() != vfs.NodeDir {
			m.nav.AppendLine("cd failed: target is not a directory")
			return false
		}
		m.nav.cwd = next
	case "exit":
		return true
	default:
		target, err := vfs.Resolve(m.nav.cwd, m.root, cmd)
		if err != nil {
			m.nav.AppendLine("unknown command or path: " + cmd)
			return false
		}
		if target.Type() == vfs.NodeDir {
			m.nav.AppendLine("cannot execute directory: " + cmd)
			return false
		}

		args := []string{}
		if len(fields) > 1 {
			args = fields[1:]
		}
		return m.executeNode(target, args)
	}
	return false
}

func (m *Model) executeNode(target *vfs.Node, args []string) bool {
	switch target.Name() {
	case "status":
		m.nav.AppendBlock(strings.TrimRight(format.Status(status.FromState(m.state)), "\n"))
	case "exit":
		return true
	case "create_boss":
		m.handleCreateBoss(args)
	case "info":
		m.handleQuestInfo(target)
	case "party":
		questPath, ok := questPathForNode(target.Parent())
		if !ok {
			m.nav.AppendLine("party failed: target is not a quest")
			return false
		}
		selection := clonePartySelection(m.partyByQuest[questPath])
		m.pendingQuest = questPath
		m.pendingParty = selection
		m.party = NewPartySetupModel(questPath, m.memberList, selection.MemberIDs)
		m.activeScreen = screenParty
	case "clear":
		questPath, ok := questPathForNode(target.Parent())
		if !ok {
			m.nav.AppendLine("clear failed: target is not a quest")
			return false
		}
		delete(m.partyByQuest, questPath)
		m.nav.AppendLine("Party selection cleared.")
	default:
		m.nav.AppendLine("unknown view or action: " + target.Name())
	}
	return false
}

func (m *Model) handleQuestInfo(target *vfs.Node) {
	if target.ProfileID == "" || m.bossProfiles == nil {
		m.nav.AppendLine("info: no profile associated")
		return
	}
	_, profile, err := m.bossProfiles.ProfileByID(target.ProfileID)
	if err != nil {
		m.nav.AppendLine("info failed: " + err.Error())
		return
	}
	m.nav.AppendBlock(strings.TrimRight(format.QuestInfo(profile), "\n"))
}

func (m *Model) handleCreateBoss(args []string) {
	if len(args) < 1 || args[0] == "" {
		m.nav.AppendLine("create_boss requires profile_id")
		return
	}
	profileID := args[0]

	var seedOpt *int64
	if len(args) >= 2 {
		var seed int64
		if _, err := fmt.Sscanf(args[1], "%d", &seed); err != nil {
			m.nav.AppendLine("invalid seed: " + args[1])
			return
		}
		seedOpt = &seed
	}

	boss, err := monsters.GenerateBoss(profileID, seedOpt)
	if err != nil {
		m.nav.AppendLine("create_boss failed: " + err.Error())
		return
	}
	m.nav.AppendBlock(strings.TrimRight(format.Boss(boss), "\n"))
}

func (m *Model) memberListFromIDs(ids []string) []memberRosterEntry {
	list := make([]memberRosterEntry, 0, len(ids))
	for _, id := range ids {
		if member, ok := m.membersByID[id]; ok {
			list = append(list, member)
		}
	}
	return list
}

func questPathForNode(node *vfs.Node) (string, bool) {
	if node == nil || node.Type() != vfs.NodeDir {
		return "", false
	}
	path := vfs.DirPath(node)
	if !strings.HasPrefix(path, "/quests/") {
		return "", false
	}
	return path, true
}

func clonePartySelection(selection PartySelection) PartySelection {
	cloned := PartySelection{
		MemberIDs:         append([]string(nil), selection.MemberIDs...),
		EquipmentByMember: map[string]MemberEquipment{},
	}
	for memberID, equip := range selection.EquipmentByMember {
		cloned.EquipmentByMember[memberID] = equip
	}
	return cloned
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

type NavModel struct {
	root        *vfs.Node
	cwd         *vfs.Node
	input       []rune
	cursor      int
	lines       []string
	width       int
	height      int
	suggestions []string
}

func NewNavModel(root *vfs.Node) NavModel {
	return NavModel{root: root, cwd: root}
}

func (m *NavModel) HandleKey(key tea.KeyMsg) (string, bool) {
	switch key.Type {
	case tea.KeyEnter:
		line := string(m.input)
		m.input = nil
		m.cursor = 0
		m.suggestions = nil
		return line, false
	case tea.KeyBackspace, tea.KeyCtrlH:
		if m.cursor > 0 {
			m.input = append(m.input[:m.cursor-1], m.input[m.cursor:]...)
			m.cursor--
		}
		m.refreshSuggestions()
	case tea.KeyLeft:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyRight:
		if m.cursor < len(m.input) {
			m.cursor++
		}
	case tea.KeyHome:
		m.cursor = 0
	case tea.KeyEnd:
		m.cursor = len(m.input)
	case tea.KeySpace:
		m.insertRunes([]rune{' '})
	case tea.KeyTab:
		m.applyCompletion()
	case tea.KeyEsc:
		m.input = nil
		m.cursor = 0
		m.suggestions = nil
	default:
		if key.Type == tea.KeyRunes {
			m.insertRunes(key.Runes)
		}
	}
	return "", false
}

func (m *NavModel) insertRunes(runes []rune) {
	next := append([]rune{}, m.input[:m.cursor]...)
	next = append(next, runes...)
	next = append(next, m.input[m.cursor:]...)
	m.input = next
	m.cursor += len(runes)
	m.refreshSuggestions()
}

func (m *NavModel) refreshSuggestions() {
	m.suggestions = m.completeLine(string(m.input[:m.cursor]))
}

func (m *NavModel) applyCompletion() {
	line := string(m.input[:m.cursor])
	suggestions := m.completeLine(line)
	m.suggestions = suggestions
	if len(suggestions) == 0 {
		return
	}

	tokenStart, token := currentToken(line)
	if len(suggestions) == 1 {
		m.replaceToken(tokenStart, token, suggestions[0])
		return
	}

	common := longestCommonPrefix(suggestions)
	if common != "" && common != token {
		m.replaceToken(tokenStart, token, common)
	}
}

func (m *NavModel) replaceToken(start int, oldToken, replacement string) {
	full := string(m.input)
	end := start + len(oldToken)
	updated := []rune(full[:start] + replacement + full[end:])
	m.input = updated
	m.cursor = start + len([]rune(replacement))
	m.refreshSuggestions()
}

func (m *NavModel) AppendLine(line string) {
	m.lines = append(m.lines, line)
}

func (m *NavModel) AppendBlock(block string) {
	for _, line := range strings.Split(block, "\n") {
		m.lines = append(m.lines, line)
	}
}

func (m NavModel) View() string {
	var b strings.Builder
	b.WriteString("Chronicle of a Clan\n")
	b.WriteString("Path: ")
	b.WriteString(vfs.DirPath(m.cwd))
	b.WriteString("\n\n")
	for _, line := range m.lines {
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("> ")
	b.WriteString(string(m.input))
	if len(m.suggestions) > 0 {
		b.WriteString("\n")
		b.WriteString(strings.Join(m.suggestions, "  "))
	}
	return b.String()
}

func (m NavModel) completeLine(line string) []string {
	fields := strings.Fields(line)
	commands := []string{"ls", "cd", "status", "exit"}

	if len(fields) == 0 {
		return commands
	}

	if len(fields) == 1 && !strings.HasSuffix(line, " ") {
		first := fields[0]
		suggestions := filterPrefix(commands, first)
		suggestions = append(suggestions, vfs.CompletePathSuggestions(m.cwd, m.root, first)...)
		return suggestions
	}

	if fields[0] == "cd" || fields[0] == "ls" {
		token := ""
		if strings.HasSuffix(line, " ") {
			fields = append(fields, "")
		}
		if len(fields) >= 2 {
			token = fields[1]
		}
		return vfs.CompletePathSuggestions(m.cwd, m.root, token)
	}

	return nil
}

func filterPrefix(items []string, prefix string) []string {
	suggestions := make([]string, 0, len(items))
	for _, item := range items {
		if strings.HasPrefix(item, prefix) {
			suggestions = append(suggestions, item)
		}
	}
	return suggestions
}

func currentToken(line string) (int, string) {
	if line == "" {
		return 0, ""
	}
	start := strings.LastIndex(line, " ")
	if start == -1 {
		return 0, line
	}
	return start + 1, line[start+1:]
}

func longestCommonPrefix(items []string) string {
	if len(items) == 0 {
		return ""
	}
	prefix := items[0]
	for _, item := range items[1:] {
		for !strings.HasPrefix(item, prefix) {
			if prefix == "" {
				return ""
			}
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix
}

type partyAction int

const (
	partyActionNone partyAction = iota
	partyActionCancel
	partyActionConfirm
)

type PartySetupModel struct {
	questPath string
	members   []memberRosterEntry
	selected  map[string]bool
	cursor    int
	message   string
}

func NewPartySetupModel(questPath string, membersList []memberRosterEntry, selectedIDs []string) PartySetupModel {
	selected := make(map[string]bool, len(selectedIDs))
	for _, id := range selectedIDs {
		selected[id] = true
	}
	return PartySetupModel{
		questPath: questPath,
		members:   append([]memberRosterEntry(nil), membersList...),
		selected:  selected,
	}
}

func (m *PartySetupModel) HandleKey(key tea.KeyMsg) partyAction {
	switch key.Type {
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown:
		if m.cursor < len(m.members)-1 {
			m.cursor++
		}
	case tea.KeyEnter:
		if len(m.members) == 0 {
			return partyActionNone
		}
		id := m.members[m.cursor].Member.ID
		if m.selected[id] {
			delete(m.selected, id)
			m.message = ""
			return partyActionNone
		}
		if m.SelectedCount() >= 4 {
			m.message = "You can select up to 4 members."
			return partyActionNone
		}
		m.selected[id] = true
		m.message = ""
	case tea.KeyEsc:
		return partyActionCancel
	case tea.KeyRunes:
		if len(key.Runes) == 1 && (key.Runes[0] == 'f' || key.Runes[0] == 'F') {
			if m.SelectedCount() == 0 {
				m.message = "Select at least one member."
				return partyActionNone
			}
			m.message = ""
			return partyActionConfirm
		}
	}
	return partyActionNone
}

func (m PartySetupModel) SelectedCount() int {
	count := 0
	for _, selected := range m.selected {
		if selected {
			count++
		}
	}
	return count
}

func (m PartySetupModel) SelectedIDs() []string {
	ids := make([]string, 0, m.SelectedCount())
	for _, member := range m.members {
		if m.selected[member.Member.ID] {
			ids = append(ids, member.Member.ID)
		}
	}
	return ids
}

func (m PartySetupModel) View() string {
	var b strings.Builder
	b.WriteString("Party Setup - Select Members\n\n")
	for i, member := range m.members {
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}
		check := "[ ]"
		if m.selected[member.Member.ID] {
			check = "[x]"
		}
		b.WriteString(fmt.Sprintf("%s%s %-10s Lv%d\n", prefix, check, member.Member.Name, member.Level))
	}
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Selected: %d/4\n", m.SelectedCount()))
	b.WriteString("Enter: Toggle\n")
	b.WriteString("F: Confirm Members\n")
	b.WriteString("Esc: Cancel\n")
	if m.message != "" {
		b.WriteString("\n")
		b.WriteString(m.message)
		b.WriteString("\n")
	}
	return b.String()
}

type equipFocus int

const (
	focusWeapon equipFocus = iota
	focusArmor
)

type equipAction int

const (
	equipActionNone equipAction = iota
	equipActionBack
	equipActionDone
)

type memberEquipState struct {
	member       memberRosterEntry
	weapons      []equipment.EquipmentOption
	armor        []equipment.EquipmentOption
	weaponCursor int
	armorCursor  int
	selected     MemberEquipment
}

type EquipMemberModel struct {
	questPath    string
	members      []memberEquipState
	currentIndex int
	focus        equipFocus
}

func NewEquipMemberModel(questPath string, selection PartySelection, membersList []memberRosterEntry, catalog *equipment.Catalog, inventory save.Inventory) EquipMemberModel {
	model := EquipMemberModel{
		questPath: questPath,
		members:   make([]memberEquipState, 0, len(membersList)),
	}
	for _, member := range membersList {
		weapons, armor := equipment.CandidatesForMember(catalog, inventory, member.Level, member.BaseStats)
		state := memberEquipState{
			member:   member,
			weapons:  weapons,
			armor:    armor,
			selected: selection.EquipmentByMember[member.Member.ID],
		}
		state.weaponCursor = selectedCursor(weapons, state.selected.WeaponID)
		state.armorCursor = selectedCursor(armor, state.selected.ArmorID)
		if state.selected.WeaponID == "" && len(weapons) > 0 {
			state.selected.WeaponID = weapons[state.weaponCursor].ID
		}
		if state.selected.ArmorID == "" && len(armor) > 0 {
			state.selected.ArmorID = armor[state.armorCursor].ID
		}
		model.members = append(model.members, state)
	}
	return model
}

func (m *EquipMemberModel) HandleKey(key tea.KeyMsg) equipAction {
	if len(m.members) == 0 {
		if key.Type == tea.KeyEsc {
			return equipActionBack
		}
		return equipActionDone
	}

	current := &m.members[m.currentIndex]
	switch key.Type {
	case tea.KeyUp:
		if m.focus == focusWeapon && current.weaponCursor > 0 {
			current.weaponCursor--
		}
		if m.focus == focusArmor && current.armorCursor > 0 {
			current.armorCursor--
		}
	case tea.KeyDown:
		if m.focus == focusWeapon && current.weaponCursor < len(current.weapons)-1 {
			current.weaponCursor++
		}
		if m.focus == focusArmor && current.armorCursor < len(current.armor)-1 {
			current.armorCursor++
		}
	case tea.KeyTab:
		if m.focus == focusWeapon {
			m.focus = focusArmor
		} else {
			m.focus = focusWeapon
		}
	case tea.KeyEnter:
		if m.focus == focusWeapon && len(current.weapons) > 0 {
			current.selected.WeaponID = current.weapons[current.weaponCursor].ID
		}
		if m.focus == focusArmor && len(current.armor) > 0 {
			current.selected.ArmorID = current.armor[current.armorCursor].ID
		}
	case tea.KeyEsc:
		return equipActionBack
	case tea.KeyRunes:
		if len(key.Runes) == 1 && (key.Runes[0] == 'n' || key.Runes[0] == 'N') {
			if m.currentIndex == len(m.members)-1 {
				return equipActionDone
			}
			m.currentIndex++
			m.focus = focusWeapon
		}
	}
	return equipActionNone
}

func (m EquipMemberModel) Selection() PartySelection {
	selection := PartySelection{
		MemberIDs:         make([]string, 0, len(m.members)),
		EquipmentByMember: map[string]MemberEquipment{},
	}
	for _, member := range m.members {
		selection.MemberIDs = append(selection.MemberIDs, member.member.Member.ID)
		selection.EquipmentByMember[member.member.Member.ID] = member.selected
	}
	return selection
}

func (m EquipMemberModel) View() string {
	if len(m.members) == 0 {
		return "Equip Member\n\nNo members selected.\n"
	}

	current := m.members[m.currentIndex]
	selectedWeapon := current.selectedOption(focusWeapon)
	selectedArmor := current.selectedOption(focusArmor)
	currentStats := current.member.BaseStats
	if selectedWeapon != nil {
		currentStats = currentStats.Add(selectedWeapon.StatModifiers)
	}
	if selectedArmor != nil {
		currentStats = currentStats.Add(selectedArmor.StatModifiers)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Equip Member - %s\n\n", current.member.Member.Name))
	b.WriteString(sectionTitle("Weapon", m.focus == focusWeapon))
	b.WriteString(renderOptions(current.weapons, current.weaponCursor, current.selected.WeaponID))
	b.WriteString("\n")
	b.WriteString(sectionTitle("Armor", m.focus == focusArmor))
	b.WriteString(renderOptions(current.armor, current.armorCursor, current.selected.ArmorID))
	b.WriteString("\nCurrent Stats\n")
	b.WriteString(fmt.Sprintf("Might:    %d\n", currentStats.Might))
	b.WriteString(fmt.Sprintf("Mastery:  %d\n", currentStats.Mastery))
	b.WriteString(fmt.Sprintf("Tactics:  %d\n", currentStats.Tactics))
	b.WriteString(fmt.Sprintf("Survival: %d\n", currentStats.Survival))
	b.WriteString("\nPreview\n")
	b.WriteString("Weapon: ")
	b.WriteString(formatPreview(current.previewOption(focusWeapon)))
	b.WriteString("\nArmor:  ")
	b.WriteString(formatPreview(current.previewOption(focusArmor)))
	b.WriteString("\n\nEnter: Confirm\nTab:   Switch weapon/armor\nN:     Next member\nEsc:   Back\n")
	return b.String()
}

func (m memberEquipState) selectedOption(focus equipFocus) *equipment.EquipmentOption {
	id := m.selected.WeaponID
	options := m.weapons
	if focus == focusArmor {
		id = m.selected.ArmorID
		options = m.armor
	}
	for i := range options {
		if options[i].ID == id {
			return &options[i]
		}
	}
	return nil
}

func (m memberEquipState) previewOption(focus equipFocus) *equipment.EquipmentOption {
	options := m.weapons
	cursor := m.weaponCursor
	if focus == focusArmor {
		options = m.armor
		cursor = m.armorCursor
	}
	if len(options) == 0 || cursor >= len(options) {
		return nil
	}
	return &options[cursor]
}

func renderOptions(options []equipment.EquipmentOption, cursor int, selectedID string) string {
	if len(options) == 0 {
		return "  (none)\n"
	}
	var b strings.Builder
	for i, option := range options {
		prefix := "  "
		if i == cursor {
			prefix = "> "
		}
		marker := " "
		if option.ID == selectedID {
			marker = "*"
		}
		b.WriteString(fmt.Sprintf("%s%s %s  %s\n", prefix, marker, option.ID, option.Name))
	}
	return b.String()
}

func sectionTitle(title string, focused bool) string {
	if focused {
		return title + " [focus]\n"
	}
	return title + "\n"
}

func formatPreview(option *equipment.EquipmentOption) string {
	if option == nil {
		return "(none)"
	}

	parts := make([]string, 0, 5)
	if option.StatModifiers.Might != 0 {
		parts = append(parts, fmt.Sprintf("+Might %d", option.StatModifiers.Might))
	}
	if option.StatModifiers.Mastery != 0 {
		parts = append(parts, fmt.Sprintf("+Mastery %d", option.StatModifiers.Mastery))
	}
	if option.StatModifiers.Tactics != 0 {
		parts = append(parts, fmt.Sprintf("+Tactics %d", option.StatModifiers.Tactics))
	}
	if option.StatModifiers.Survival != 0 {
		parts = append(parts, fmt.Sprintf("+Survival %d", option.StatModifiers.Survival))
	}
	if option.Prot > 0 {
		parts = append(parts, fmt.Sprintf("PROT %d", option.Prot))
	}
	if len(parts) == 0 {
		return "No bonus"
	}
	return strings.Join(parts, ", ")
}

func selectedCursor(options []equipment.EquipmentOption, selectedID string) int {
	for i, option := range options {
		if option.ID == selectedID {
			return i
		}
	}
	return 0
}
