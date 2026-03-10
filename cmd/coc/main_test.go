package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestUsageUsesBinaryName(t *testing.T) {
	var out, errOut bytes.Buffer
	exit := run([]string{"/tmp/mybin"}, &out, &errOut)
	if exit == 0 {
		t.Fatalf("expected non-zero exit")
	}
	if !strings.Contains(errOut.String(), "Usage: mybin <save_dir>") {
		t.Fatalf("unexpected usage output: %s", errOut.String())
	}
}
