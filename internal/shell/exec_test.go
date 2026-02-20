package shell

import (
	"bytes"
	"strings"
	"testing"
)

func TestDispatch(t *testing.T) {
	s := &Shell{}

	t.Run("echo", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		s.dispatch([]string{"echo", "hello", "world"}, nil, &stdout, &stderr)
		if stdout.String() != "hello world\n" {
			t.Errorf("got %q, want %q", stdout.String(), "hello world\n")
		}
	})

	t.Run("type builtin", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		s.dispatch([]string{"type", "pwd"}, nil, &stdout, &stderr)
		if !strings.Contains(stdout.String(), "shell builtin") {
			t.Errorf("got %q, want to contain 'shell builtin'", stdout.String())
		}
	})

	t.Run("command not found", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		s.dispatch([]string{"nonexistent_cmd_xyz"}, nil, &stdout, &stderr)
		if !strings.Contains(stderr.String(), "command not found") {
			t.Errorf("got stderr %q, want to contain 'command not found'", stderr.String())
		}
	})

	t.Run("empty parts", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		s.dispatch(nil, nil, &stdout, &stderr)
		if stdout.String() != "" || stderr.String() != "" {
			t.Error("expected no output for nil parts")
		}
	})
}

func TestRunExternal(t *testing.T) {
	t.Run("command not found", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		runExternal([]string{"nonexistent_cmd_xyz"}, nil, &stdout, &stderr)
		if !strings.Contains(stderr.String(), "command not found") {
			t.Errorf("got stderr %q, want to contain 'command not found'", stderr.String())
		}
	})

	t.Run("runs true", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		runExternal([]string{"true"}, nil, &stdout, &stderr)
		if stderr.String() != "" {
			t.Errorf("unexpected stderr: %s", stderr.String())
		}
	})

	t.Run("captures stdout", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		runExternal([]string{"echo", "test output"}, nil, &stdout, &stderr)
		if strings.TrimSpace(stdout.String()) != "test output" {
			t.Errorf("got %q, want %q", strings.TrimSpace(stdout.String()), "test output")
		}
	})
}

func TestRunPipelineSingleSegment(t *testing.T) {
	s := &Shell{}
	var stdout, stderr bytes.Buffer
	s.runPipeline([][]string{{"echo", "piped"}}, &stdout, &stderr)
	if strings.TrimSpace(stdout.String()) != "piped" {
		t.Errorf("got %q, want %q", strings.TrimSpace(stdout.String()), "piped")
	}
}
