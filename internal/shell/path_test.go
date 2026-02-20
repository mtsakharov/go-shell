package shell

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func createExec(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
}

func TestFindInPath(t *testing.T) {
	dir := t.TempDir()
	createExec(t, dir, "testcmd")

	orig := os.Getenv("PATH")
	t.Setenv("PATH", dir)
	defer os.Setenv("PATH", orig)

	t.Run("found", func(t *testing.T) {
		got := findInPath("testcmd")
		want := filepath.Join(dir, "testcmd")
		if got != want {
			t.Errorf("findInPath(\"testcmd\") = %q, want %q", got, want)
		}
	})

	t.Run("not found", func(t *testing.T) {
		got := findInPath("nonexistent_xyz")
		if got != "" {
			t.Errorf("findInPath(\"nonexistent_xyz\") = %q, want empty", got)
		}
	})

	t.Run("skips directories", func(t *testing.T) {
		subdir := filepath.Join(dir, "adir")
		os.Mkdir(subdir, 0755)
		got := findInPath("adir")
		if got != "" {
			t.Errorf("findInPath(\"adir\") = %q, want empty (should skip dirs)", got)
		}
	})

	t.Run("skips non-executable", func(t *testing.T) {
		path := filepath.Join(dir, "noexec")
		os.WriteFile(path, []byte("data"), 0644)
		got := findInPath("noexec")
		if got != "" {
			t.Errorf("findInPath(\"noexec\") = %q, want empty", got)
		}
	})
}

func TestExecutablesInPath(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	createExec(t, dir1, "foo-alpha")
	createExec(t, dir1, "foo-beta")
	createExec(t, dir2, "foo-gamma")
	createExec(t, dir2, "foo-alpha") // duplicate
	createExec(t, dir2, "bar-one")

	orig := os.Getenv("PATH")
	t.Setenv("PATH", dir1+string(os.PathListSeparator)+dir2)
	defer os.Setenv("PATH", orig)

	t.Run("prefix match", func(t *testing.T) {
		got := executablesInPath("foo-")
		sort.Strings(got)
		want := []string{"foo-alpha", "foo-beta", "foo-gamma"}
		if len(got) != len(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
			}
		}
	})

	t.Run("deduplicates", func(t *testing.T) {
		got := executablesInPath("foo-alpha")
		if len(got) != 1 {
			t.Errorf("got %v, want exactly one result", got)
		}
	})

	t.Run("no match", func(t *testing.T) {
		got := executablesInPath("zzz_no_match")
		if len(got) != 0 {
			t.Errorf("got %v, want empty", got)
		}
	})
}
