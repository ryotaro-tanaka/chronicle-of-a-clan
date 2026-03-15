package format

import (
	"fmt"
	"strings"
	"unicode"

	"chronicle-of-a-clan/internal/core/monsters"
)

// QuestInfo formats a boss profile for the quest info view: Name, Lv, Specialties, Reward, Description.
func QuestInfo(p monsters.RawProfile) string {
	specs := make([]string, len(p.Stats))
	for i, s := range p.Stats {
		specs[i] = titleCase(s.Stat)
	}
	specsStr := strings.Join(specs, ", ")
	return fmt.Sprintf(
		"Name: %s\nLv: %d-%d\nSpecialties: %s\nReward:\nDescription: %s\n",
		p.Name,
		p.LevelMin,
		p.LevelMax,
		specsStr,
		p.Description,
	)
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	for i := 1; i < len(r); i++ {
		r[i] = unicode.ToLower(r[i])
	}
	return string(r)
}
