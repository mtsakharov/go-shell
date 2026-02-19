package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

var builtins = []string{"echo", "exit", "type", "pwd", "cd"}

// ── Path utilities ────────────────────────────────────────────────────────────

func findInPath(cmd string) string {
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		full := filepath.Join(dir, cmd)
		if info, err := os.Stat(full); err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
			return full
		}
	}
	return ""
}

func executablesInPath(prefix string) []string {
	seen := map[string]bool{}
	var results []string
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if seen[e.Name()] || !strings.HasPrefix(e.Name(), prefix) {
				continue
			}
			info, err := e.Info()
			if err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
				results = append(results, e.Name())
				seen[e.Name()] = true
			}
		}
	}
	return results
}

// ── Parsing ───────────────────────────────────────────────────────────────────

func parseArgs(line string) []string {
	var args []string
	var current strings.Builder
	inSingle, inDouble := false, false

	flush := func() {
		if current.Len() > 0 {
			args = append(args, current.String())
			current.Reset()
		}
	}

	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case inSingle:
			if c == '\'' {
				inSingle = false
			} else {
				current.WriteByte(c)
			}
		case inDouble:
			if c == '"' {
				inDouble = false
			} else if c == '\\' && i+1 < len(line) {
				next := line[i+1]
				if next == '"' || next == '\\' || next == '$' {
					current.WriteByte(next)
					i++
				} else {
					current.WriteByte(c)
				}
			} else {
				current.WriteByte(c)
			}
		case c == '\'':
			inSingle = true
		case c == '"':
			inDouble = true
		case c == ' ' || c == '\t':
			flush()
		case c == '\\' && i+1 < len(line):
			current.WriteByte(line[i+1])
			i++
		default:
			current.WriteByte(c)
		}
	}
	flush()
	return args
}

type redirect struct {
	outFile   string
	outAppend bool
	errFile   string
	errAppend bool
}

func extractRedirect(parts []string) (args []string, r redirect) {
	for i := 0; i < len(parts); i++ {
		tok := parts[i]
		hasNext := i+1 < len(parts)
		switch {
		case (tok == ">>" || tok == "1>>") && hasNext:
			r.outFile, r.outAppend = parts[i+1], true
			i++
		case (tok == ">" || tok == "1>") && hasNext:
			r.outFile, r.outAppend = parts[i+1], false
			i++
		case tok == "2>>" && hasNext:
			r.errFile, r.errAppend = parts[i+1], true
			i++
		case tok == "2>" && hasNext:
			r.errFile, r.errAppend = parts[i+1], false
			i++
		default:
			args = append(args, tok)
		}
	}
	return
}

func openOutput(path string, appendMode bool) (*os.File, error) {
	if appendMode {
		return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}
	return os.Create(path)
}

// ── Tab completion ────────────────────────────────────────────────────────────

type shellCompleter struct{}

func (sc *shellCompleter) Do(line []rune, pos int) ([][]rune, int) {
	input := string(line[:pos])

	// only complete the first word
	if strings.ContainsAny(input, " \t") {
		return nil, 0
	}

	if input == "" {
		return nil, 0
	}

	seen := map[string]bool{}
	var matches []string

	for _, b := range builtins {
		if strings.HasPrefix(b, input) && !seen[b] {
			matches = append(matches, b)
			seen[b] = true
		}
	}
	for _, name := range executablesInPath(input) {
		if !seen[name] {
			matches = append(matches, name)
			seen[name] = true
		}
	}

	if len(matches) == 0 {
		fmt.Fprint(os.Stderr, "\x07") // ring the bell
		return nil, 0
	}

	completions := make([][]rune, len(matches))
	for i, m := range matches {
		suffix := m[len(input):]
		if len(matches) == 1 {
			suffix += " "
		}
		completions[i] = []rune(suffix)
	}
	return completions, len(input)
}

// ── Builtins ──────────────────────────────────────────────────────────────────

func runEcho(args []string, stdout io.Writer) {
	if len(args) > 0 {
		fmt.Fprintln(stdout, strings.Join(args, " "))
	} else {
		fmt.Fprintln(stdout)
	}
}

func runType(args []string, stdout, stderr io.Writer) {
	if len(args) == 0 {
		return
	}
	cmd := args[0]
	for _, b := range builtins {
		if b == cmd {
			fmt.Fprintf(stdout, "%s is a shell builtin\n", cmd)
			return
		}
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

func runExternal(parts []string, stdout, stderr io.Writer) {
	path := findInPath(parts[0])
	if path == "" {
		fmt.Fprintf(stderr, "%s: command not found\n", parts[0])
		return
	}
	cmd := exec.Command(path, parts[1:]...)
	cmd.Args = parts
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}

// ── Main loop ─────────────────────────────────────────────────────────────────

func main() {
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
			os.Exit(0)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := parseArgs(line)
		parts, redir := extractRedirect(parts)
		if len(parts) == 0 {
			continue
		}

		stdout, stderr := resolveStreams(redir)

		switch parts[0] {
		case "exit":
			return
		case "echo":
			runEcho(parts[1:], stdout)
		case "type":
			runType(parts[1:], stdout, stderr)
		case "pwd":
			runPwd(stdout)
		case "cd":
			runCd(parts[1:], stderr)
		default:
			runExternal(parts, stdout, stderr)
		}
	}
}

func resolveStreams(r redirect) (stdout, stderr io.Writer) {
	stdout = os.Stdout
	stderr = os.Stderr

	if r.outFile != "" {
		f, err := openOutput(r.outFile, r.outAppend)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot open %s: %v\n", r.outFile, err)
		} else {
			stdout = f
		}
	}

	if r.errFile != "" {
		f, err := openOutput(r.errFile, r.errAppend)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot open %s: %v\n", r.errFile, err)
		} else {
			stderr = f
		}
	}

	return
}
