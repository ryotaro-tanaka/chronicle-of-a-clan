package save

type Member struct {
	ID           string
	Name         string
	GrowthTypeID string
	XP           int
}

type InventoryItem struct {
	ID    string
	Count int
}

type Inventory struct {
	Weapons []InventoryItem
	Armor   []InventoryItem
}

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
	Members              []Member
	Inventory            Inventory
}

const SupportedSaveVersion = 1
