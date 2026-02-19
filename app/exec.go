package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func runExternal(parts []string, stdin io.Reader, stdout, stderr io.Writer) {
	path := findInPath(parts[0])
	if path == "" {
		fmt.Fprintf(stderr, "%s: command not found\n", parts[0])
		return
	}
	cmd := exec.Command(path, parts[1:]...)
	cmd.Args = parts
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Run()
}

// runSegment runs either a builtin or external command with given streams
func runSegment(parts []string, stdin io.Reader, stdout, stderr io.Writer) {
	if len(parts) == 0 {
		return
	}
	switch parts[0] {
	case "echo":
		runEcho(parts[1:], stdout)
	case "type":
		runType(parts[1:], stdout, stderr)
	case "pwd":
		runPwd(stdout)
	case "cd":
		runCd(parts[1:], stderr)
	default:
		runExternal(parts, stdin, stdout, stderr)
	}
}

func runPipeline(segments [][]string, stdout, stderr io.Writer) {
	if len(segments) == 1 {
		runSegment(segments[0], os.Stdin, stdout, stderr)
		return
	}

	n := len(segments)

	// build n-1 pipes
	readers := make([]*os.File, n-1)
	writers := make([]*os.File, n-1)
	for i := 0; i < n-1; i++ {
		r, w, err := os.Pipe()
		if err != nil {
			fmt.Fprintf(stderr, "pipe error: %v\n", err)
			return
		}
		readers[i] = r
		writers[i] = w
	}

	// resolve stdin/stdout for each segment
	stdinFor := func(i int) io.Reader {
		if i == 0 {
			return os.Stdin
		}
		return readers[i-1]
	}
	stdoutFor := func(i int) io.Writer {
		if i == n-1 {
			return stdout
		}
		return writers[i]
	}

	// check if segment is a builtin
	isBuiltin := func(parts []string) bool {
		if len(parts) == 0 {
			return false
		}
		switch parts[0] {
		case "echo", "type", "pwd", "cd":
			return true
		}
		return false
	}

	// start external commands
	var cmds []*exec.Cmd
	for i, seg := range segments {
		if isBuiltin(seg) {
			continue
		}
		path := findInPath(seg[0])
		if path == "" {
			fmt.Fprintf(stderr, "%s: command not found\n", seg[0])
			return
		}
		cmd := exec.Command(path, seg[1:]...)
		cmd.Args = seg
		cmd.Stdin = stdinFor(i)
		cmd.Stdout = stdoutFor(i)
		cmd.Stderr = stderr
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(stderr, "start error: %v\n", err)
			return
		}
		cmds = append(cmds, cmd)
	}

	// run builtins in goroutines (they need to write to pipe concurrently)
	done := make(chan struct{})
	builtinCount := 0
	for i, seg := range segments {
		if !isBuiltin(seg) {
			continue
		}
		builtinCount++
		i, seg := i, seg
		go func() {
			runSegment(seg, stdinFor(i), stdoutFor(i), stderr)
			// close write end so next command gets EOF
			if i < n-1 {
				writers[i].Close()
			}
			done <- struct{}{}
		}()
	}

	// close all pipe ends in parent after everyone has started
	for i := 0; i < n-1; i++ {
		// only close write end if not owned by a builtin goroutine
		if !isBuiltin(segments[i]) {
			writers[i].Close()
		}
		// read ends are always closed in parent
		if !isBuiltin(segments[i+1]) {
			readers[i].Close()
		}
	}

	// wait for builtins
	for i := 0; i < builtinCount; i++ {
		<-done
	}

	// wait for external commands
	for _, cmd := range cmds {
		cmd.Wait()
	}

	// close any remaining readers
	for _, r := range readers {
		r.Close()
	}
}

// isBuiltinName checks if a command name is a builtin (used in main.go)
func isBuiltinName(name string) bool {
	switch name {
	case "echo", "exit", "type", "pwd", "cd":
		return true
	}
	return false
}

// update main.go switch to use runSegment too
func dispatchSingle(parts []string, stdout, stderr io.Writer) {
	if len(parts) == 0 {
		return
	}
	switch parts[0] {
	case "exit":
		os.Exit(0)
	default:
		runSegment(parts, os.Stdin, stdout, stderr)
	}
}

// formatPipelineInput handles echo with escape sequences in pipelines
func expandEscapes(s string) string {
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	return s
}
