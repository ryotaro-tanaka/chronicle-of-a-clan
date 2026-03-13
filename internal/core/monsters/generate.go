package monsters

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// BossStats holds the four combat stats for a boss.
type BossStats struct {
	Power   int
	Guard   int
	Evasion int
	Cunning int
}

// Boss is the fully generated boss monster.
type Boss struct {
	Region       string
	QuestLevel   int
	MonsterLevel int
	Rank         int
	Overall      int
	ProfileID    string
	Name         string
	Stats        BossStats
}

// GenerateBoss creates one boss for the given region and quest level.
// seedOpt can be nil to use a time-based seed.
func GenerateBoss(region string, questLevel int, seedOpt *int64) (Boss, error) {
	profilesCfg, err := LoadBossProfiles()
	if err != nil {
		return Boss{}, err
	}
	levelsCfg, err := LoadLevels()
	if err != nil {
		return Boss{}, err
	}

	regionCfg, ok := profilesCfg.Regions[region]
	if !ok {
		return Boss{}, fmt.Errorf("unknown region: %s", region)
	}
	if len(regionCfg.Profiles) == 0 {
		return Boss{}, fmt.Errorf("region %s has no profiles", region)
	}

	qlRange, err := levelsCfg.QuestLevelRangeFor(questLevel)
	if err != nil {
		return Boss{}, err
	}

	seed := time.Now().UnixNano()
	if seedOpt != nil {
		seed = *seedOpt
	}
	rnd := rand.New(rand.NewSource(seed))

	monsterLevel := randomIntInRange(rnd, qlRange.MonsterLevelMin, qlRange.MonsterLevelMax)
	budget := levelsCfg.MonsterLevelBudgetModel.BudgetForMonsterLevel(monsterLevel)

	profile := weightedRandomProfile(rnd, regionCfg.Profiles)
	ratios := baseRatiosWithVariation(rnd, regionCfg, profile)

	stats := scaleStats(budget, ratios)
	overall := levelsCfg.OverallFor(qlRange, monsterLevel)

	return Boss{
		Region:       region,
		QuestLevel:   questLevel,
		MonsterLevel: monsterLevel,
		Rank:         profile.Rank,
		Overall:      overall,
		ProfileID:    profile.ID,
		Name:         profile.Name,
		Stats:        stats,
	}, nil
}

func randomIntInRange(rnd *rand.Rand, min, max int) int {
	if max <= min {
		return min
	}
	return rnd.Intn(max-min+1) + min
}

func weightedRandomProfile(rnd *rand.Rand, profiles []RawProfile) RawProfile {
	var total float64
	for _, p := range profiles {
		total += p.Weight
	}
	if total <= 0 {
		// Fallback to uniform choice
		return profiles[rnd.Intn(len(profiles))]
	}
	r := rnd.Float64() * total
	var acc float64
	for _, p := range profiles {
		acc += p.Weight
		if r <= acc {
			return p
		}
	}
	return profiles[len(profiles)-1]
}

func baseRatiosWithVariation(rnd *rand.Rand, region RawRegion, profile RawProfile) map[string]float64 {
	// Base ratios from profile.stats
	allStats := []string{"power", "guard", "evasion", "cunning"}
	ratios := make(map[string]float64, len(allStats))

	var focusSum float64
	focusSet := make(map[string]struct{}, len(profile.Stats))
	for _, sf := range profile.Stats {
		ratios[sf.Stat] += sf.Ratio
		focusSum += sf.Ratio
		focusSet[sf.Stat] = struct{}{}
	}

	// Distribute remaining evenly over non-focus stats
	remaining := 1.0 - focusSum
	if remaining < 0 {
		remaining = 0
	}
	var nonFocus []string
	for _, s := range allStats {
		if _, ok := focusSet[s]; !ok {
			nonFocus = append(nonFocus, s)
		}
	}
	if len(nonFocus) > 0 && remaining > 0 {
		share := remaining / float64(len(nonFocus))
		for _, s := range nonFocus {
			ratios[s] += share
		}
	}

	// Apply variation based on number of focused stats
	variation := variationFor(region, len(profile.Stats))
	if variation > 0 {
		var sum float64
		for _, s := range allStats {
			base := ratios[s]
			if base <= 0 {
				continue
			}
			noise := (rnd.Float64()*2 - 1) * variation
			val := base * (1 + noise)
			if val < 0 {
				val = 0
			}
			ratios[s] = val
			sum += val
		}
		if sum > 0 {
			for _, s := range allStats {
				ratios[s] /= sum
			}
		}
	}

	return ratios
}

func variationFor(region RawRegion, focusedCount int) float64 {
	for _, v := range region.Variation {
		if v.FocusedStatsCount == focusedCount {
			return v.Variation
		}
	}
	return 0
}

func scaleStats(budget int, ratios map[string]float64) BossStats {
	if budget <= 0 {
		return BossStats{}
	}
	power := int(math.Round(float64(budget) * ratios["power"]))
	guard := int(math.Round(float64(budget) * ratios["guard"]))
	evasion := int(math.Round(float64(budget) * ratios["evasion"]))
	cunning := int(math.Round(float64(budget) * ratios["cunning"]))
	return BossStats{
		Power:   power,
		Guard:   guard,
		Evasion: evasion,
		Cunning: cunning,
	}
}

