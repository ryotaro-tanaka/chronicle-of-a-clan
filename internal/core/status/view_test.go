package status

import (
	"testing"

	"chronicle-of-a-clan/internal/core/save"
)

func TestFromState(t *testing.T) {
	v := FromState(save.State{ClanName: "Chronicle Clan", CurrentDay: 7, Gold: 123, Fame: 9, MembersCount: 2, ActiveQuestsCount: 3})

	if v.ClanName != "Chronicle Clan" || v.Day != 7 || v.Gold != 123 || v.Fame != 9 || v.MembersCount != 2 || v.ActiveQuestsCount != 3 {
		t.Fatalf("unexpected view: %+v", v)
	}
}
