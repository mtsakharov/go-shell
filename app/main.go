package main

import (
	"bufio"
	"fmt"
	_ "log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
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

	for i := 0; i < len(line); i++ {
		c := line[i]
		if inSingle {
			if c == '\'' {
				inSingle = false
			} else {
				current.WriteByte(c)
			}
		} else {
			switch c {
			case '\'':
				inSingle = true
			case ' ', '\t':
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
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

		switch parts[0] {
		case "exit":
			return

		case "echo":
			if len(parts) > 1 {
				fmt.Println(strings.Join(parts[1:], " "))
			} else {
				fmt.Println()
			}

		case "type":
			if len(parts) < 2 {
				continue
			}

			switch parts[1] {
			case "echo", "exit", "type", "pwd":
				fmt.Printf("%s is a shell builtin\n", parts[1])
			default:
				if path := findInPath(parts[1]); path != "" {
					fmt.Printf("%s is %s\n", parts[1], path)
				} else {
					fmt.Printf("%s: not found\n", parts[1])
				}
			}

		case "pwd":
			if wd, err := os.Getwd(); err == nil {
				fmt.Println(wd)
			}

		case "cd":
			dir := ""
			if len(parts) < 2 || parts[1] == "~" {
				dir = os.Getenv("HOME")
			} else {
				dir = parts[1]
			}
			if err := os.Chdir(dir); err != nil {
				fmt.Printf("cd: %s: No such file or directory\n", parts[1])
			}

		default:
			if path := findInPath(parts[0]); path != "" {
				cmd := exec.Command(path, parts[1:]...)
				cmd.Args = parts
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Stdin = os.Stdin
				err := cmd.Run()
				if err != nil {
					fmt.Printf("error")
				}
			} else {
				fmt.Println(parts[0] + ": command not found")
			}
		}
	}
}
