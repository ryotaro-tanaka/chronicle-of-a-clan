package monsters

import (
	"encoding/json"
	"fmt"
	"os"
)

// JSON structures for boss_profiles.json

type RawVariation struct {
	FocusedStatsCount int     `json:"focused_stats_count"`
	Variation         float64 `json:"variation"`
}

type RawStatFocus struct {
	Stat  string  `json:"stat"`
	Ratio float64 `json:"ratio"`
}

type RawProfile struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Rank   int            `json:"rank"`
	Stats  []RawStatFocus `json:"stats"`
	Weight float64        `json:"weight"`
}

type RawRegion struct {
	Variation []RawVariation `json:"variation"`
	Profiles  []RawProfile   `json:"profiles"`
}

type BossProfilesConfig struct {
	Regions map[string]RawRegion `json:"regions"`
}

const bossProfilesPath = "data/boss_profiles.json"

// LoadBossProfiles loads boss profile configuration from the default JSON file.
func LoadBossProfiles() (*BossProfilesConfig, error) {
	f, err := os.Open(bossProfilesPath)
	if err != nil {
		return nil, fmt.Errorf("open boss profiles: %w", err)
	}
	defer f.Close()

	var cfg BossProfilesConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode boss profiles: %w", err)
	}

	// Minimal validation for MVP.
	for regionID, region := range cfg.Regions {
		if len(region.Profiles) == 0 {
			return nil, fmt.Errorf("region %q has no profiles", regionID)
		}
		var weightSum float64
		for _, p := range region.Profiles {
			if p.ID == "" {
				return nil, fmt.Errorf("region %q has profile with empty id", regionID)
			}
			if p.Weight < 0 {
				return nil, fmt.Errorf("region %q profile %q has negative weight", regionID, p.ID)
			}
			var ratioSum float64
			for _, sf := range p.Stats {
				if sf.Ratio < 0 {
					return nil, fmt.Errorf("region %q profile %q has negative ratio", regionID, p.ID)
				}
				ratioSum += sf.Ratio
			}
			if ratioSum > 1.0+1e-9 {
				return nil, fmt.Errorf("region %q profile %q has stats ratio sum > 1.0", regionID, p.ID)
			}
			weightSum += p.Weight
		}
		if weightSum <= 0 {
			return nil, fmt.Errorf("region %q has non-positive total weight", regionID)
		}
		// variation entries are optional but must not duplicate focused_stats_count
		seen := make(map[int]struct{}, len(region.Variation))
		for _, v := range region.Variation {
			if _, ok := seen[v.FocusedStatsCount]; ok {
				return nil, fmt.Errorf("region %q has duplicate variation for focused_stats_count=%d", regionID, v.FocusedStatsCount)
			}
			seen[v.FocusedStatsCount] = struct{}{}
		}
	}

	return &cfg, nil
}

