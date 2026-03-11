package format

import (
	"fmt"

	"chronicle-of-a-clan/internal/core/status"
)

func Status(v status.View) string {
	return fmt.Sprintf(
		"Clan: %s   Day: %d\nGold: %d   Fame: %d\nMembers: %d   ActiveQuests: %d\n",
		v.ClanName,
		v.Day,
		v.Gold,
		v.Fame,
		v.MembersCount,
		v.ActiveQuestsCount,
	)
}
