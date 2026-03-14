package format

import (
	"fmt"

	"chronicle-of-a-clan/internal/core/monsters"
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

func Boss(b monsters.Boss) string {
	return fmt.Sprintf(
		"Boss: profile_id=%s name=%q description=%q region=%s level=%d\nStats: Power=%d Guard=%d Evasion=%d Cunning=%d\n",
		b.ProfileID,
		b.Name,
		b.Description,
		b.Region,
		b.Level,
		b.Stats.Power,
		b.Stats.Guard,
		b.Stats.Evasion,
		b.Stats.Cunning,
	)
}
