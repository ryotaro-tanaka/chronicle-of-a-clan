package equipment

import (
	"encoding/json"
	"os"
	"sync"

	"chronicle-of-a-clan/internal/core/datafiles"
	"chronicle-of-a-clan/internal/core/members"
	"chronicle-of-a-clan/internal/core/save"
)

type Kind string

const (
	KindWeapon Kind = "weapon"
	KindArmor  Kind = "armor"
)

type Source string

const (
	SourceRental  Source = "rental"
	SourceCrafted Source = "crafted"
)

type EquipmentOption struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Kind          Kind          `json:"-"`
	Source        Source        `json:"-"`
	LevelMin      int           `json:"level_min"`
	LevelMax      int           `json:"level_max"`
	RequiredStats members.Stats `json:"required_stats"`
	StatModifiers members.Stats `json:"stat_modifiers"`
	Prot          int           `json:"prot"`
}

type Catalog struct {
	Weapons []EquipmentOption
	Armor   []EquipmentOption
}

type rawCatalog struct {
	Weapons []EquipmentOption `json:"weapons"`
	Armor   []EquipmentOption `json:"armor"`
}

var (
	catalogOnce sync.Once
	cached      *Catalog
	cachedErr   error
)

func LoadCatalog() (*Catalog, error) {
	catalogOnce.Do(func() {
		rental, err := loadCatalogFile(datafiles.Path("data/items/equipments/rental_equipment.json"), SourceRental)
		if err != nil {
			cachedErr = err
			return
		}
		crafted, err := loadCatalogFile(datafiles.Path("data/items/equipments/crafted_equipment.json"), SourceCrafted)
		if err != nil {
			cachedErr = err
			return
		}

		cached = &Catalog{
			Weapons: append(rental.Weapons, crafted.Weapons...),
			Armor:   append(rental.Armor, crafted.Armor...),
		}
	})
	return cached, cachedErr
}

func CandidatesForMember(catalog *Catalog, inventory save.Inventory, level int, stats members.Stats) ([]EquipmentOption, []EquipmentOption) {
	ownedWeapons := ownedIDs(inventory.Weapons)
	ownedArmor := ownedIDs(inventory.Armor)

	weapons := filterOptions(catalog.Weapons, ownedWeapons, level, stats)
	armor := filterOptions(catalog.Armor, ownedArmor, level, stats)
	return weapons, armor
}

func filterOptions(options []EquipmentOption, owned map[string]bool, level int, stats members.Stats) []EquipmentOption {
	filtered := make([]EquipmentOption, 0, len(options))
	for _, option := range options {
		if option.Source == SourceCrafted && !owned[option.ID] {
			continue
		}
		if option.LevelMin != 0 && level < option.LevelMin {
			continue
		}
		if option.LevelMax != 0 && level > option.LevelMax {
			continue
		}
		if !stats.Meets(option.RequiredStats) {
			continue
		}
		filtered = append(filtered, option)
	}
	return filtered
}

func ownedIDs(entries []save.InventoryEntry) map[string]bool {
	owned := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if entry.Count > 0 {
			owned[entry.ID] = true
		}
	}
	return owned
}

func loadCatalogFile(path string, source Source) (*Catalog, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw rawCatalog
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}

	applyMetadata(raw.Weapons, KindWeapon, source)
	applyMetadata(raw.Armor, KindArmor, source)
	return &Catalog{
		Weapons: raw.Weapons,
		Armor:   raw.Armor,
	}, nil
}

func applyMetadata(options []EquipmentOption, kind Kind, source Source) {
	for i := range options {
		options[i].Kind = kind
		options[i].Source = source
	}
}
