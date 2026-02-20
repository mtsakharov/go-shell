package shell

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExtractRedirect(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		wantArgs []string
		wantR    redirect
	}{
		{
			name:     "no redirect",
			parts:    []string{"echo", "hello"},
			wantArgs: []string{"echo", "hello"},
			wantR:    redirect{},
		},
		{
			name:     "stdout truncate",
			parts:    []string{"echo", "hello", ">", "out.txt"},
			wantArgs: []string{"echo", "hello"},
			wantR:    redirect{outFile: "out.txt", outAppend: false},
		},
		{
			name:     "stdout append",
			parts:    []string{"echo", "hello", ">>", "out.txt"},
			wantArgs: []string{"echo", "hello"},
			wantR:    redirect{outFile: "out.txt", outAppend: true},
		},
		{
			name:     "stdout with 1>",
			parts:    []string{"echo", "hello", "1>", "out.txt"},
			wantArgs: []string{"echo", "hello"},
			wantR:    redirect{outFile: "out.txt", outAppend: false},
		},
		{
			name:     "stdout with 1>>",
			parts:    []string{"echo", "hello", "1>>", "out.txt"},
			wantArgs: []string{"echo", "hello"},
			wantR:    redirect{outFile: "out.txt", outAppend: true},
		},
		{
			name:     "stderr truncate",
			parts:    []string{"cmd", "2>", "err.log"},
			wantArgs: []string{"cmd"},
			wantR:    redirect{errFile: "err.log", errAppend: false},
		},
		{
			name:     "stderr append",
			parts:    []string{"cmd", "2>>", "err.log"},
			wantArgs: []string{"cmd"},
			wantR:    redirect{errFile: "err.log", errAppend: true},
		},
		{
			name:     "both stdout and stderr",
			parts:    []string{"cmd", ">", "out.txt", "2>", "err.log"},
			wantArgs: []string{"cmd"},
			wantR:    redirect{outFile: "out.txt", errFile: "err.log"},
		},
		{
			name:     "redirect without target is kept as arg",
			parts:    []string{"echo", ">"},
			wantArgs: []string{"echo", ">"},
			wantR:    redirect{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, r := extractRedirect(tt.parts)
			if !reflect.DeepEqual(args, tt.wantArgs) {
				t.Errorf("args = %v, want %v", args, tt.wantArgs)
			}
			if r != tt.wantR {
				t.Errorf("redirect = %+v, want %+v", r, tt.wantR)
			}
		})
	}
}

func TestOpenOutput(t *testing.T) {
	dir := t.TempDir()

	t.Run("create truncate", func(t *testing.T) {
		path := filepath.Join(dir, "truncate.txt")
		f, err := openOutput(path, false)
		if err != nil {
			t.Fatal(err)
		}
		f.WriteString("hello\n")
		f.Close()

		f, err = openOutput(path, false)
		if err != nil {
			t.Fatal(err)
		}
		f.WriteString("world\n")
		f.Close()

		data, _ := os.ReadFile(path)
		if string(data) != "world\n" {
			t.Errorf("got %q, want %q", string(data), "world\n")
		}
	})

	t.Run("create append", func(t *testing.T) {
		path := filepath.Join(dir, "append.txt")
		f, err := openOutput(path, true)
		if err != nil {
			t.Fatal(err)
		}
		f.WriteString("hello\n")
		f.Close()

		f, err = openOutput(path, true)
		if err != nil {
			t.Fatal(err)
		}
		f.WriteString("world\n")
		f.Close()

		data, _ := os.ReadFile(path)
		if string(data) != "hello\nworld\n" {
			t.Errorf("got %q, want %q", string(data), "hello\nworld\n")
		}
	})
}
