package shell

import (
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "simple command",
			input: "echo hello world",
			want:  []string{"echo", "hello", "world"},
		},
		{
			name:  "single quotes",
			input: "echo 'hello world'",
			want:  []string{"echo", "hello world"},
		},
		{
			name:  "double quotes",
			input: `echo "hello world"`,
			want:  []string{"echo", "hello world"},
		},
		{
			name:  "escaped space",
			input: `echo hello\ world`,
			want:  []string{"echo", "hello world"},
		},
		{
			name:  "escaped quote in double quotes",
			input: `echo "hello \"world\""`,
			want:  []string{"echo", `hello "world"`},
		},
		{
			name:  "escaped backslash in double quotes",
			input: `echo "hello\\world"`,
			want:  []string{"echo", `hello\world`},
		},
		{
			name:  "escaped dollar in double quotes",
			input: `echo "hello\$world"`,
			want:  []string{"echo", "hello$world"},
		},
		{
			name:  "non-special escape in double quotes preserved",
			input: `echo "hello\nworld"`,
			want:  []string{"echo", `hello\nworld`},
		},
		{
			name:  "mixed quotes",
			input: `echo 'hello'"world"`,
			want:  []string{"echo", "helloworld"},
		},
		{
			name:  "tabs as separator",
			input: "echo\thello\tworld",
			want:  []string{"echo", "hello", "world"},
		},
		{
			name:  "multiple spaces between args",
			input: "echo    hello    world",
			want:  []string{"echo", "hello", "world"},
		},
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "only spaces",
			input: "   ",
			want:  nil,
		},
		{
			name:  "pipe token",
			input: "echo hello | cat",
			want:  []string{"echo", "hello", "|", "cat"},
		},
		{
			name:  "redirection tokens",
			input: "echo hello > file.txt",
			want:  []string{"echo", "hello", ">", "file.txt"},
		},
		{
			name:  "single quote preserves special chars",
			input: "echo 'hello|world>foo'",
			want:  []string{"echo", "hello|world>foo"},
		},
		{
			name:  "backslash outside quotes",
			input: `echo hello\|world`,
			want:  []string{"echo", "hello|world"},
		},
		{
			name:  "adjacent quoted segments",
			input: `echo 'he'"ll"o`,
			want:  []string{"echo", "hello"},
		},
		{
			name:  "empty single quotes produce no arg",
			input: "echo ''hello",
			want:  []string{"echo", "hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseArgs(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseArgs(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSplitPipeline(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  [][]string
	}{
		{
			name:  "single command",
			parts: []string{"echo", "hello"},
			want:  [][]string{{"echo", "hello"}},
		},
		{
			name:  "two commands",
			parts: []string{"echo", "hello", "|", "cat"},
			want:  [][]string{{"echo", "hello"}, {"cat"}},
		},
		{
			name:  "three commands",
			parts: []string{"ls", "|", "grep", "go", "|", "head"},
			want:  [][]string{{"ls"}, {"grep", "go"}, {"head"}},
		},
		{
			name:  "leading pipe ignored",
			parts: []string{"|", "echo", "hello"},
			want:  [][]string{{"echo", "hello"}},
		},
		{
			name:  "consecutive pipes",
			parts: []string{"echo", "|", "|", "cat"},
			want:  [][]string{{"echo"}, {"cat"}},
		},
		{
			name:  "empty input",
			parts: []string{},
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitPipeline(tt.parts)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitPipeline(%v) = %v, want %v", tt.parts, got, tt.want)
			}
		})
	}
}
