package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Suggest struct {
	Text string
}

type Document struct {
	text string
}

func (d Document) CurrentLineBeforeCursor() string {
	return d.text
}

type Executor func(string)
type Completer func(Document) []Suggest

type option struct {
	prefix string
}

type Option func(*option)

func OptionPrefix(prefix string) Option {
	return func(o *option) {
		o.prefix = prefix
	}
}

type Prompt struct {
	executor  Executor
	completer Completer
	opt       option
}

func New(executor Executor, completer Completer, opts ...Option) *Prompt {
	p := &Prompt{executor: executor, completer: completer, opt: option{prefix: "> "}}
	for _, opt := range opts {
		opt(&p.opt)
	}
	return p
}

func (p *Prompt) Run() {
	s := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprint(os.Stdout, p.opt.prefix)
		if !s.Scan() {
			return
		}
		line := strings.TrimRight(s.Text(), "\r\n")
		p.executor(line)
	}
}

func FilterHasPrefix(items []Suggest, prefix string, ignoreCase bool) []Suggest {
	out := make([]Suggest, 0, len(items))
	matchPrefix := prefix
	for _, item := range items {
		target := item.Text
		if ignoreCase {
			target = strings.ToLower(target)
			matchPrefix = strings.ToLower(prefix)
		}
		if strings.HasPrefix(target, matchPrefix) {
			out = append(out, item)
		}
	}
	return out
}
