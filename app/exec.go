package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
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

// runPipeline connects multiple commands with pipes:
// cmd1 | cmd2 | cmd3 ...
// stdout of each feeds into stdin of the next.
// Final command writes to the provided stdout/stderr.
func runPipeline(segments [][]string, stdout, stderr io.Writer) {
	if len(segments) == 1 {
		runExternal(segments[0], os.Stdin, stdout, stderr)
		return
	}

	cmds := make([]*exec.Cmd, len(segments))
	for i, seg := range segments {
		path := findInPath(seg[0])
		if path == "" {
			fmt.Fprintf(stderr, "%s: command not found\n", seg[0])
			return
		}
		cmds[i] = exec.Command(path, seg[1:]...)
		cmds[i].Args = seg
		cmds[i].Stderr = stderr
	}

	// wire up pipes between adjacent commands
	for i := 0; i < len(cmds)-1; i++ {
		r, w, err := os.Pipe()
		if err != nil {
			fmt.Fprintf(stderr, "pipe error: %v\n", err)
			return
		}
		cmds[i].Stdout = w
		cmds[i+1].Stdin = r
		// close write end after child starts so read end gets EOF
		defer w.Close()
		defer r.Close()
	}

	cmds[0].Stdin = os.Stdin
	cmds[len(cmds)-1].Stdout = stdout

	// start all commands
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(stderr, "start error: %v\n", err)
			return
		}
	}

	// wait for all to finish
	for _, cmd := range cmds {
		cmd.Wait()
	}
}
