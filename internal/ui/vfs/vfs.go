package vfs

import (
	"fmt"
	"sort"
	"strings"

	"chronicle-of-a-clan/internal/core/monsters"
	"chronicle-of-a-clan/internal/core/quests"
	"chronicle-of-a-clan/internal/core/save"
)

type NodeType string

const (
	NodeDir  NodeType = "DIR"
	NodeView NodeType = "VIEW"
	NodeAct  NodeType = "ACT"
)

type Node struct {
	name      string
	typeTag   NodeType
	parent    *Node
	children  map[string]*Node
	ProfileID string // for quest info VIEW nodes: boss profile_id to display
}

func (n *Node) Name() string               { return n.name }
func (n *Node) Type() NodeType             { return n.typeTag }
func (n *Node) Parent() *Node              { return n.parent }
func (n *Node) Children() map[string]*Node { return n.children }

func NewTree() *Node {
	root := newDir("/", nil)
	clan := newDir("clan", root)
	newNode("status", NodeView, clan, "")
	dev := newDir("dev", root)
	newNode("create_boss", NodeAct, dev, "")
	newNode("exit", NodeAct, root, "")
	return root
}

// AttachQuests adds the quests subtree to root: quests/keys and quests/<region>.
// keys/ shows only the quest whose order == current_order (the next one to advance the story).
// quests/<region>/ shows all key quests with order <= current_order in that region.
func AttachQuests(root *Node, state save.State, keyQuestEntries []quests.Entry, bossProfiles *monsters.BossProfilesConfig) {
	keysCurrent := quests.CurrentOrder(keyQuestEntries, state.KeyQuestCurrentOrder)
	available := quests.Available(keyQuestEntries, state.KeyQuestCurrentOrder)

	questsDir := newDir("quests", root)
	keysDir := newDir("keys", questsDir)

	regionDirs := make(map[string]*Node)
	for regionID := range bossProfiles.Regions {
		regionDirs[regionID] = newDir(regionID, questsDir)
	}

	// keys/: only the quest(s) with order == current_order (one quest per order in practice)
	for _, e := range keysCurrent {
		_, profile, err := bossProfiles.ProfileByID(e.ProfileID)
		if err != nil {
			continue
		}
		slug := monsters.NameToSlug(profile.Name)
		huntName := "hunt_" + slug
		huntUnderKeys := newDir(huntName, keysDir)
		newViewWithProfile("info", e.ProfileID, huntUnderKeys)
		newNode("party", NodeAct, huntUnderKeys, "")
		newNode("clear", NodeAct, huntUnderKeys, "")
	}

	// quests/<region>/: all available (order <= current_order) in that region
	for _, e := range available {
		regionID, profile, err := bossProfiles.ProfileByID(e.ProfileID)
		if err != nil {
			continue
		}
		slug := monsters.NameToSlug(profile.Name)
		huntName := "hunt_" + slug
		if regionDir, ok := regionDirs[regionID]; ok {
			huntUnderRegion := newDir(huntName, regionDir)
			newViewWithProfile("info", e.ProfileID, huntUnderRegion)
			newNode("party", NodeAct, huntUnderRegion, "")
			newNode("clear", NodeAct, huntUnderRegion, "")
		}
	}
}

func newDir(name string, parent *Node) *Node {
	n := newNode(name, NodeDir, parent, "")
	return n
}

func newNode(name string, typ NodeType, parent *Node, profileID string) *Node {
	n := &Node{name: name, typeTag: typ, parent: parent, ProfileID: profileID}
	if typ == NodeDir {
		n.children = map[string]*Node{}
	}
	if parent != nil {
		parent.children[name] = n
	}
	return n
}

func newViewWithProfile(name, profileID string, parent *Node) *Node {
	return newNode(name, NodeView, parent, profileID)
}

func Resolve(cwd, root *Node, path string) (*Node, error) {
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
			if current.parent == nil {
				return nil, fmt.Errorf("cannot move above root")
			}
			current = current.parent
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

func List(dir *Node) []string {
	names := make([]string, 0, len(dir.children))
	for name := range dir.children {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]string, 0, len(names))
	for _, name := range names {
		node := dir.children[name]
		out = append(out, fmt.Sprintf("[%s] %s", node.typeTag, name))
	}
	return out
}

func DirPath(node *Node) string {
	if node == nil {
		return "/"
	}
	if node.parent == nil {
		return "/"
	}
	parts := []string{}
	cur := node
	for cur != nil && cur.parent != nil {
		parts = append([]string{cur.name}, parts...)
		cur = cur.parent
	}
	return "/" + strings.Join(parts, "/")
}

func CompletePathSuggestions(cwd, root *Node, token string) []string {
	base := token
	prefix := ""
	if idx := strings.LastIndex(token, "/"); idx >= 0 {
		base = token[idx+1:]
		prefix = token[:idx+1]
	}

	targetPath := prefix
	if targetPath == "" {
		targetPath = "."
	}
	dir, err := Resolve(cwd, root, targetPath)
	if err != nil || dir.typeTag != NodeDir {
		return nil
	}

	suggestions := make([]string, 0)
	for name, node := range dir.children {
		if !strings.HasPrefix(name, base) {
			continue
		}
		candidate := prefix + name
		if node.typeTag == NodeDir {
			candidate += "/"
		}
		suggestions = append(suggestions, candidate)
	}
	sort.Strings(suggestions)
	return suggestions
}
