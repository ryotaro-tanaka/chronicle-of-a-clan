package repl

import (
	"bytes"
	"strings"
	"testing"

	"chronicle-of-a-clan/internal/core/monsters"
	"chronicle-of-a-clan/internal/core/quests"
	"chronicle-of-a-clan/internal/core/save"
	"chronicle-of-a-clan/internal/ui/vfs"
)

func newTestSession(state save.State, out, err *bytes.Buffer) *Session {
	return NewSession(state, vfs.NewTree(), nil, out, err)
}

func TestLSPathDoesNotChangeCWD(t *testing.T) {
	var out, err bytes.Buffer
	s := newTestSession(save.State{}, &out, &err)

	s.ExecuteLine("cd clan")
	if got := s.CurrentPath(); got != "/clan" {
		t.Fatalf("cwd before ls path = %s", got)
	}
	s.ExecuteLine("ls /")
	if got := s.CurrentPath(); got != "/clan" {
		t.Fatalf("ls <path> changed cwd: %s", got)
	}
}

func TestRelativePathsAndNormalization(t *testing.T) {
	var out, err bytes.Buffer
	s := newTestSession(save.State{}, &out, &err)

	s.ExecuteLine("cd clan/")
	if got := s.CurrentPath(); got != "/clan" {
		t.Fatalf("expected /clan, got %s", got)
	}
	s.ExecuteLine("cd ./..")
	if got := s.CurrentPath(); got != "/" {
		t.Fatalf("expected /, got %s", got)
	}
	s.ExecuteLine("cd ./clan")
	if got := s.CurrentPath(); got != "/clan" {
		t.Fatalf("expected /clan, got %s", got)
	}
}

func TestStatusFromClanPathInvocation(t *testing.T) {
	var out, err bytes.Buffer
	s := newTestSession(save.State{ClanName: "A", CurrentDay: 1, Gold: 2, Fame: 3, MembersCount: 4, ActiveQuestsCount: 5}, &out, &err)
	s.ExecuteLine("clan/status")
	if !strings.Contains(out.String(), "Clan: A   Day: 1") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestDevCreateBossInvalidProfileID(t *testing.T) {
	var out, err bytes.Buffer
	s := newTestSession(save.State{}, &out, &err)
	s.ExecuteLine("dev/create_boss nonexistent")
	if err.String() == "" && out.String() == "" {
		t.Fatalf("expected error or message for dev/create_boss with invalid profile_id")
	}
	if !strings.Contains(err.String(), "create_boss") && !strings.Contains(err.String(), "profile") {
		t.Fatalf("expected error to mention create_boss or profile, got err=%q", err.String())
	}
}

func TestDevCreateBossMissingProfileID(t *testing.T) {
	var out, err bytes.Buffer
	s := newTestSession(save.State{}, &out, &err)
	s.ExecuteLine("dev/create_boss")
	if !strings.Contains(err.String(), "profile_id") {
		t.Fatalf("expected error to mention profile_id when missing, got err=%q", err.String())
	}
}

func TestDevCreateBossReproducibleWithSeed(t *testing.T) {
	if _, err := monsters.LoadBossProfiles(); err != nil {
		t.Skipf("LoadBossProfiles: %v (run from repo root or data missing)", err)
	}
	var out1, err1 bytes.Buffer
	s1 := newTestSession(save.State{}, &out1, &err1)
	s1.ExecuteLine("dev/create_boss forest_003 99999")

	var out2, err2 bytes.Buffer
	s2 := newTestSession(save.State{}, &out2, &err2)
	s2.ExecuteLine("dev/create_boss forest_003 99999")

	if err1.String() != "" {
		t.Fatalf("first create_boss failed: %s", err1.String())
	}
	if err2.String() != "" {
		t.Fatalf("second create_boss failed: %s", err2.String())
	}
	if out1.String() != out2.String() {
		t.Fatalf("same profile_id + seed should produce identical output; got %q vs %q", out1.String(), out2.String())
	}
}

func TestExitSetsDone(t *testing.T) {
	var out, err bytes.Buffer
	s := newTestSession(save.State{}, &out, &err)
	if s.IsDone() {
		t.Fatal("session should not be done initially")
	}
	s.ExecuteLine("exit")
	if !s.IsDone() {
		t.Fatal("session should be done after exit")
	}
}

func TestPathBasedExitSetsDone(t *testing.T) {
	var out, err bytes.Buffer
	s := newTestSession(save.State{}, &out, &err)
	s.ExecuteLine("cd clan")
	if s.IsDone() {
		t.Fatal("session should not be done after cd")
	}
	s.ExecuteLine("../exit")
	if !s.IsDone() {
		t.Fatal("session should be done after ../exit from /clan")
	}
}

func TestQuestListAndInfo(t *testing.T) {
	keyQuests, err := quests.LoadKeyQuests()
	if err != nil {
		t.Skipf("LoadKeyQuests: %v (missing data/quests/key_quests.json?)", err)
	}
	bossProfiles, err := monsters.LoadBossProfiles()
	if err != nil {
		t.Skipf("LoadBossProfiles: %v (missing data/combat/boss_profiles.json?)", err)
	}
	state := save.State{KeyQuestCurrentOrder: 1}
	root := vfs.NewTree()
	vfs.AttachQuests(root, state, keyQuests, bossProfiles)
	var out, errBuf bytes.Buffer
	s := NewSession(state, root, bossProfiles, &out, &errBuf)

	s.ExecuteLine("ls")
	if !strings.Contains(out.String(), "[DIR] quests") {
		t.Errorf("ls at root should show [DIR] quests, got: %s", out.String())
	}
	out.Reset()
	errBuf.Reset()

	s.ExecuteLine("ls quests/")
	if !strings.Contains(out.String(), "[DIR] keys") {
		t.Errorf("ls quests/ should show [DIR] keys, got: %s", out.String())
	}
	if !strings.Contains(out.String(), "[DIR] forest") {
		t.Errorf("ls quests/ should show [DIR] forest, got: %s", out.String())
	}
	out.Reset()
	errBuf.Reset()

	s.ExecuteLine("ls quests/keys/")
	if !strings.Contains(out.String(), "hunt_ambushjaw_gator") {
		t.Errorf("ls quests/keys/ should show hunt_ambushjaw_gator (current_order=1), got: %s", out.String())
	}
	out.Reset()
	errBuf.Reset()

	s.ExecuteLine("ls quests/keys/hunt_ambushjaw_gator/")
	if !strings.Contains(out.String(), "[VIEW] info") {
		t.Errorf("ls quests/keys/hunt_ambushjaw_gator/ should show [VIEW] info, got: %s", out.String())
	}
	if !strings.Contains(out.String(), "[ACT] party") {
		t.Errorf("ls quests/keys/hunt_ambushjaw_gator/ should show [ACT] party, got: %s", out.String())
	}
	if !strings.Contains(out.String(), "[ACT] clear") {
		t.Errorf("ls quests/keys/hunt_ambushjaw_gator/ should show [ACT] clear, got: %s", out.String())
	}
	out.Reset()
	errBuf.Reset()

	s.ExecuteLine("quests/keys/hunt_ambushjaw_gator/info")
	if errBuf.String() != "" {
		t.Errorf("info should not error: %s", errBuf.String())
	}
	if !strings.Contains(out.String(), "Name: Ambushjaw Gator") {
		t.Errorf("info should show Name: Ambushjaw Gator, got: %s", out.String())
	}
	if !strings.Contains(out.String(), "Lv: 1-5") {
		t.Errorf("info should show Lv: 1-5, got: %s", out.String())
	}
	if !strings.Contains(out.String(), "Specialties:") {
		t.Errorf("info should show Specialties, got: %s", out.String())
	}
	if !strings.Contains(out.String(), "Reward:") {
		t.Errorf("info should show Reward line, got: %s", out.String())
	}
}

func TestPartyAndClearHooks(t *testing.T) {
	keyQuests, err := quests.LoadKeyQuests()
	if err != nil {
		t.Skipf("LoadKeyQuests: %v", err)
	}
	bossProfiles, err := monsters.LoadBossProfiles()
	if err != nil {
		t.Skipf("LoadBossProfiles: %v", err)
	}
	root := vfs.NewTree()
	vfs.AttachQuests(root, save.State{KeyQuestCurrentOrder: 1}, keyQuests, bossProfiles)
	var out, errBuf bytes.Buffer
	s := NewSession(save.State{}, root, bossProfiles, &out, &errBuf)

	var partyPath, clearPath string
	s.SetActionHooks(func(q string) { partyPath = q }, func(q string) { clearPath = q })
	s.ExecuteLine("quests/keys/hunt_ambushjaw_gator/party")
	if partyPath != "/quests/keys/hunt_ambushjaw_gator" {
		t.Fatalf("unexpected party path: %s", partyPath)
	}
	s.ExecuteLine("quests/keys/hunt_ambushjaw_gator/clear")
	if clearPath != "/quests/keys/hunt_ambushjaw_gator" {
		t.Fatalf("unexpected clear path: %s", clearPath)
	}
}

func TestInvalidLSAndCDArguments(t *testing.T) {
	var out, err bytes.Buffer
	s := newTestSession(save.State{}, &out, &err)

	s.ExecuteLine("ls a b")
	if !strings.Contains(err.String(), "ls accepts zero or one path argument") {
		t.Fatalf("unexpected ls arg validation error: %q", err.String())
	}
	err.Reset()

	s.ExecuteLine("cd")
	if !strings.Contains(err.String(), "cd requires exactly one path argument") {
		t.Fatalf("unexpected cd arg validation error: %q", err.String())
	}
}
