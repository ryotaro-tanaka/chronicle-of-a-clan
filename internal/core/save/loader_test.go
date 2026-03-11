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

func withCWD(t *testing.T, dir string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
}

func TestLoad_InvalidSlotName(t *testing.T) {
	_, err := Load("bad/name")
	if err == nil || !strings.Contains(err.Error(), "invalid slot name") {
		t.Fatalf("expected invalid slot name error, got %v", err)
	}
}

func TestLoad_MissingSlot(t *testing.T) {
	dir := t.TempDir()
	withCWD(t, dir)

	_, err := Load("missing-slot")
	if err == nil || !strings.Contains(err.Error(), "slot not found") {
		t.Fatalf("expected missing slot error, got %v", err)
	}
}

func TestLoad_MissingClanJSON(t *testing.T) {
	dir := t.TempDir()
	withCWD(t, dir)
	if err := os.MkdirAll(filepath.Join("saves", "slot1"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	_, err := Load("slot1")
	if err == nil || !strings.Contains(err.Error(), "clan.json not found") {
		t.Fatalf("expected missing clan.json error, got %v", err)
	}
}

func TestLoad_UnsupportedSaveVersion(t *testing.T) {
	dir := t.TempDir()
	withCWD(t, dir)
	slotDir := filepath.Join("saves", "slot1")
	if err := os.MkdirAll(slotDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, slotDir, "clan.json", `{"meta":{"save_version":999},"clan":{"name":"A","day":1,"gold":1,"fame":1}}`)
	_, err := Load("slot1")
	if err == nil || !strings.Contains(err.Error(), "unsupported save_version") {
		t.Fatalf("expected unsupported save_version error, got %v", err)
	}
}

func TestInit_CreatesSlotFromTemplate(t *testing.T) {
	dir := t.TempDir()
	withCWD(t, dir)
	if err := os.MkdirAll(filepath.Join("examples", "save_init_template"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, filepath.Join("examples", "save_init_template"), "clan.json", `{"meta":{"save_version":1},"clan":{"name":"A"}}`)

	if err := Init("slot1"); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join("saves", "slot1", "clan.json")); err != nil {
		t.Fatalf("expected copied clan.json: %v", err)
	}
}

func TestInit_FailsWhenSlotExists(t *testing.T) {
	dir := t.TempDir()
	withCWD(t, dir)
	if err := os.MkdirAll(filepath.Join("examples", "save_init_template"), 0o755); err != nil {
		t.Fatalf("mkdir template: %v", err)
	}
	writeFile(t, filepath.Join("examples", "save_init_template"), "clan.json", "{}")
	if err := os.MkdirAll(filepath.Join("saves", "slot1"), 0o755); err != nil {
		t.Fatalf("mkdir slot: %v", err)
	}

	err := Init("slot1")
	if err == nil || !strings.Contains(err.Error(), "slot already exists") {
		t.Fatalf("expected slot exists error, got %v", err)
	}
}

func TestInit_FailsForInvalidSlotName(t *testing.T) {
	err := Init("-bad")
	if err == nil || !strings.Contains(err.Error(), "invalid slot name") {
		t.Fatalf("expected invalid slot name error, got %v", err)
	}
}
