package main

import (
	"os"
	"path/filepath"
	"strings"
)

func findInPath(cmd string) string {
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		full := filepath.Join(dir, cmd)
		if info, err := os.Stat(full); err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
			return full
		}
	}
	return ""
}

func executablesInPath(prefix string) []string {
	seen := map[string]bool{}
	var results []string
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if seen[e.Name()] || !strings.HasPrefix(e.Name(), prefix) {
				continue
			}
			if info, err := e.Info(); err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
				results = append(results, e.Name())
				seen[e.Name()] = true
			}
		}
	}
	return results
}
