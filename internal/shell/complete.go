package shell

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

type completer struct {
	lastInput string
	tabCount  int
}

func newCompleter() readline.AutoCompleter {
	return &completer{}
}

func (c *completer) Do(line []rune, pos int) ([][]rune, int) {
	input := string(line[:pos])

	if strings.ContainsAny(input, " \t") || input == "" {
		return nil, 0
	}

	if input != c.lastInput {
		c.lastInput = input
		c.tabCount = 0
	}
	c.tabCount++

	seen := map[string]bool{}
	var matches []string

	for _, b := range builtinNames {
		if strings.HasPrefix(b, input) && !seen[b] {
			matches = append(matches, b)
			seen[b] = true
		}
	}
	for _, name := range executablesInPath(input) {
		if !seen[name] {
			matches = append(matches, name)
			seen[name] = true
		}
	}

	if len(matches) == 0 {
		fmt.Fprint(os.Stderr, "\x07")
		return nil, 0
	}

	sort.Strings(matches)

	if len(matches) == 1 {
		c.tabCount = 0
		return [][]rune{[]rune(matches[0][len(input):] + " ")}, len(input)
	}

	lcp := longestCommonPrefix(matches)
	if len(lcp) > len(input) {
		c.lastInput = lcp
		c.tabCount = 0
		return [][]rune{[]rune(lcp[len(input):])}, len(input)
	}

	if c.tabCount == 1 {
		fmt.Fprint(os.Stderr, "\x07")
		return nil, 0
	}

	fmt.Fprintf(os.Stdout, "\n%s\n$ %s", strings.Join(matches, "  "), input)
	c.tabCount = 0
	return nil, 0
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strs[0]
	for _, s := range strs[1:] {
		for !strings.HasPrefix(s, prefix) {
			prefix = prefix[:len(prefix)-1]
			if prefix == "" {
				return ""
			}
		}
	}
	return prefix
}
