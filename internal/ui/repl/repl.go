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
)

type Session struct {
	state        save.State
	root         *vfs.Node
	cwd          *vfs.Node
	bossProfiles *monsters.BossProfilesConfig
	out          io.Writer
	err          io.Writer
	done         bool
	onParty      func(questPath string)
	onClear      func(questPath string)
}

func NewSession(state save.State, root *vfs.Node, bossProfiles *monsters.BossProfilesConfig, out, err io.Writer) *Session {
	return &Session{state: state, root: root, cwd: root, bossProfiles: bossProfiles, out: out, err: err}
}

func (s *Session) SetActionHooks(onParty, onClear func(questPath string)) {
	s.onParty = onParty
	s.onClear = onClear
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
		s.handleLS(fields)
	case "cd":
		s.handleCD(fields)
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

func (s *Session) handleLS(fields []string) {
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
}

func (s *Session) handleCD(fields []string) {
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
}

func (s *Session) executeNode(target *vfs.Node, args []string) {
	switch target.Name() {
	case "status":
		fmt.Fprint(s.out, format.Status(status.FromState(s.state)))
	case "exit":
		s.done = true
	case "create_boss":
		s.handleCreateBoss(args)
	case "info":
		s.handleQuestInfo(target)
	case "party":
		if s.onParty != nil && target.Parent() != nil {
			s.onParty(vfs.DirPath(target.Parent()))
		}
	case "clear":
		if s.onClear != nil && target.Parent() != nil {
			s.onClear(vfs.DirPath(target.Parent()))
		}
	default:
		fmt.Fprintf(s.err, "unknown view or action: %s\n", target.Name())
	}
}

func (s *Session) handleQuestInfo(target *vfs.Node) {
	if target.ProfileID == "" || s.bossProfiles == nil {
		fmt.Fprintln(s.err, "info: no profile associated")
		return
	}
	_, profile, err := s.bossProfiles.ProfileByID(target.ProfileID)
	if err != nil {
		fmt.Fprintf(s.err, "info failed: %v\n", err)
		return
	}
	fmt.Fprint(s.out, format.QuestInfo(profile))
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

func parseInt64(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

func (s *Session) CurrentPath() string { return vfs.DirPath(s.cwd) }
