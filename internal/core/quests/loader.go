package quests

import (
	"encoding/json"
	"fmt"
	"os"
)

const keyQuestsPath = "data/key_quests.json"

// Entry represents a single key quest (order + profile_id).
type Entry struct {
	Order     int    `json:"order"`
	ProfileID string `json:"profile_id"`
}

type keyQuestsFile struct {
	KeyQuests []Entry `json:"key_quests"`
}

// LoadKeyQuests loads the key quest list from data/key_quests.json.
func LoadKeyQuests() ([]Entry, error) {
	f, err := os.Open(keyQuestsPath)
	if err != nil {
		return nil, fmt.Errorf("open key quests: %w", err)
	}
	defer f.Close()

	var file keyQuestsFile
	if err := json.NewDecoder(f).Decode(&file); err != nil {
		return nil, fmt.Errorf("decode key quests: %w", err)
	}
	return file.KeyQuests, nil
}

// Available returns entries with order <= currentOrder (for region listings).
func Available(entries []Entry, currentOrder int) []Entry {
	out := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if e.Order <= currentOrder {
			out = append(out, e)
		}
	}
	return out
}

// CurrentOrder returns entries with order == currentOrder (for keys/ listing: the next quest to advance the story).
func CurrentOrder(entries []Entry, currentOrder int) []Entry {
	out := make([]Entry, 0, 1)
	for _, e := range entries {
		if e.Order == currentOrder {
			out = append(out, e)
		}
	}
	return out
}
