package save

type State struct {
	SaveDir              string
	SaveVersion          int
	ClanName             string
	CurrentDay           int
	Gold                 int
	Fame                 int
	MembersCount         int
	ActiveQuestsCount    int
	WeaponsCount         int
	ArmorCount           int
	InProgressCount      int
	ChronicleEntryCount  int
	HasChronicle         bool
	KeyQuestCurrentOrder int
}

type DetailedState struct {
	SaveDir     string
	SaveVersion int
	Clan        Clan
	Members     []Member
	Inventory   Inventory
}

type Clan struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Day  int    `json:"day"`
	Gold int    `json:"gold"`
	Fame int    `json:"fame"`
}

type Member struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	GrowthTypeID string `json:"growth_type_id"`
	XP           int    `json:"xp"`
	JoinedDay    int    `json:"joined_day"`
}

type Inventory struct {
	Weapons []InventoryEntry `json:"weapons"`
	Armor   []InventoryEntry `json:"armor"`
}

type InventoryEntry struct {
	ID    string `json:"id"`
	Count int    `json:"count"`
}

const SupportedSaveVersion = 1
