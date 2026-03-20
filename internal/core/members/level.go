package members

import (
	"encoding/json"
	"os"
	"sync"

	"chronicle-of-a-clan/internal/core/datafiles"
)

type levelThreshold struct {
	Level int `json:"level"`
	MinXP int `json:"min_xp"`
}

type xpConfig struct {
	Thresholds []levelThreshold `json:"member_level_thresholds"`
}

var (
	xpOnce       sync.Once
	xpThresholds []levelThreshold
	xpErr        error
)

func LevelFromXP(xp int) int {
	thresholds, err := loadXPThresholds()
	if err != nil || len(thresholds) == 0 {
		return 1
	}

	level := thresholds[0].Level
	for _, threshold := range thresholds {
		if xp < threshold.MinXP {
			break
		}
		level = threshold.Level
	}
	return level
}

func loadXPThresholds() ([]levelThreshold, error) {
	xpOnce.Do(func() {
		b, err := os.ReadFile(datafiles.Path("data/combat/xp.json"))
		if err != nil {
			xpErr = err
			return
		}

		var cfg xpConfig
		if err := json.Unmarshal(b, &cfg); err != nil {
			xpErr = err
			return
		}
		xpThresholds = cfg.Thresholds
	})
	return xpThresholds, xpErr
}
