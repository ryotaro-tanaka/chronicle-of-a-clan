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

func TestDevCreateBossUnknownArgsDoNotPanic(t *testing.T) {
	var out, err bytes.Buffer
	s := NewSession(save.State{}, &out, &err)
	// region does not exist; should report an error, not panic.
	s.ExecuteLine("dev/create_boss nowhere 1")
	if err.String() == "" && out.String() == "" {
		t.Fatalf("expected some output or error for dev/create_boss with invalid region")
	}
}

func TestExitSetsDone(t *testing.T) {
	var out, err bytes.Buffer
	s := NewSession(save.State{}, &out, &err)
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
	s := NewSession(save.State{}, &out, &err)
	s.ExecuteLine("cd clan")
	if s.IsDone() {
		t.Fatal("session should not be done after cd")
	}
	s.ExecuteLine("../exit")
	if !s.IsDone() {
		t.Fatal("session should be done after ../exit from /clan")
	}
}

func TestCompletionEmptyInputReturnsCommands(t *testing.T) {
	var out, err bytes.Buffer
	s := NewSession(save.State{}, &out, &err)
	got := s.completeLine("")
	texts := make([]string, len(got))
	for i, g := range got {
		texts[i] = g.Text
	}
	for _, want := range []string{"ls", "cd", "status", "exit"} {
		if !contains(texts, want) {
			t.Errorf("completion for empty input should include %q, got %v", want, texts)
		}
	}
}

func TestCompletionPrefixFiltersCommands(t *testing.T) {
	var out, err bytes.Buffer
	s := NewSession(save.State{}, &out, &err)
	got := s.completeLine("s")
	texts := make([]string, len(got))
	for i, g := range got {
		texts[i] = g.Text
	}
	if !contains(texts, "status") {
		t.Errorf("completion for \"s\" should include status, got %v", texts)
	}
}

func TestCompletionCdPathSuggestsClan(t *testing.T) {
	var out, err bytes.Buffer
	s := NewSession(save.State{}, &out, &err)
	got := s.completeLine("cd c")
	texts := make([]string, len(got))
	for i, g := range got {
		texts[i] = g.Text
	}
	if !contains(texts, "clan/") {
		t.Errorf("completion for \"cd c\" should include clan/, got %v", texts)
	}
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
