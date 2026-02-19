package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var _ = fmt.Print

func findInPath(cmd string) string {
	pathEnv := os.Getenv("PATH")
	dirs := strings.Split(pathEnv, string(os.PathListSeparator))
	for _, dir := range dirs {
		full := filepath.Join(dir, cmd)
		info, err := os.Stat(full)
		if err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
			return full
		}
	}
	return ""
}

func parseArgs(line string) []string {
	var args []string
	var current strings.Builder
	inSingle := false
	inDouble := false

	for i := 0; i < len(line); i++ {
		c := line[i]
		if inSingle {
			if c == '\'' {
				inSingle = false
			} else {
				current.WriteByte(c)
			}
		} else if inDouble {
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
		} else {
			switch c {
			case '\'':
				inSingle = true
			case '"':
				inDouble = true
			case ' ', '\t':
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			case '\\':
				if i+1 < len(line) {
					current.WriteByte(line[i+1])
					i++
				}
			default:
				current.WriteByte(c)
			}
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

func extractRedirect(parts []string) (args []string, outFile string, outAppend bool, errFile string, errAppend bool) {
	for i := 0; i < len(parts); i++ {
		if (parts[i] == ">>" || parts[i] == "1>>") && i+1 < len(parts) {
			outFile = parts[i+1]
			outAppend = true
			i++
		} else if (parts[i] == ">" || parts[i] == "1>") && i+1 < len(parts) {
			outFile = parts[i+1]
			outAppend = false
			i++
		} else if parts[i] == "2>>" && i+1 < len(parts) {
			errFile = parts[i+1]
			errAppend = true
			i++
		} else if parts[i] == "2>" && i+1 < len(parts) {
			errFile = parts[i+1]
			errAppend = false
			i++
		} else {
			args = append(args, parts[i])
		}
	}
	return
}

func openOutput(path string, append bool) (*os.File, error) {
	if append {
		return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}
	return os.Create(path)
}
func openFile(path string) (*os.File, error) {
	return os.Create(path)
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("$ ")

		line, err := reader.ReadString('\n')
		if err != nil {
			os.Exit(0)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := parseArgs(line)
		parts, outFile, outAppend, errFile, errAppend := extractRedirect(parts)

		var stdout io.Writer = os.Stdout
		var stderr io.Writer = os.Stderr

		if outFile != "" {
			f, err := openOutput(outFile, outAppend)
			if err != nil {
				fmt.Fprintf(os.Stderr, "cannot open %s: %v\n", outFile, err)
				continue
			}
			defer f.Close()
			stdout = f
		}

		if errFile != "" {
			f, err := openOutput(errFile, errAppend)
			if err != nil {
				fmt.Fprintf(os.Stderr, "cannot open %s: %v\n", errFile, err)
				continue
			}
			defer f.Close()
			stderr = f
		}

		switch parts[0] {
		case "exit":
			return

		case "echo":
			if len(parts) > 1 {
				fmt.Fprintln(stdout, strings.Join(parts[1:], " "))
			} else {
				fmt.Fprintln(stdout)
			}

		case "type":
			if len(parts) < 2 {
				continue
			}
			switch parts[1] {
			case "echo", "exit", "type", "pwd", "cd":
				fmt.Fprintf(stdout, "%s is a shell builtin\n", parts[1])
			default:
				if path := findInPath(parts[1]); path != "" {
					fmt.Fprintf(stdout, "%s is %s\n", parts[1], path)
				} else {
					fmt.Fprintf(stderr, "%s: not found\n", parts[1])
				}
			}

		case "pwd":
			if wd, err := os.Getwd(); err == nil {
				fmt.Fprintln(stdout, wd)
			}

		case "cd":
			dir := ""
			if len(parts) < 2 || parts[1] == "~" {
				dir = os.Getenv("HOME")
			} else {
				dir = parts[1]
			}
			if err := os.Chdir(dir); err != nil {
				fmt.Fprintf(stderr, "cd: %s: No such file or directory\n", dir)
			}

		default:
			if path := findInPath(parts[0]); path != "" {
				cmd := exec.Command(path, parts[1:]...)
				cmd.Args = parts
				cmd.Stdout = stdout
				cmd.Stderr = stderr
				cmd.Stdin = os.Stdin
				cmd.Run()
			} else {
				fmt.Fprintf(stderr, "%s: command not found\n", parts[0])
			}
		}
	}
}
