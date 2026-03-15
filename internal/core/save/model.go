package save

type State struct {
	SaveDir             string
	SaveVersion         int
	ClanName            string
	CurrentDay          int
	Gold                int
	Fame                int
	MembersCount        int
	ActiveQuestsCount   int
	WeaponsCount        int
	ArmorCount          int
	InProgressCount     int
	ChronicleEntryCount int
	HasChronicle        bool
	KeyQuestCurrentOrder int
}

const SupportedSaveVersion = 1
