package repl

import (
	"fmt"
	"io"
	"strings"

	"chronicle-of-a-clan/internal/core/save"
	"chronicle-of-a-clan/internal/ui/vfs"
	prompt "github.com/c-bata/go-prompt"
)

type Session struct {
	state save.State
	root  *vfs.Node
	cwd   *vfs.Node
	out   io.Writer
	err   io.Writer
	done  bool
}

func NewSession(state save.State, out, err io.Writer) *Session {
	root := vfs.NewTree()
	return &Session{state: state, root: root, cwd: root, out: out, err: err}
}

func (s *Session) RunPrompt() int {
	executor := func(in string) {
		s.ExecuteLine(in)
	}
	completer := func(d prompt.Document) []prompt.Suggest {
		return s.complete(d)
	}
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("> "),
	)
	p.Run()
	return 0
}

func (s *Session) IsDone() bool { return s.done }

func (s *Session) ExecuteLine(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	parts := strings.Fields(line)
	cmd := parts[0]

	switch cmd {
	case "ls":
		target := s.cwd
		if len(parts) == 2 {
			next, err := vfs.Resolve(s.cwd, s.root, parts[1])
			if err != nil {
				fmt.Fprintf(s.err, "ls failed: %v\n", err)
				return
			}
			if next.Type() != vfs.NodeDir {
				fmt.Fprintln(s.err, "ls failed: target is not a directory")
				return
			}
			target = next
		} else if len(parts) > 2 {
			fmt.Fprintln(s.err, "ls accepts zero or one path argument")
			return
		}
		for _, row := range vfs.List(target) {
			fmt.Fprintln(s.out, row)
		}
	case "cd":
		if len(parts) != 2 {
			fmt.Fprintln(s.err, "cd requires exactly one path argument")
			return
		}
		next, err := vfs.Resolve(s.cwd, s.root, parts[1])
		if err != nil {
			fmt.Fprintf(s.err, "cd failed: %v\n", err)
			return
		}
		if next.Type() != vfs.NodeDir {
			fmt.Fprintln(s.err, "cd failed: target is not a directory")
			return
		}
		s.cwd = next
	case "exit":
		s.done = true
	default:
		target, err := vfs.Resolve(s.cwd, s.root, cmd)
		if err != nil {
			fmt.Fprintf(s.err, "unknown command or path: %s\n", cmd)
			return
		}
		if target.Type() == vfs.NodeDir {
			fmt.Fprintf(s.err, "cannot execute directory: %s\n", cmd)
			return
		}
		s.executeNode(target)
	}
}

func (s *Session) executeNode(target *vfs.Node) {
	switch target.Name() {
	case "status":
		fmt.Fprint(s.out, save.FormatStatus(s.state))
	case "exit":
		s.done = true
	}
}

func (s *Session) complete(d prompt.Document) []prompt.Suggest {
	line := d.CurrentLineBeforeCursor()
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return prompt.FilterHasPrefix([]prompt.Suggest{{Text: "ls"}, {Text: "cd"}, {Text: "status"}, {Text: "exit"}}, "", true)
	}

	commands := []prompt.Suggest{{Text: "ls"}, {Text: "cd"}, {Text: "status"}, {Text: "exit"}}
	if len(fields) == 1 && !strings.HasSuffix(line, " ") {
		first := fields[0]
		suggest := prompt.FilterHasPrefix(commands, first, true)
		for _, p := range vfs.CompletePathSuggestions(s.cwd, s.root, first) {
			suggest = append(suggest, prompt.Suggest{Text: p})
		}
		return suggest
	}

	if fields[0] == "cd" || fields[0] == "ls" {
		if strings.HasSuffix(line, " ") {
			fields = append(fields, "")
		}
		token := ""
		if len(fields) >= 2 {
			token = fields[1]
		}
		pathSuggestions := vfs.CompletePathSuggestions(s.cwd, s.root, token)
		items := make([]prompt.Suggest, 0, len(pathSuggestions))
		for _, p := range pathSuggestions {
			items = append(items, prompt.Suggest{Text: p})
		}
		return items
	}
	return nil
}

func (s *Session) CurrentPath() string {
	return vfs.DirPath(s.cwd)
}
