package save

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", name, err)
	}
}

func TestLoad_InvalidPath(t *testing.T) {
	_, err := Load("/does/not/exist")
	if err == nil || !strings.Contains(err.Error(), "invalid save directory path") {
		t.Fatalf("expected invalid save path error, got %v", err)
	}
}

func TestLoad_MissingClanJSON(t *testing.T) {
	dir := t.TempDir()
	_, err := Load(dir)
	if err == nil || !strings.Contains(err.Error(), "clan.json not found") {
		t.Fatalf("expected missing clan.json error, got %v", err)
	}
}

func TestLoad_UnsupportedSaveVersion(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "clan.json", `{"meta":{"save_version":999},"clan":{"name":"A","day":1,"gold":1,"fame":1}}`)
	_, err := Load(dir)
	if err == nil || !strings.Contains(err.Error(), "unsupported save_version") {
		t.Fatalf("expected unsupported save_version error, got %v", err)
	}
}

func TestFormatStatus(t *testing.T) {
	got := FormatStatus(State{ClanName: "Chronicle Clan", CurrentDay: 7, Gold: 123, Fame: 9, MembersCount: 2, ActiveQuestsCount: 3})
	for _, line := range []string{
		"Clan: Chronicle Clan   Day: 7",
		"Gold: 123   Fame: 9",
		"Members: 2   ActiveQuests: 3",
	} {
		if !strings.Contains(got, line) {
			t.Fatalf("status line missing: %q in %q", line, got)
		}
	}
}
