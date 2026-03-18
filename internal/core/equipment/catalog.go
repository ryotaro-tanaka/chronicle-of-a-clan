package equipment

import (
	"encoding/json"
	"os"
	"sync"

	"chronicle-of-a-clan/internal/core/members"
	"chronicle-of-a-clan/internal/core/save"
)

type Item struct {
	ID            string
	Name          string
	LevelMin      int
	LevelMax      int
	RequiredStats members.Stats
	Modifiers     members.Stats
	Prot          int
}

type fileFormat struct {
	Weapons []struct {
		ID            string    `json:"id"`
		Name          string    `json:"name"`
		LevelMin      int       `json:"level_min"`
		LevelMax      int       `json:"level_max"`
		RequiredStats statsJSON `json:"required_stats"`
		Modifiers     statsJSON `json:"stat_modifiers"`
	} `json:"weapons"`
	Armor []struct {
		ID            string    `json:"id"`
		Name          string    `json:"name"`
		LevelMin      int       `json:"level_min"`
		LevelMax      int       `json:"level_max"`
		RequiredStats statsJSON `json:"required_stats"`
		Modifiers     statsJSON `json:"stat_modifiers"`
		Prot          int       `json:"prot"`
	} `json:"armor"`
}

type statsJSON struct {
	Might    int `json:"might"`
	Mastery  int `json:"mastery"`
	Tactics  int `json:"tactics"`
	Survival int `json:"survival"`
}

var (
	once           sync.Once
	rentalWeapons  []Item
	rentalArmor    []Item
	craftedWeapons map[string]Item
	craftedArmor   map[string]Item
)

func EligibleWeapons(inv save.Inventory, level int, stats members.Stats) []Item {
	once.Do(load)
	out := make([]Item, 0)
	for _, stack := range inv.Weapons {
		if stack.Count <= 0 {
			continue
		}
		item, ok := craftedWeapons[stack.ID]
		if !ok {
			continue
		}
		if !stats.Meets(item.RequiredStats) {
			continue
		}
		out = append(out, item)
	}
	for _, item := range rentalWeapons {
		if level < item.LevelMin || level > item.LevelMax {
			continue
		}
		if !stats.Meets(item.RequiredStats) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func EligibleArmor(inv save.Inventory, level int, stats members.Stats) []Item {
	once.Do(load)
	out := make([]Item, 0)
	for _, stack := range inv.Armor {
		if stack.Count <= 0 {
			continue
		}
		item, ok := craftedArmor[stack.ID]
		if !ok {
			continue
		}
		if !stats.Meets(item.RequiredStats) {
			continue
		}
		out = append(out, item)
	}
	for _, item := range rentalArmor {
		if level < item.LevelMin || level > item.LevelMax {
			continue
		}
		if !stats.Meets(item.RequiredStats) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func load() {
	craftedWeapons = map[string]Item{}
	craftedArmor = map[string]Item{}
	rentalWeapons = loadFile("data/items/equipments/rental_equipment.json", true)
	rentalArmor = loadFile("data/items/equipments/rental_equipment.json", false)
	for _, item := range loadFile("data/items/equipments/crafted_equipment.json", true) {
		craftedWeapons[item.ID] = item
	}
	for _, item := range loadFile("data/items/equipments/crafted_equipment.json", false) {
		craftedArmor[item.ID] = item
	}
}

func loadFile(path string, weapon bool) []Item {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var f fileFormat
	if err := json.Unmarshal(b, &f); err != nil {
		return nil
	}
	if weapon {
		out := make([]Item, 0, len(f.Weapons))
		for _, w := range f.Weapons {
			out = append(out, Item{
				ID:            w.ID,
				Name:          w.Name,
				LevelMin:      w.LevelMin,
				LevelMax:      w.LevelMax,
				RequiredStats: members.Stats{Might: w.RequiredStats.Might, Mastery: w.RequiredStats.Mastery, Tactics: w.RequiredStats.Tactics, Survival: w.RequiredStats.Survival},
				Modifiers:     members.Stats{Might: w.Modifiers.Might, Mastery: w.Modifiers.Mastery, Tactics: w.Modifiers.Tactics, Survival: w.Modifiers.Survival},
			})
		}
		return out
	}
	out := make([]Item, 0, len(f.Armor))
	for _, a := range f.Armor {
		out = append(out, Item{
			ID:            a.ID,
			Name:          a.Name,
			LevelMin:      a.LevelMin,
			LevelMax:      a.LevelMax,
			RequiredStats: members.Stats{Might: a.RequiredStats.Might, Mastery: a.RequiredStats.Mastery, Tactics: a.RequiredStats.Tactics, Survival: a.RequiredStats.Survival},
			Modifiers:     members.Stats{Might: a.Modifiers.Might, Mastery: a.Modifiers.Mastery, Tactics: a.Modifiers.Tactics, Survival: a.Modifiers.Survival},
			Prot:          a.Prot,
		})
	}
	return out
}
