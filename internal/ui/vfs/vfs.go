package vfs

import (
	"fmt"
	"sort"
	"strings"
)

type NodeType string

const (
	NodeDir  NodeType = "DIR"
	NodeView NodeType = "VIEW"
	NodeAct  NodeType = "ACT"
)

type Node struct {
	name     string
	typeTag  NodeType
	parent   *Node
	children map[string]*Node
}

func (n *Node) Name() string    { return n.name }
func (n *Node) Type() NodeType  { return n.typeTag }
func (n *Node) Parent() *Node   { return n.parent }
func (n *Node) Children() map[string]*Node { return n.children }

func NewTree() *Node {
	root := newDir("/", nil)
	clan := newDir("clan", root)
	newNode("status", NodeView, clan)
	dev := newDir("dev", root)
	newNode("create_boss", NodeAct, dev)
	newNode("exit", NodeAct, root)
	return root
}

func newDir(name string, parent *Node) *Node {
	n := newNode(name, NodeDir, parent)
	return n
}

func newNode(name string, typ NodeType, parent *Node) *Node {
	n := &Node{name: name, typeTag: typ, parent: parent}
	if typ == NodeDir {
		n.children = map[string]*Node{}
	}
	if parent != nil {
		parent.children[name] = n
	}
	return n
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
