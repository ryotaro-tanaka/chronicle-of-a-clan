package repl

import (
	"fmt"
	"io"
	"strings"

	"chronicle-of-a-clan/internal/core/monsters"
	"chronicle-of-a-clan/internal/core/save"
	"chronicle-of-a-clan/internal/core/status"
	"chronicle-of-a-clan/internal/ui/format"
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
	exitChecker := func(in string, breakline bool) bool {
		if !breakline {
			return false
		}
		in = strings.TrimSpace(in)
		if in == "" {
			return false
		}
		parts := strings.Fields(in)
		cmd := parts[0]
		if cmd == "exit" {
			return true
		}
		target, err := vfs.Resolve(s.cwd, s.root, cmd)
		if err != nil {
			return false
		}
		return target.Type() == vfs.NodeAct && target.Name() == "exit"
	}
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("> "),
		prompt.OptionSetExitCheckerOnInput(exitChecker),
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
	fields := strings.Fields(line)
	cmd := fields[0]

	switch cmd {
	case "ls":
		target := s.cwd
		if len(fields) == 2 {
			next, err := vfs.Resolve(s.cwd, s.root, fields[1])
			if err != nil {
				fmt.Fprintf(s.err, "ls failed: %v\n", err)
				return
			}
			if next.Type() != vfs.NodeDir {
				fmt.Fprintln(s.err, "ls failed: target is not a directory")
				return
			}
			target = next
		} else if len(fields) > 2 {
			fmt.Fprintln(s.err, "ls accepts zero or one path argument")
			return
		}
		for _, row := range vfs.List(target) {
			fmt.Fprintln(s.out, row)
		}
	case "cd":
		if len(fields) != 2 {
			fmt.Fprintln(s.err, "cd requires exactly one path argument")
			return
		}
		next, err := vfs.Resolve(s.cwd, s.root, fields[1])
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
		args := []string{}
		if len(fields) > 1 {
			args = fields[1:]
		}
		s.executeNode(target, args)
	}
}

func (s *Session) executeNode(target *vfs.Node, args []string) {
	switch target.Name() {
	case "status":
		fmt.Fprint(s.out, format.Status(status.FromState(s.state)))
	case "exit":
		s.done = true
	case "create_boss":
		s.handleCreateBoss(args)
	}
}

func (s *Session) handleCreateBoss(args []string) {
	if len(args) < 1 || args[0] == "" {
		fmt.Fprintln(s.err, "create_boss requires profile_id")
		return
	}
	profileID := args[0]

	var seedOpt *int64
	if len(args) >= 2 {
		if v, err := parseInt64(args[1]); err == nil {
			seedOpt = &v
		} else {
			fmt.Fprintf(s.err, "invalid seed: %s\n", args[1])
			return
		}
	}

	boss, err := monsters.GenerateBoss(profileID, seedOpt)
	if err != nil {
		fmt.Fprintf(s.err, "create_boss failed: %v\n", err)
		return
	}
	fmt.Fprint(s.out, format.Boss(boss))
}

func parseInt(s string) (int, error) {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

func parseInt64(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

func (s *Session) complete(d prompt.Document) []prompt.Suggest {
	return s.completeLine(d.CurrentLineBeforeCursor())
}

func (s *Session) completeLine(line string) []prompt.Suggest {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil
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
