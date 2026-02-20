package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var builtins = []string{"echo", "exit", "type", "pwd", "cd", "history"}

func runEcho(args []string, stdout io.Writer) {
	fmt.Fprintln(stdout, strings.Join(args, " "))
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

func runHistory(args []string, stdout io.Writer) {
	// history -r <file> — читаем из файла
	if len(args) >= 2 && args[0] == "-r" {
		data, err := os.ReadFile(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "history: %s: %v\n", args[1], err)
			return
		}
		lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				commandHistory = append(commandHistory, line)
			}
		}
		return
	}

	// history -w <file> — пишем в файл
	if len(args) >= 2 && args[0] == "-w" {
		var sb strings.Builder
		for _, cmd := range commandHistory {
			sb.WriteString(cmd)
			sb.WriteByte('\n')
		}
		err := os.WriteFile(args[1], []byte(sb.String()), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "history: %s: %v\n", args[1], err)
		}
		return
	}

	// history <n> — последние n записей
	history := commandHistory
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err == nil && n > 0 && n < len(history) {
			history = history[len(history)-n:]
		}
	}

	offset := len(commandHistory) - len(history) + 1
	for i, cmd := range history {
		fmt.Fprintf(stdout, "    %d  %s\n", offset+i, cmd)
	}
}
