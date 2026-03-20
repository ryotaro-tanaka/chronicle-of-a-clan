package members

import "testing"

func TestBaseStats(t *testing.T) {
	stats, err := BaseStats("GROWTH_005", 3)
	if err != nil {
		t.Fatalf("BaseStats: %v", err)
	}
	if stats.Might != 190 || stats.Mastery != 190 || stats.Tactics != 182 || stats.Survival != 182 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}
