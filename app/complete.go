package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type shellCompleter struct {
	lastInput string
	tabCount  int
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

func (sc *shellCompleter) Do(line []rune, pos int) ([][]rune, int) {
	input := string(line[:pos])

	if strings.ContainsAny(input, " \t") || input == "" {
		return nil, 0
	}

	if input != sc.lastInput {
		sc.lastInput = input
		sc.tabCount = 0
	}
	sc.tabCount++

	seen := map[string]bool{}
	var matches []string

	for _, b := range builtins {
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
		sc.tabCount = 0
		return [][]rune{[]rune(matches[0][len(input):] + " ")}, len(input)
	}

	lcp := longestCommonPrefix(matches)
	if len(lcp) > len(input) {
		sc.lastInput = lcp
		sc.tabCount = 0
		return [][]rune{[]rune(lcp[len(input):])}, len(input)
	}

	if sc.tabCount == 1 {
		fmt.Fprint(os.Stderr, "\x07")
		return nil, 0
	}

	fmt.Fprintf(os.Stdout, "\n%s\n$ %s", strings.Join(matches, "  "), input)
	sc.tabCount = 0
	return nil, 0
}
