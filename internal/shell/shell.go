package shell

import (
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

// Shell is the main interactive shell instance.
type Shell struct {
	rl            *readline.Instance
	history       []string
	historyOffset int
}

// New creates and initializes a new Shell instance.
func New() (*Shell, error) {
	s := &Shell{}
	s.loadHistory()

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "$ ",
		AutoComplete:    newCompleter(),
		InterruptPrompt: "^C",
	})
	if err != nil {
		return nil, err
	}
	s.rl = rl

	return s, nil
}

// Run starts the REPL loop and returns the exit code.
func (s *Shell) Run() int {
	defer s.rl.Close()

	for {
		line, err := s.rl.Readline()
		if err != nil {
			s.saveHistory()
			return 0
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		s.history = append(s.history, line)

		parts := parseArgs(line)
		if len(parts) == 0 {
			continue
		}

		if parts[0] == "exit" {
			s.saveHistory()
			return 0
		}

		segments := splitPipeline(parts)

		if len(segments) > 1 {
			lastSeg, redir := extractRedirect(segments[len(segments)-1])
			segments[len(segments)-1] = lastSeg
			stdout, stderr := resolveStreams(redir)
			s.runPipeline(segments, stdout, stderr)
			continue
		}

		parts, redir := extractRedirect(segments[0])
		if len(parts) == 0 {
			continue
		}
		stdout, stderr := resolveStreams(redir)

		if parts[0] == "exit" {
			s.saveHistory()
			return 0
		}

		s.dispatch(parts, os.Stdin, stdout, stderr)
	}
}

func (s *Shell) loadHistory() {
	histFile := os.Getenv("HISTFILE")
	if histFile == "" {
		return
	}
	data, err := os.ReadFile(histFile)
	if err != nil {
		return
	}
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			s.history = append(s.history, line)
		}
	}
	s.historyOffset = len(s.history)
}

func (s *Shell) saveHistory() {
	histFile := os.Getenv("HISTFILE")
	if histFile == "" {
		return
	}
	f, err := os.Create(histFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "history: %s: %v\n", histFile, err)
		return
	}
	defer f.Close()
	for _, cmd := range s.history {
		fmt.Fprintln(f, cmd)
	}
}
