package members

import "testing"

func TestLevelFromXP(t *testing.T) {
	if got := LevelFromXP(0); got != 1 {
		t.Fatalf("LevelFromXP(0): want 1, got %d", got)
	}
	if got := LevelFromXP(199); got != 2 {
		t.Fatalf("LevelFromXP(199): want 2, got %d", got)
	}
	if got := LevelFromXP(9150); got != 50 {
		t.Fatalf("LevelFromXP(9150): want 50, got %d", got)
	}
}
