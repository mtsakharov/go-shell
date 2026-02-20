package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// dispatch routes a single command to the appropriate handler.
func (s *Shell) dispatch(parts []string, stdin io.Reader, stdout, stderr io.Writer) {
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
	case "history":
		s.runHistory(parts[1:], stdout)
	case "exit":
		os.Exit(0)
	default:
		runExternal(parts, stdin, stdout, stderr)
	}
}

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
	_ = cmd.Run()
}

// runPipeline executes a sequence of piped commands.
func (s *Shell) runPipeline(segments [][]string, stdout, stderr io.Writer) {
	if len(segments) == 1 {
		s.dispatch(segments[0], os.Stdin, stdout, stderr)
		return
	}

	n := len(segments)

	readers := make([]*os.File, n-1)
	writers := make([]*os.File, n-1)
	for i := range n - 1 {
		r, w, err := os.Pipe()
		if err != nil {
			fmt.Fprintf(stderr, "pipe error: %v\n", err)
			return
		}
		readers[i] = r
		writers[i] = w
	}

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

	segIsBuiltin := func(parts []string) bool {
		return len(parts) > 0 && isBuiltin(parts[0]) && parts[0] != "exit"
	}

	// start external commands
	var cmds []*exec.Cmd
	for i, seg := range segments {
		if segIsBuiltin(seg) {
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

	// run builtins in goroutines so they can write to pipes concurrently
	done := make(chan struct{})
	builtinCount := 0
	for i, seg := range segments {
		if !segIsBuiltin(seg) {
			continue
		}
		builtinCount++
		go func(idx int, parts []string) {
			s.dispatch(parts, stdinFor(idx), stdoutFor(idx), stderr)
			if idx < n-1 {
				writers[idx].Close()
			}
			done <- struct{}{}
		}(i, seg)
	}

	// close pipe ends not owned by builtin goroutines
	for i := range n - 1 {
		if !segIsBuiltin(segments[i]) {
			writers[i].Close()
		}
		if !segIsBuiltin(segments[i+1]) {
			readers[i].Close()
		}
	}

	for range builtinCount {
		<-done
	}
	for _, cmd := range cmds {
		_ = cmd.Wait()
	}
	for _, r := range readers {
		r.Close()
	}
}
