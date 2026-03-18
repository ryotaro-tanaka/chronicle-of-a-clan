package members

import (
	"encoding/json"
	"os"
	"sync"
)

type xpThresholdFile struct {
	MemberLevelThresholds []struct {
		Level int `json:"level"`
		MinXP int `json:"min_xp"`
	} `json:"member_level_thresholds"`
}

var (
	xpOnce      sync.Once
	xpThreshold []struct {
		Level int
		MinXP int
	}
)

func LevelFromXP(xp int) int {
	xpOnce.Do(loadXPThresholds)
	level := 1
	for _, t := range xpThreshold {
		if xp >= t.MinXP {
			level = t.Level
		}
	}
	return level
}

func loadXPThresholds() {
	b, err := os.ReadFile("data/combat/xp.json")
	if err != nil {
		xpThreshold = []struct {
			Level int
			MinXP int
		}{{Level: 1, MinXP: 0}}
		return
	}
	var cfg xpThresholdFile
	if err := json.Unmarshal(b, &cfg); err != nil || len(cfg.MemberLevelThresholds) == 0 {
		xpThreshold = []struct {
			Level int
			MinXP int
		}{{Level: 1, MinXP: 0}}
		return
	}
	xpThreshold = make([]struct {
		Level int
		MinXP int
	}, 0, len(cfg.MemberLevelThresholds))
	for _, t := range cfg.MemberLevelThresholds {
		xpThreshold = append(xpThreshold, struct {
			Level int
			MinXP int
		}{Level: t.Level, MinXP: t.MinXP})
	}
}
