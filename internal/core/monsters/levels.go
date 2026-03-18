package monsters

import (
	"encoding/json"
	"fmt"
	"os"
)

type LevelSegment struct {
	From             int `json:"from"`
	To               int `json:"to"`
	PerLevelIncrease int `json:"per_level_increase"`
}

type LevelBudgetModel struct {
	BaseBudgetAtLevel1 int            `json:"base_budget_at_level_1"`
	Segments           []LevelSegment `json:"segments"`
}

type LevelsConfig struct {
	MonsterLevelBudgetModel LevelBudgetModel `json:"monster_level_budget_model"`
	MemberLevelBudgetModel  LevelBudgetModel `json:"member_level_budget_model"`
}

const levelsPath = "data/combat/levels.json"

// LoadLevels loads level budget configuration from JSON.
func LoadLevels() (*LevelsConfig, error) {
	f, err := os.Open(levelsPath)
	if err != nil {
		return nil, fmt.Errorf("open levels: %w", err)
	}
	defer f.Close()

	var cfg LevelsConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode levels: %w", err)
	}

	return &cfg, nil
}

// BudgetForMonsterLevel returns the total stat budget for the given monster level.
func (m LevelBudgetModel) BudgetForMonsterLevel(level int) int {
	if level <= 1 {
		return m.BaseBudgetAtLevel1
	}
	budget := m.BaseBudgetAtLevel1
	for lvl := 2; lvl <= level; lvl++ {
		for _, seg := range m.Segments {
			if lvl >= seg.From && lvl <= seg.To {
				budget += seg.PerLevelIncrease
				break
			}
		}
	}
	return budget
}
