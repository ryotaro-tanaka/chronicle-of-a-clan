package save

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func Load(saveDir string) (State, error) {
	info, err := os.Stat(saveDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return State{}, fmt.Errorf("invalid save directory path: directory not found: %s", saveDir)
		}
		return State{}, fmt.Errorf("invalid save directory path: %w", err)
	}
	if !info.IsDir() {
		return State{}, fmt.Errorf("invalid save directory path: not a directory: %s", saveDir)
	}

	clanPath := filepath.Join(saveDir, "clan.json")
	clanBytes, err := os.ReadFile(clanPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return State{}, errors.New("missing required file: clan.json not found")
		}
		return State{}, fmt.Errorf("cannot read clan.json: %w", err)
	}

	var clan clanFile
	if err := json.Unmarshal(clanBytes, &clan); err != nil {
		return State{}, fmt.Errorf("invalid JSON (clan.json): %w", err)
	}
	if clan.Meta.SaveVersion != SupportedSaveVersion {
		return State{}, fmt.Errorf("unsupported save_version: detected=%d supported=[%d]", clan.Meta.SaveVersion, SupportedSaveVersion)
	}

	questsCount, err := countActiveQuests(filepath.Join(saveDir, "quests.json"))
	if err != nil {
		return State{}, err
	}
	chronicleCount, hasChronicle, err := countChronicleEntries(filepath.Join(saveDir, "chronicle.jsonl"))
	if err != nil {
		return State{}, err
	}

	return State{
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
	s.Buffer(make([]byte, 1024), 1024*1024)
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
