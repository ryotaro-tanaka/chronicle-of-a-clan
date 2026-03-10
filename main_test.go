package main

import (
	"bytes"
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

func TestRun_NoArgs(t *testing.T) {
	var out, errOut bytes.Buffer
	exit := run(nil, strings.NewReader(""), &out, &errOut)
	if exit == 0 {
		t.Fatalf("expected non-zero exit")
	}
	if !strings.Contains(errOut.String(), "Missing required argument") {
		t.Fatalf("unexpected stderr: %s", errOut.String())
	}
}

func TestRun_InvalidPath(t *testing.T) {
	var out, errOut bytes.Buffer
	exit := run([]string{"/does/not/exist"}, strings.NewReader(""), &out, &errOut)
	if exit == 0 {
		t.Fatalf("expected non-zero exit")
	}
	if !strings.Contains(errOut.String(), "invalid save directory path") {
		t.Fatalf("unexpected stderr: %s", errOut.String())
	}
}

func TestRun_MissingClanJSON(t *testing.T) {
	dir := t.TempDir()
	var out, errOut bytes.Buffer
	exit := run([]string{dir}, strings.NewReader(""), &out, &errOut)
	if exit == 0 {
		t.Fatalf("expected non-zero exit")
	}
	if !strings.Contains(errOut.String(), "clan.json not found") {
		t.Fatalf("unexpected stderr: %s", errOut.String())
	}
}

func TestRun_InvalidClanJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "clan.json", "{invalid")
	var out, errOut bytes.Buffer
	exit := run([]string{dir}, strings.NewReader(""), &out, &errOut)
	if exit == 0 {
		t.Fatalf("expected non-zero exit")
	}
	if !strings.Contains(errOut.String(), "invalid JSON (clan.json)") {
		t.Fatalf("unexpected stderr: %s", errOut.String())
	}
}

func TestRun_UnsupportedSaveVersion(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "clan.json", `{"meta":{"save_version":999},"clan":{"name":"A","day":1,"gold":1,"fame":1}}`)
	var out, errOut bytes.Buffer
	exit := run([]string{dir}, strings.NewReader(""), &out, &errOut)
	if exit == 0 {
		t.Fatalf("expected non-zero exit")
	}
	if !strings.Contains(errOut.String(), "unsupported save_version") {
		t.Fatalf("unexpected stderr: %s", errOut.String())
	}
}

func TestRun_SuccessAndStatusFields(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "clan.json", `{
"meta":{"save_version":1},
"clan":{"name":"Chronicle Clan","day":7,"gold":123,"fame":9},
"members":[{"id":"m1"}],
"inventory":{"weapons":[{"id":"w1"}],"armor":[]},
"in_progress":{"crafting":[{"id":"c1"}],"upgrading":[]}
}`)
	writeFile(t, dir, "quests.json", `{"active":[{"id":"q1"},{"id":"q2"}]}`)
	writeFile(t, dir, "chronicle.jsonl", "{\"day\":1}\n{\"day\":2}\n")

	var out, errOut bytes.Buffer
	exit := run([]string{dir}, strings.NewReader("status\n../exit\n"), &out, &errOut)
	if exit != 0 {
		t.Fatalf("expected success exit, got %d stderr=%s", exit, errOut.String())
	}

	output := out.String()
	required := []string{
		"Save Directory:",
		"Save Version:",
		"Clan Name:",
		"Current Day:",
		"Gold:",
		"Fame:",
		"Total Members:",
		"Active Quests:",
		"Weapons:",
		"Armor:",
		"In-Progress Craft/Upgrade:",
	}
	for _, field := range required {
		if !strings.Contains(output, field) {
			t.Fatalf("missing required status field %q in output: %s", field, output)
		}
	}
}
