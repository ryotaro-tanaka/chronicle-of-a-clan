package app

import (
	"chronicle-of-a-clan/internal/core/members"
	"chronicle-of-a-clan/internal/core/save"
)

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
