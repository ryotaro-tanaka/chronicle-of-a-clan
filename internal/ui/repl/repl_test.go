package repl

import (
	"bytes"
	"strings"
	"testing"

	"chronicle-of-a-clan/internal/core/save"
)

func TestLSPathDoesNotChangeCWD(t *testing.T) {
	var out, err bytes.Buffer
	s := NewSession(save.State{}, &out, &err)

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
	s := NewSession(save.State{}, &out, &err)

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
	s := NewSession(save.State{ClanName: "A", CurrentDay: 1, Gold: 2, Fame: 3, MembersCount: 4, ActiveQuestsCount: 5}, &out, &err)
	s.ExecuteLine("clan/status")
	if !strings.Contains(out.String(), "Clan: A   Day: 1") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}
