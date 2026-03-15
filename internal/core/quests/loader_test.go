package quests

import (
	"testing"
)

func TestLoadKeyQuests(t *testing.T) {
	entries, err := LoadKeyQuests()
	if err != nil {
		t.Fatalf("LoadKeyQuests: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one key quest")
	}
	first := entries[0]
	if first.Order != 1 {
		t.Errorf("first entry order: want 1, got %d", first.Order)
	}
	if first.ProfileID != "forest_003" {
		t.Errorf("first entry profile_id: want forest_003, got %s", first.ProfileID)
	}
	last := entries[len(entries)-1]
	if last.Order != 25 {
		t.Errorf("last entry order: want 25, got %d", last.Order)
	}
	if last.ProfileID != "forest_001" {
		t.Errorf("last entry profile_id: want forest_001, got %s", last.ProfileID)
	}
}

func TestAvailable(t *testing.T) {
	entries := []Entry{
		{Order: 1, ProfileID: "a"},
		{Order: 2, ProfileID: "b"},
		{Order: 3, ProfileID: "c"},
	}
	got := Available(entries, 2)
	if len(got) != 2 {
		t.Fatalf("Available(_, 2): want 2 entries, got %d", len(got))
	}
	if got[0].Order != 1 || got[1].Order != 2 {
		t.Errorf("Available(_, 2): want order 1,2 got %d,%d", got[0].Order, got[1].Order)
	}
	if len(Available(entries, 0)) != 0 {
		t.Errorf("Available(_, 0): want 0 entries")
	}
}

func TestCurrentOrder(t *testing.T) {
	entries := []Entry{
		{Order: 1, ProfileID: "a"},
		{Order: 2, ProfileID: "b"},
		{Order: 3, ProfileID: "c"},
	}
	got := CurrentOrder(entries, 2)
	if len(got) != 1 {
		t.Fatalf("CurrentOrder(_, 2): want 1 entry, got %d", len(got))
	}
	if got[0].Order != 2 || got[0].ProfileID != "b" {
		t.Errorf("CurrentOrder(_, 2): want order=2 profile_id=b, got %d %s", got[0].Order, got[0].ProfileID)
	}
	if len(CurrentOrder(entries, 0)) != 0 {
		t.Errorf("CurrentOrder(_, 0): want 0 entries")
	}
}
