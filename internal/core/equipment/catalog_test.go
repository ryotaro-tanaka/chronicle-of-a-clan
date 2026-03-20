package equipment

import (
	"testing"

	"chronicle-of-a-clan/internal/core/members"
	"chronicle-of-a-clan/internal/core/save"
)

func TestCandidatesForMember(t *testing.T) {
	catalog, err := LoadCatalog()
	if err != nil {
		t.Skipf("LoadCatalog: %v", err)
	}

	stats, err := members.BaseStats("GROWTH_005", 1)
	if err != nil {
		t.Fatalf("BaseStats: %v", err)
	}

	inventory := save.Inventory{
		Weapons: []save.InventoryEntry{{ID: "CW_001", Count: 1}},
		Armor:   []save.InventoryEntry{{ID: "CA_001", Count: 1}},
	}

	weapons, armor := CandidatesForMember(catalog, inventory, 1, stats)
	if !containsOption(weapons, "R_W_001") {
		t.Fatalf("expected rental weapon candidate, got %+v", weapons)
	}
	if !containsOption(weapons, "CW_001") {
		t.Fatalf("expected crafted weapon candidate, got %+v", weapons)
	}
	if containsOption(weapons, "R_W_003") {
		t.Fatalf("unexpected high-level rental candidate, got %+v", weapons)
	}
	if !containsOption(armor, "R_A_001") || !containsOption(armor, "CA_001") {
		t.Fatalf("expected armor candidates, got %+v", armor)
	}
}

func containsOption(options []EquipmentOption, id string) bool {
	for _, option := range options {
		if option.ID == id {
			return true
		}
	}
	return false
}
