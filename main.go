package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	supportedSaveVersion = 1
	usageText            = "Usage: ./myapp <save_dir>\nExample: ./myapp ./saves/slot1"
)

type clanFile struct {
	Meta struct {
		SaveVersion int `json:"save_version"`
	} `json:"meta"`
	Clan struct {
		Name string `json:"name"`
		Day  int    `json:"day"`
		Gold int    `json:"gold"`
		Fame int    `json:"fame"`
	} `json:"clan"`
	Members   []json.RawMessage `json:"members"`
	Inventory struct {
		Weapons []json.RawMessage `json:"weapons"`
		Armor   []json.RawMessage `json:"armor"`
	} `json:"inventory"`
	InProgress struct {
		Crafting  []json.RawMessage `json:"crafting"`
		Upgrading []json.RawMessage `json:"upgrading"`
	} `json:"in_progress"`
}

type saveState struct {
	SaveDir             string
	SaveVersion         int
	ClanName            string
	CurrentDay          int
	Gold                int
	Fame                int
	MembersCount        int
	ActiveQuestsCount   int
	WeaponsCount        int
	ArmorCount          int
	InProgressCount     int
	ChronicleEntryCount int
	HasChronicle        bool
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, in io.Reader, out, errOut io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(errOut, "Missing required argument: <save_dir>")
		fmt.Fprintln(errOut, usageText)
		return 1
	}

	state, err := loadSave(args[0])
	if err != nil {
		fmt.Fprintf(errOut, "Failed to load save: %v\n", err)
		return 1
	}

	return repl(state, in, out, errOut)
}

func loadSave(saveDir string) (saveState, error) {
	info, err := os.Stat(saveDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return saveState{}, fmt.Errorf("invalid save directory path: directory not found: %s", saveDir)
		}
		return saveState{}, fmt.Errorf("invalid save directory path: %w", err)
	}
	if !info.IsDir() {
		return saveState{}, fmt.Errorf("invalid save directory path: not a directory: %s", saveDir)
	}

	clanPath := filepath.Join(saveDir, "clan.json")
	clanBytes, err := os.ReadFile(clanPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return saveState{}, errors.New("missing required file: clan.json not found")
		}
		return saveState{}, fmt.Errorf("cannot read clan.json: %w", err)
	}

	var clan clanFile
	if err := json.Unmarshal(clanBytes, &clan); err != nil {
		return saveState{}, fmt.Errorf("invalid JSON (clan.json): %w", err)
	}

	if clan.Meta.SaveVersion != supportedSaveVersion {
		return saveState{}, fmt.Errorf(
			"unsupported save_version: detected=%d supported=[%d]",
			clan.Meta.SaveVersion,
			supportedSaveVersion,
		)
	}

	questsCount, err := countActiveQuests(filepath.Join(saveDir, "quests.json"))
	if err != nil {
		return saveState{}, err
	}

	chronicleCount, hasChronicle, err := countChronicleEntries(filepath.Join(saveDir, "chronicle.jsonl"))
	if err != nil {
		return saveState{}, err
	}

	return saveState{
		SaveDir:             saveDir,
		SaveVersion:         clan.Meta.SaveVersion,
		ClanName:            clan.Clan.Name,
		CurrentDay:          clan.Clan.Day,
		Gold:                clan.Clan.Gold,
		Fame:                clan.Clan.Fame,
		MembersCount:        len(clan.Members),
		ActiveQuestsCount:   questsCount,
		WeaponsCount:        len(clan.Inventory.Weapons),
		ArmorCount:          len(clan.Inventory.Armor),
		InProgressCount:     len(clan.InProgress.Crafting) + len(clan.InProgress.Upgrading),
		ChronicleEntryCount: chronicleCount,
		HasChronicle:        hasChronicle,
	}, nil
}

func countActiveQuests(path string) (int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, fmt.Errorf("cannot read quests.json: %w", err)
	}

	var asMap map[string]json.RawMessage
	if err := json.Unmarshal(b, &asMap); err == nil {
		if raw, ok := asMap["active"]; ok {
			var active []json.RawMessage
			if err := json.Unmarshal(raw, &active); err != nil {
				return 0, fmt.Errorf("invalid JSON (quests.json active): %w", err)
			}
			return len(active), nil
		}
	}

	var arr []json.RawMessage
	if err := json.Unmarshal(b, &arr); err == nil {
		return len(arr), nil
	}

	return 0, fmt.Errorf("invalid JSON (quests.json): expected object with 'active' or array")
}

func countChronicleEntries(path string) (int, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("cannot read chronicle.jsonl: %w", err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	count := 0
	for s.Scan() {
		if strings.TrimSpace(s.Text()) != "" {
			count++
		}
	}
	if err := s.Err(); err != nil {
		return 0, false, fmt.Errorf("cannot scan chronicle.jsonl: %w", err)
	}
	return count, true, nil
}

type nodeType string

const (
	nodeDir  nodeType = "DIR"
	nodeView nodeType = "VIEW"
	nodeAct  nodeType = "ACT"
)

type vfsNode struct {
	name     string
	typeTag  nodeType
	children map[string]*vfsNode
	action   func(saveState, io.Writer)
}

func repl(state saveState, in io.Reader, out, errOut io.Writer) int {
	root := buildVFS()
	cwd := root
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprint(out, "> ")
		if !scanner.Scan() {
			return 0
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]

		switch cmd {
		case "ls":
			printLS(cwd, out)
		case "cd":
			if len(parts) != 2 {
				fmt.Fprintln(errOut, "cd requires exactly one path argument")
				continue
			}
			next, err := resolvePath(cwd, root, parts[1])
			if err != nil {
				fmt.Fprintf(errOut, "cd failed: %v\n", err)
				continue
			}
			if next.typeTag != nodeDir {
				fmt.Fprintln(errOut, "cd failed: target is not a directory")
				continue
			}
			cwd = next
		case "exit":
			return 0
		default:
			target, err := resolvePath(cwd, root, cmd)
			if err != nil {
				fmt.Fprintf(errOut, "unknown command or path: %s\n", cmd)
				continue
			}
			if target.typeTag == nodeDir {
				fmt.Fprintf(errOut, "cannot execute directory: %s\n", cmd)
				continue
			}
			if target.action != nil {
				target.action(state, out)
			}
			if target.name == "exit" {
				return 0
			}
		}
	}
}

func buildVFS() *vfsNode {
	root := &vfsNode{name: "/", typeTag: nodeDir, children: map[string]*vfsNode{}}
	clan := &vfsNode{name: "clan", typeTag: nodeDir, children: map[string]*vfsNode{}}
	status := &vfsNode{name: "status", typeTag: nodeView, action: printStatus}
	exit := &vfsNode{name: "exit", typeTag: nodeAct}

	root.children["clan"] = clan
	root.children["status"] = status
	root.children["exit"] = exit
	clan.children["status"] = status

	root.children["."] = root
	clan.children["."] = clan
	root.children[".."] = root
	clan.children[".."] = root

	return root
}

func resolvePath(cwd, root *vfsNode, path string) (*vfsNode, error) {
	if path == "/" {
		return root, nil
	}
	current := cwd
	if strings.HasPrefix(path, "/") {
		current = root
		path = strings.TrimPrefix(path, "/")
	}
	for _, part := range strings.Split(path, "/") {
		if part == "" || part == "." {
			continue
		}
		if part == ".." {
			next, ok := current.children[".."]
			if !ok {
				return nil, fmt.Errorf("cannot move above root")
			}
			current = next
			continue
		}
		next, ok := current.children[part]
		if !ok {
			return nil, fmt.Errorf("path not found: %s", path)
		}
		current = next
	}
	return current, nil
}

func printLS(cwd *vfsNode, out io.Writer) {
	names := make([]string, 0, len(cwd.children))
	for name := range cwd.children {
		if name == "." || name == ".." {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		node := cwd.children[name]
		fmt.Fprintf(out, "[%s] %s\n", node.typeTag, name)
	}
}

func printStatus(s saveState, out io.Writer) {
	fmt.Fprintln(out, "Status")
	fmt.Fprintf(out, "Save Directory: %s\n", s.SaveDir)
	fmt.Fprintf(out, "Save Version: %d\n", s.SaveVersion)
	fmt.Fprintf(out, "Clan Name: %s\n", s.ClanName)
	fmt.Fprintf(out, "Current Day: %d\n", s.CurrentDay)
	fmt.Fprintf(out, "Gold: %d\n", s.Gold)
	fmt.Fprintf(out, "Fame: %d\n", s.Fame)
	fmt.Fprintf(out, "Total Members: %d\n", s.MembersCount)
	fmt.Fprintf(out, "Active Quests: %d\n", s.ActiveQuestsCount)
	fmt.Fprintf(out, "Weapons: %d\n", s.WeaponsCount)
	fmt.Fprintf(out, "Armor: %d\n", s.ArmorCount)
	fmt.Fprintf(out, "In-Progress Craft/Upgrade: %d\n", s.InProgressCount)
	if s.HasChronicle {
		fmt.Fprintf(out, "Chronicle Entries: %d\n", s.ChronicleEntryCount)
	} else {
		fmt.Fprintln(out, "Chronicle Entries: -")
	}
}
