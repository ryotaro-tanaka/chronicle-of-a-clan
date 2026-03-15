package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"chronicle-of-a-clan/internal/core/monsters"
	"chronicle-of-a-clan/internal/core/quests"
	"chronicle-of-a-clan/internal/core/save"
	"chronicle-of-a-clan/internal/ui/repl"
	"chronicle-of-a-clan/internal/ui/vfs"
)

func main() {
	code := run(os.Args, os.Stdout, os.Stderr)
	// Must run before os.Exit; defer would not run.
	restoreTerminal()
	os.Exit(code)
}

// restoreTerminal restores the terminal to cooked mode after go-prompt exits.
// go-prompt v0.2.6 leaves the terminal in raw mode (no echo); stty sane fixes it.
// Linux/WSL only; no-op on Windows where stty is not available.
func restoreTerminal() {
    c := exec.Command("stty", "sane")  // または "/bin/stty", "-raw", "echo"
    c.Stdin = os.Stdin
    _ = c.Run()
}

func run(argv []string, out, errOut io.Writer) int {
	if len(argv) < 2 {
		printUsage(errOut, argv)
		return 1
	}

	if argv[1] == "init" {
		if len(argv) != 3 {
			fmt.Fprintln(errOut, "Missing required argument: <save_dir>")
			fmt.Fprintln(errOut, initUsage(argv))
			return 1
		}
		if err := save.Init(argv[2]); err != nil {
			fmt.Fprintf(errOut, "Failed to initialize save: %v\n", err)
			return 1
		}
		fmt.Fprintf(out, "Initialized save slot: %s\n", argv[2])
		return 0
	}

	if len(argv) != 2 {
		printUsage(errOut, argv)
		return 1
	}

	state, err := save.Load(argv[1])
	if err != nil {
		fmt.Fprintf(errOut, "Failed to load save: %v\n", err)
		return 1
	}

	keyQuestEntries, err := quests.LoadKeyQuests()
	if err != nil {
		fmt.Fprintf(errOut, "Failed to load key quests: %v\n", err)
		return 1
	}

	bossProfiles, err := monsters.LoadBossProfiles()
	if err != nil {
		fmt.Fprintf(errOut, "Failed to load boss profiles: %v\n", err)
		return 1
	}

	root := vfs.NewTree()
	vfs.AttachQuests(root, state, keyQuestEntries, bossProfiles)

	session := repl.NewSession(state, root, bossProfiles, out, errOut)
	return session.RunPrompt()
}

func usage(argv []string) string {
	bin := binaryName(argv)
	return fmt.Sprintf("Usage: %s <save_dir>", bin)
}

func initUsage(argv []string) string {
	bin := binaryName(argv)
	return fmt.Sprintf("Usage: %s init <save_dir>", bin)
}

func printUsage(w io.Writer, argv []string) {
	fmt.Fprintln(w, usage(argv))
	fmt.Fprintln(w, initUsage(argv))
}

func binaryName(argv []string) string {
	bin := "coc"
	if len(argv) > 0 && argv[0] != "" {
		bin = filepath.Base(argv[0])
	}
	return bin
}
