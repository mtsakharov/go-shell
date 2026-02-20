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
	history := commandHistory

	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err == nil && n > 0 && n < len(history) {
			history = history[len(history)-n:]
		}
	}

	// offset — номер первой записи в срезе
	offset := len(commandHistory) - len(history) + 1
	for i, cmd := range history {
		fmt.Fprintf(stdout, "    %d  %s\n", offset+i, cmd)
	}
}
