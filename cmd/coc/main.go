package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"chronicle-of-a-clan/internal/core/save"
	"chronicle-of-a-clan/internal/ui/repl"
)

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

func run(argv []string, out, errOut io.Writer) int {
	if len(argv) < 2 {
		fmt.Fprintln(errOut, "Missing required argument: <save_dir>")
		fmt.Fprintln(errOut, usage(argv))
		return 1
	}

	state, err := save.Load(argv[1])
	if err != nil {
		fmt.Fprintf(errOut, "Failed to load save: %v\n", err)
		return 1
	}

	session := repl.NewSession(state, out, errOut)
	return session.RunPrompt()
}

func usage(argv []string) string {
	bin := "coc"
	if len(argv) > 0 && argv[0] != "" {
		bin = filepath.Base(argv[0])
	}
	return fmt.Sprintf("Usage: %s <save_dir>\nExample: %s ./saves/slot1", bin, bin)
}
