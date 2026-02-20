package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

var commandHistory []string

var historyOffset int

func main() {
	if histFile := os.Getenv("HISTFILE"); histFile != "" {
		data, err := os.ReadFile(histFile)
		if err == nil {
			lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					commandHistory = append(commandHistory, line)
				}
			}
			historyOffset = len(commandHistory) // новые команды пишем после загруженных
		}
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "$ ",
		AutoComplete:    &shellCompleter{},
		InterruptPrompt: "^C",
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "readline init error:", err)
		os.Exit(1)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			saveHistory()
			os.Exit(0)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		commandHistory = append(commandHistory, line)

		parts := parseArgs(line)
		if len(parts) == 0 {
			continue
		}

		// при команде exit
		if parts[0] == "exit" {
			saveHistory()
			return
		}

		segments := splitPipeline(parts)

		if len(segments) > 1 {
			lastSeg, redir := extractRedirect(segments[len(segments)-1])
			segments[len(segments)-1] = lastSeg
			stdout, stderr := resolveStreams(redir)
			runPipeline(segments, stdout, stderr)
			continue
		}

		parts, redir := extractRedirect(segments[0])
		if len(parts) == 0 {
			continue
		}
		stdout, stderr := resolveStreams(redir)

		if parts[0] == "exit" {
			return
		}

		dispatchSingle(parts, stdout, stderr)
	}
}

func saveHistory() {
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
	for _, cmd := range commandHistory {
		fmt.Fprintln(f, cmd)
	}
}
