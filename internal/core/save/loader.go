package save

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var slotNamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

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
	Members []struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		GrowthTypeID string `json:"growth_type_id"`
		XP           int    `json:"xp"`
	} `json:"members"`
	Inventory struct {
		Weapons []struct {
			ID    string `json:"id"`
			Count int    `json:"count"`
		} `json:"weapons"`
		Armor []struct {
			ID    string `json:"id"`
			Count int    `json:"count"`
		} `json:"armor"`
	} `json:"inventory"`
	InProgress struct {
		Crafting  []json.RawMessage `json:"crafting"`
		Upgrading []json.RawMessage `json:"upgrading"`
	} `json:"in_progress"`
	KeyQuestProgress struct {
		CurrentOrder int `json:"current_order"`
	} `json:"key_quest_progress"`
}

func Load(slotName string) (State, error) {
	if err := validateSlotName(slotName); err != nil {
		return State{}, err
	}
	saveDir := slotPath(slotName)

	info, err := os.Stat(saveDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return State{}, fmt.Errorf("slot not found: %s", saveDir)
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

	members := make([]Member, 0, len(clan.Members))
	for _, m := range clan.Members {
		members = append(members, Member{
			ID:           m.ID,
			Name:         m.Name,
			GrowthTypeID: m.GrowthTypeID,
			XP:           m.XP,
		})
	}

	inv := Inventory{
		Weapons: make([]InventoryItem, 0, len(clan.Inventory.Weapons)),
		Armor:   make([]InventoryItem, 0, len(clan.Inventory.Armor)),
	}
	for _, w := range clan.Inventory.Weapons {
		inv.Weapons = append(inv.Weapons, InventoryItem{ID: w.ID, Count: w.Count})
	}
	for _, a := range clan.Inventory.Armor {
		inv.Armor = append(inv.Armor, InventoryItem{ID: a.ID, Count: a.Count})
	}

	keyOrder := clan.KeyQuestProgress.CurrentOrder
	if keyOrder < 1 {
		keyOrder = 1
	}

	return State{
		SaveDir:              saveDir,
		SaveVersion:          clan.Meta.SaveVersion,
		ClanName:             clan.Clan.Name,
		CurrentDay:           clan.Clan.Day,
		Gold:                 clan.Clan.Gold,
		Fame:                 clan.Clan.Fame,
		MembersCount:         len(clan.Members),
		ActiveQuestsCount:    questsCount,
		WeaponsCount:         len(clan.Inventory.Weapons),
		ArmorCount:           len(clan.Inventory.Armor),
		InProgressCount:      len(clan.InProgress.Crafting) + len(clan.InProgress.Upgrading),
		ChronicleEntryCount:  chronicleCount,
		HasChronicle:         hasChronicle,
		KeyQuestCurrentOrder: keyOrder,
		Members:              members,
		Inventory:            inv,
	}, nil
}

func Init(slotName string) error {
	if err := validateSlotName(slotName); err != nil {
		return err
	}

	root := savesRoot()
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("failed to create saves directory: %w", err)
	}

	target := slotPath(slotName)
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("slot already exists: %s", target)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to inspect slot: %w", err)
	}

	if err := os.MkdirAll(target, 0o755); err != nil {
		return fmt.Errorf("failed to create slot directory: %w", err)
	}

	template := filepath.Join("data", "save_init")
	if err := copyDir(template, target); err != nil {
		_ = os.RemoveAll(target)
		return fmt.Errorf("failed to copy init template: %w", err)
	}
	return nil
}

func validateSlotName(slotName string) error {
	if slotName == "" {
		return errors.New("invalid slot name: <save_dir> is required")
	}
	if strings.HasPrefix(slotName, "-") {
		return errors.New("invalid slot name: must not start with '-'; allowed characters are A-Za-z0-9._-")
	}
	if !slotNamePattern.MatchString(slotName) {
		return errors.New("invalid slot name: allowed characters are A-Za-z0-9._- and it must not contain path separators")
	}
	return nil
}

func savesRoot() string { return "saves" }

func slotPath(slotName string) string {
	return filepath.Join(savesRoot(), slotName)
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !d.Type().IsRegular() {
			return nil
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
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
