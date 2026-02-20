package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadHistory(t *testing.T) {
	dir := t.TempDir()
	histFile := filepath.Join(dir, "history")
	os.WriteFile(histFile, []byte("cmd1\ncmd2\ncmd3\n"), 0644)

	t.Setenv("HISTFILE", histFile)

	s := &Shell{}
	s.loadHistory()

	if len(s.history) != 3 {
		t.Fatalf("len(history) = %d, want 3", len(s.history))
	}
	if s.history[0] != "cmd1" || s.history[2] != "cmd3" {
		t.Errorf("history = %v", s.history)
	}
	if s.historyOffset != 3 {
		t.Errorf("historyOffset = %d, want 3", s.historyOffset)
	}
}

func TestLoadHistoryNoFile(t *testing.T) {
	t.Setenv("HISTFILE", "/nonexistent/path/hist")
	s := &Shell{}
	s.loadHistory()
	if len(s.history) != 0 {
		t.Errorf("expected empty history, got %v", s.history)
	}
}

func TestLoadHistoryNoEnv(t *testing.T) {
	t.Setenv("HISTFILE", "")
	s := &Shell{}
	s.loadHistory()
	if len(s.history) != 0 {
		t.Errorf("expected empty history, got %v", s.history)
	}
}

func TestSaveHistory(t *testing.T) {
	dir := t.TempDir()
	histFile := filepath.Join(dir, "history")

	t.Setenv("HISTFILE", histFile)

	s := &Shell{history: []string{"alpha", "beta", "gamma"}}
	s.saveHistory()

	data, err := os.ReadFile(histFile)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 || lines[0] != "alpha" || lines[2] != "gamma" {
		t.Errorf("saved history = %v, want [alpha beta gamma]", lines)
	}
}

func TestSaveHistoryNoEnv(t *testing.T) {
	t.Setenv("HISTFILE", "")
	s := &Shell{history: []string{"test"}}
	s.saveHistory() // should not panic
}

func TestLoadHistorySkipsBlankLines(t *testing.T) {
	dir := t.TempDir()
	histFile := filepath.Join(dir, "history")
	os.WriteFile(histFile, []byte("cmd1\n\n  \ncmd2\n"), 0644)

	t.Setenv("HISTFILE", histFile)

	s := &Shell{}
	s.loadHistory()

	if len(s.history) != 2 {
		t.Fatalf("len(history) = %d, want 2", len(s.history))
	}
}
