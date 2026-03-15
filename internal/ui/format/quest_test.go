package format

import (
	"strings"
	"testing"

	"chronicle-of-a-clan/internal/core/monsters"
)

func TestQuestInfo(t *testing.T) {
	p := monsters.RawProfile{
		Name:        "Ambushjaw Gator",
		Description: "A root-side ambusher.",
		LevelMin:    1,
		LevelMax:    5,
		Stats: []monsters.RawStatFocus{
			{Stat: "cunning", Ratio: 0.36},
			{Stat: "evasion", Ratio: 0.3},
		},
	}
	got := QuestInfo(p)
	for _, want := range []string{
		"Name: Ambushjaw Gator",
		"Lv: 1-5",
		"Specialties: Cunning, Evasion",
		"Reward:",
		"Description: A root-side ambusher.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("QuestInfo output missing %q in:\n%s", want, got)
		}
	}
}
