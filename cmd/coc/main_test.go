package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withCWD(t *testing.T, dir string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
}

func TestNoArgsUsageIncludesInit(t *testing.T) {
	var out, errOut bytes.Buffer
	exit := run([]string{"/tmp/mybin"}, &out, &errOut)
	if exit == 0 {
		t.Fatalf("expected non-zero exit")
	}
	got := errOut.String()
	if !strings.Contains(got, "Usage: mybin <save_dir>") || !strings.Contains(got, "Usage: mybin init <save_dir>") {
		t.Fatalf("unexpected usage output: %s", got)
	}
}

func TestInitMissingArgShowsUsage(t *testing.T) {
	var out, errOut bytes.Buffer
	exit := run([]string{"coc", "init"}, &out, &errOut)
	if exit == 0 {
		t.Fatalf("expected non-zero exit")
	}
	if !strings.Contains(errOut.String(), "Usage: coc init <save_dir>") {
		t.Fatalf("unexpected output: %s", errOut.String())
	}
}

func TestInitCreatesSaveSlot(t *testing.T) {
	dir := t.TempDir()
	withCWD(t, dir)

	templateDir := filepath.Join("data", "save_init")
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		t.Fatalf("mkdir template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "clan.json"), []byte(`{"meta":{"save_version":1},"clan":{"name":"A","day":1,"gold":1,"fame":1}}`), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	var out, errOut bytes.Buffer
	exit := run([]string{"coc", "init", "slot1"}, &out, &errOut)
	if exit != 0 {
		t.Fatalf("expected zero exit, got %d (%s)", exit, errOut.String())
	}
	if _, err := os.Stat(filepath.Join("saves", "slot1", "clan.json")); err != nil {
		t.Fatalf("expected slot to be created: %v", err)
	}
}
