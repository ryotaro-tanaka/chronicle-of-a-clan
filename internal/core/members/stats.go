package members

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Stats struct {
	Might    int
	Mastery  int
	Tactics  int
	Survival int
}

func (s Stats) Meets(req Stats) bool {
	return s.Might >= req.Might && s.Mastery >= req.Mastery && s.Tactics >= req.Tactics && s.Survival >= req.Survival
}

type growthTypeFile struct {
	GrowthTypes []struct {
		ID             string `json:"id"`
		PerLevelGrowth struct {
			Might    int `json:"might"`
			Mastery  int `json:"mastery"`
			Tactics  int `json:"tactics"`
			Survival int `json:"survival"`
		} `json:"per_level_growth"`
	} `json:"growth_types"`
}

var (
	growthOnce sync.Once
	growthByID map[string]Stats
)

func StatsFor(growthTypeID string, level int) (Stats, error) {
	growthOnce.Do(loadGrowthTypes)
	growth, ok := growthByID[growthTypeID]
	if !ok {
		return Stats{}, fmt.Errorf("unknown growth_type_id: %s", growthTypeID)
	}
	if level < 1 {
		level = 1
	}
	base := Stats{Might: 180, Mastery: 180, Tactics: 180, Survival: 180}
	if level == 1 {
		return base, nil
	}
	gains := level - 1
	return Stats{
		Might:    base.Might + growth.Might*gains,
		Mastery:  base.Mastery + growth.Mastery*gains,
		Tactics:  base.Tactics + growth.Tactics*gains,
		Survival: base.Survival + growth.Survival*gains,
	}, nil
}

func loadGrowthTypes() {
	growthByID = map[string]Stats{}
	b, err := os.ReadFile("data/combat/member_growth_types.json")
	if err != nil {
		return
	}
	var cfg growthTypeFile
	if err := json.Unmarshal(b, &cfg); err != nil {
		return
	}
	for _, gt := range cfg.GrowthTypes {
		growthByID[gt.ID] = Stats{
			Might:    gt.PerLevelGrowth.Might,
			Mastery:  gt.PerLevelGrowth.Mastery,
			Tactics:  gt.PerLevelGrowth.Tactics,
			Survival: gt.PerLevelGrowth.Survival,
		}
	}
}
