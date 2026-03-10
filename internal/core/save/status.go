package save

import "fmt"

func FormatStatus(s State) string {
	return fmt.Sprintf(
		"Clan: %s   Day: %d\nGold: %d   Fame: %d\nMembers: %d   ActiveQuests: %d\n",
		s.ClanName,
		s.CurrentDay,
		s.Gold,
		s.Fame,
		s.MembersCount,
		s.ActiveQuestsCount,
	)
}
