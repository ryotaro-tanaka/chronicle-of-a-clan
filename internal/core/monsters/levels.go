package monsters

import (
	"encoding/json"
	"fmt"
	"os"
)

type QuestLevelRange struct {
	ID              int `json:"id"`
	MonsterLevelMin int `json:"monster_level_min"`
	MonsterLevelMax int `json:"monster_level_max"`
}

type LevelSegment struct {
	From             int `json:"from"`
	To               int `json:"to"`
	PerLevelIncrease int `json:"per_level_increase"`
}

type LevelBudgetModel struct {
	BaseBudgetAtLevel1 int            `json:"base_budget_at_level_1"`
	Segments           []LevelSegment `json:"segments"`
}

type OverallRatingModel struct {
	Type  string `json:"type"`
	Steps int    `json:"steps"`
}

type LevelsConfig struct {
	QuestLevels             []QuestLevelRange `json:"quest_levels"`
	MonsterLevelBudgetModel LevelBudgetModel  `json:"monster_level_budget_model"`
	MemberLevelBudgetModel  LevelBudgetModel  `json:"member_level_budget_model"`
	OverallRating           OverallRatingModel `json:"overall_rating"`
}

const levelsPath = "data/levels.json"

// LoadLevels loads quest/monster level related configuration from JSON.
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

	for _, ql := range cfg.QuestLevels {
		if ql.MonsterLevelMin > ql.MonsterLevelMax {
			return nil, fmt.Errorf("quest level %d has min > max", ql.ID)
		}
	}

	return &cfg, nil
}

// QuestLevelRangeFor returns the range for a given quest level id.
func (c *LevelsConfig) QuestLevelRangeFor(id int) (QuestLevelRange, error) {
	for _, ql := range c.QuestLevels {
		if ql.ID == id {
			return ql, nil
		}
	}
	return QuestLevelRange{}, fmt.Errorf("unknown quest level: %d", id)
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

// OverallFor returns the overall rating (1..steps) for a monster level within a quest level range.
func (c *LevelsConfig) OverallFor(ql QuestLevelRange, monsterLevel int) int {
	steps := c.OverallRating.Steps
	if steps <= 0 {
		steps = 5
	}
	min := ql.MonsterLevelMin
	max := ql.MonsterLevelMax
	if max <= min {
		return 1
	}
	if monsterLevel <= min {
		return 1
	}
	if monsterLevel >= max {
		return steps
	}
	span := float64(max - min)
	pos := float64(monsterLevel-min) / span
	rating := int(pos*float64(steps)) + 1
	if rating < 1 {
		rating = 1
	}
	if rating > steps {
		rating = steps
	}
	return rating
}

