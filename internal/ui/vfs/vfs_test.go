package vfs

import (
	"strings"
	"testing"

	"chronicle-of-a-clan/internal/core/monsters"
	"chronicle-of-a-clan/internal/core/quests"
	"chronicle-of-a-clan/internal/core/save"
)

func TestAttachQuestsAddsPartyAndClear(t *testing.T) {
	keyQuests, err := quests.LoadKeyQuests()
	if err != nil {
		t.Skipf("LoadKeyQuests: %v", err)
	}
	bossProfiles, err := monsters.LoadBossProfiles()
	if err != nil {
		t.Skipf("LoadBossProfiles: %v", err)
	}

	root := NewTree()
	AttachQuests(root, save.State{KeyQuestCurrentOrder: 1}, keyQuests, bossProfiles)

	questDir, err := Resolve(root, root, "/quests/keys/hunt_ambushjaw_gator")
	if err != nil {
		t.Fatalf("Resolve quest dir: %v", err)
	}
	rows := strings.Join(List(questDir), "\n")
	if !strings.Contains(rows, "[ACT] party") {
		t.Fatalf("expected party action, got %s", rows)
	}
	if !strings.Contains(rows, "[ACT] clear") {
		t.Fatalf("expected clear action, got %s", rows)
	}
}
