package tea

import (
	"bufio"
	"fmt"
	"os"
)

type Msg interface{}

type Cmd func() Msg

type Model interface {
	Init() Cmd
	Update(Msg) (Model, Cmd)
	View() string
}

type KeyMsg struct{ key string }

func (k KeyMsg) String() string { return k.key }

type quitMsg struct{}

func Quit() Msg { return quitMsg{} }

type Program struct{ model Model }

type ProgramOption interface{}

func NewProgram(m Model, _ ...ProgramOption) *Program { return &Program{model: m} }

func (p *Program) Run() (Model, error) {
	if cmd := p.model.Init(); cmd != nil {
		if msg := cmd(); msg != nil {
			if p.dispatch(msg) {
				return p.model, nil
			}
		}
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\r\x1b[2J\x1b[H")
		fmt.Print(p.model.View())
		r, _, err := reader.ReadRune()
		if err != nil {
			return p.model, err
		}
		k := string(r)
		switch r {
		case '\n', '\r':
			k = "enter"
		case 127, 8:
			k = "backspace"
		case 3:
			k = "ctrl+c"
		case 9:
			k = "tab"
		case 27:
			k = "esc"
		}
		if p.dispatch(KeyMsg{key: k}) {
			return p.model, nil
		}
	}
}

func (p *Program) dispatch(msg Msg) bool {
	next, cmd := p.model.Update(msg)
	p.model = next
	if _, ok := msg.(quitMsg); ok {
		return true
	}
	if cmd != nil {
		res := cmd()
		if _, ok := res.(quitMsg); ok {
			return true
		}
		if res != nil {
			return p.dispatch(res)
		}
	}
	return false
}
