package shell

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var builtinNames = []string{"echo", "exit", "type", "pwd", "cd", "history"}

func isBuiltin(name string) bool {
	for _, b := range builtinNames {
		if b == name {
			return true
		}
	}
	return false
}

func runEcho(args []string, stdout io.Writer) {
	fmt.Fprintln(stdout, strings.Join(args, " "))
}

func runType(args []string, stdout, stderr io.Writer) {
	if len(args) == 0 {
		return
	}
	cmd := args[0]
	if isBuiltin(cmd) {
		fmt.Fprintf(stdout, "%s is a shell builtin\n", cmd)
		return
	}
	if path := findInPath(cmd); path != "" {
		fmt.Fprintf(stdout, "%s is %s\n", cmd, path)
	} else {
		fmt.Fprintf(stderr, "%s: not found\n", cmd)
	}
}

func runPwd(stdout io.Writer) {
	if wd, err := os.Getwd(); err == nil {
		fmt.Fprintln(stdout, wd)
	}
}

func runCd(args []string, stderr io.Writer) {
	dir := os.Getenv("HOME")
	if len(args) > 0 && args[0] != "~" {
		dir = args[0]
	}
	if err := os.Chdir(dir); err != nil {
		fmt.Fprintf(stderr, "cd: %s: No such file or directory\n", dir)
	}
}

func (s *Shell) runHistory(args []string, stdout io.Writer) {
	if len(args) >= 2 {
		switch args[0] {
		case "-r":
			s.historyRead(args[1])
			return
		case "-w":
			s.historyWrite(args[1])
			return
		case "-a":
			s.historyAppend(args[1])
			return
		}
	}

	history := s.history
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err == nil && n > 0 && n < len(history) {
			history = history[len(history)-n:]
		}
	}

	offset := len(s.history) - len(history) + 1
	for i, cmd := range history {
		fmt.Fprintf(stdout, "    %d  %s\n", offset+i, cmd)
	}
}

func (s *Shell) historyRead(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "history: %s: %v\n", path, err)
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

func (s *Shell) historyWrite(path string) {
	var sb strings.Builder
	for _, cmd := range s.history {
		sb.WriteString(cmd)
		sb.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "history: %s: %v\n", path, err)
	}
}

func (s *Shell) historyAppend(path string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "history: %s: %v\n", path, err)
		return
	}
	defer f.Close()
	for _, cmd := range s.history[s.historyOffset:] {
		fmt.Fprintln(f, cmd)
	}
	s.historyOffset = len(s.history)
}
