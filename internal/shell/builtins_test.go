package shell

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)


func TestIsBuiltin(t *testing.T) {
	for _, name := range []string{"echo", "exit", "type", "pwd", "cd", "history"} {
		if !isBuiltin(name) {
			t.Errorf("isBuiltin(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"ls", "cat", "grep", "nonexistent", ""} {
		if isBuiltin(name) {
			t.Errorf("isBuiltin(%q) = true, want false", name)
		}
	}
}

func TestRunEcho(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"no args", nil, "\n"},
		{"single arg", []string{"hello"}, "hello\n"},
		{"multiple args", []string{"hello", "world"}, "hello world\n"},
		{"special chars", []string{"a>b", "c|d"}, "a>b c|d\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			runEcho(tt.args, &buf)
			if buf.String() != tt.want {
				t.Errorf("got %q, want %q", buf.String(), tt.want)
			}
		})
	}
}

func TestRunType(t *testing.T) {
	t.Run("builtin", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		runType([]string{"echo"}, &stdout, &stderr)
		if !strings.Contains(stdout.String(), "shell builtin") {
			t.Errorf("got %q, want to contain 'shell builtin'", stdout.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		runType([]string{"nonexistent_cmd_xyz"}, &stdout, &stderr)
		if !strings.Contains(stderr.String(), "not found") {
			t.Errorf("got stderr %q, want to contain 'not found'", stderr.String())
		}
	})

	t.Run("no args", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		runType(nil, &stdout, &stderr)
		if stdout.String() != "" || stderr.String() != "" {
			t.Errorf("expected no output for empty args")
		}
	})
}

func TestRunPwd(t *testing.T) {
	wd, _ := os.Getwd()
	var buf bytes.Buffer
	runPwd(&buf)
	got := strings.TrimSpace(buf.String())
	if got != wd {
		t.Errorf("got %q, want %q", got, wd)
	}
}

func TestRunCd(t *testing.T) {
	original, _ := os.Getwd()
	defer os.Chdir(original)

	t.Run("change to temp dir", func(t *testing.T) {
		dir := t.TempDir()
		// Resolve symlinks (macOS /var -> /private/var)
		dir, _ = filepath.EvalSymlinks(dir)
		var stderr bytes.Buffer
		runCd([]string{dir}, &stderr)
		if stderr.String() != "" {
			t.Errorf("unexpected stderr: %s", stderr.String())
		}
		wd, _ := os.Getwd()
		wd, _ = filepath.EvalSymlinks(wd)
		if wd != dir {
			t.Errorf("cwd = %q, want %q", wd, dir)
		}
	})

	t.Run("nonexistent dir", func(t *testing.T) {
		var stderr bytes.Buffer
		runCd([]string{"/nonexistent_dir_xyz"}, &stderr)
		if !strings.Contains(stderr.String(), "No such file or directory") {
			t.Errorf("got stderr %q, want to contain 'No such file or directory'", stderr.String())
		}
	})

	t.Run("tilde goes home", func(t *testing.T) {
		home := os.Getenv("HOME")
		if home == "" {
			t.Skip("HOME not set")
		}
		var stderr bytes.Buffer
		runCd([]string{"~"}, &stderr)
		wd, _ := os.Getwd()
		if wd != home {
			t.Errorf("cwd = %q, want %q", wd, home)
		}
	})

	t.Run("no args goes home", func(t *testing.T) {
		home := os.Getenv("HOME")
		if home == "" {
			t.Skip("HOME not set")
		}
		var stderr bytes.Buffer
		runCd(nil, &stderr)
		wd, _ := os.Getwd()
		if wd != home {
			t.Errorf("cwd = %q, want %q", wd, home)
		}
	})
}

func TestRunHistory(t *testing.T) {
	t.Run("display all", func(t *testing.T) {
		s := &Shell{history: []string{"echo a", "echo b", "echo c"}}
		var buf bytes.Buffer
		s.runHistory(nil, &buf)
		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Fatalf("got %d lines, want 3", len(lines))
		}
		if !strings.Contains(lines[0], "echo a") {
			t.Errorf("first line %q should contain 'echo a'", lines[0])
		}
	})

	t.Run("display last n", func(t *testing.T) {
		s := &Shell{history: []string{"echo a", "echo b", "echo c"}}
		var buf bytes.Buffer
		s.runHistory([]string{"2"}, &buf)
		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Fatalf("got %d lines, want 2", len(lines))
		}
		if !strings.Contains(lines[0], "echo b") {
			t.Errorf("first line %q should contain 'echo b'", lines[0])
		}
	})

	t.Run("write and read", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "hist")

		s := &Shell{history: []string{"cmd1", "cmd2"}}
		s.historyWrite(path)

		s2 := &Shell{}
		s2.historyRead(path)
		if len(s2.history) != 2 || s2.history[0] != "cmd1" || s2.history[1] != "cmd2" {
			t.Errorf("read history = %v, want [cmd1 cmd2]", s2.history)
		}
	})

	t.Run("append new entries", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "hist")

		s := &Shell{
			history:       []string{"old1", "old2", "new1"},
			historyOffset: 2,
		}
		s.historyAppend(path)

		data, _ := os.ReadFile(path)
		got := strings.TrimSpace(string(data))
		if got != "new1" {
			t.Errorf("appended = %q, want %q", got, "new1")
		}

		if s.historyOffset != 3 {
			t.Errorf("historyOffset = %d, want 3", s.historyOffset)
		}
	})
}
