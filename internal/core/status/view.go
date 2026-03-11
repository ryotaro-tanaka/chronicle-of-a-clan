package status

import "chronicle-of-a-clan/internal/core/save"

type View struct {
	ClanName          string
	Day               int
	Gold              int
	Fame              int
	MembersCount      int
	ActiveQuestsCount int
}

func FromState(s save.State) View {
	return View{
		ClanName:          s.ClanName,
		Day:               s.CurrentDay,
		Gold:              s.Gold,
		Fame:              s.Fame,
		MembersCount:      s.MembersCount,
		ActiveQuestsCount: s.ActiveQuestsCount,
	}
}
