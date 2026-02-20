package shell

import "testing"

func TestLongestCommonPrefix(t *testing.T) {
	tests := []struct {
		name string
		strs []string
		want string
	}{
		{"empty", nil, ""},
		{"single", []string{"hello"}, "hello"},
		{"common prefix", []string{"echo", "exit"}, "e"},
		{"full match", []string{"abc", "abc"}, "abc"},
		{"no common", []string{"abc", "xyz"}, ""},
		{"partial", []string{"foobar", "foobaz", "fooqux"}, "foo"},
		{"one empty", []string{"", "hello"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := longestCommonPrefix(tt.strs)
			if got != tt.want {
				t.Errorf("longestCommonPrefix(%v) = %q, want %q", tt.strs, got, tt.want)
			}
		})
	}
}
