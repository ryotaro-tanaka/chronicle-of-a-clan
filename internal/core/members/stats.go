package members

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"chronicle-of-a-clan/internal/core/datafiles"
)

const levelOneBaseStat = 180

type Stats struct {
	Might    int `json:"might"`
	Mastery  int `json:"mastery"`
	Tactics  int `json:"tactics"`
	Survival int `json:"survival"`
}

type growthType struct {
	ID             string `json:"id"`
	PerLevelGrowth Stats  `json:"per_level_growth"`
}

type growthConfig struct {
	GrowthTypes []growthType `json:"growth_types"`
}

var (
	growthOnce sync.Once
	growthByID map[string]growthType
	growthsErr error
)

func BaseStats(growthTypeID string, level int) (Stats, error) {
	growths, err := loadGrowthTypes()
	if err != nil {
		return Stats{}, err
	}
	growth, ok := growths[growthTypeID]
	if !ok {
		return Stats{}, fmt.Errorf("unknown growth type: %s", growthTypeID)
	}
	if level < 1 {
		level = 1
	}

	levelsGained := level - 1
	return Stats{
		Might:    levelOneBaseStat + (growth.PerLevelGrowth.Might * levelsGained),
		Mastery:  levelOneBaseStat + (growth.PerLevelGrowth.Mastery * levelsGained),
		Tactics:  levelOneBaseStat + (growth.PerLevelGrowth.Tactics * levelsGained),
		Survival: levelOneBaseStat + (growth.PerLevelGrowth.Survival * levelsGained),
	}, nil
}

func (s Stats) Add(other Stats) Stats {
	return Stats{
		Might:    s.Might + other.Might,
		Mastery:  s.Mastery + other.Mastery,
		Tactics:  s.Tactics + other.Tactics,
		Survival: s.Survival + other.Survival,
	}
}

func (s Stats) Meets(required Stats) bool {
	return s.Might >= required.Might &&
		s.Mastery >= required.Mastery &&
		s.Tactics >= required.Tactics &&
		s.Survival >= required.Survival
}

func loadGrowthTypes() (map[string]growthType, error) {
	growthOnce.Do(func() {
		b, err := os.ReadFile(datafiles.Path("data/combat/member_growth_types.json"))
		if err != nil {
			growthsErr = err
			return
		}

		var cfg growthConfig
		if err := json.Unmarshal(b, &cfg); err != nil {
			growthsErr = err
			return
		}

		growthByID = make(map[string]growthType, len(cfg.GrowthTypes))
		for _, growth := range cfg.GrowthTypes {
			growthByID[growth.ID] = growth
		}
	})
	return growthByID, growthsErr
}
