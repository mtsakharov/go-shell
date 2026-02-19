package main

import "strings"

func parseArgs(line string) []string {
	var args []string
	var current strings.Builder
	inSingle, inDouble := false, false

	flush := func() {
		if current.Len() > 0 {
			args = append(args, current.String())
			current.Reset()
		}
	}

	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case inSingle:
			if c == '\'' {
				inSingle = false
			} else {
				current.WriteByte(c)
			}
		case inDouble:
			if c == '"' {
				inDouble = false
			} else if c == '\\' && i+1 < len(line) {
				next := line[i+1]
				if next == '"' || next == '\\' || next == '$' {
					current.WriteByte(next)
					i++
				} else {
					current.WriteByte(c)
				}
			} else {
				current.WriteByte(c)
			}
		case c == '\'':
			inSingle = true
		case c == '"':
			inDouble = true
		case c == ' ' || c == '\t':
			flush()
		case c == '\\' && i+1 < len(line):
			current.WriteByte(line[i+1])
			i++
		default:
			current.WriteByte(c)
		}
	}
	flush()
	return args
}

// splitPipeline splits a parsed arg list on "|" tokens into segments
func splitPipeline(parts []string) [][]string {
	var segments [][]string
	var current []string
	for _, p := range parts {
		if p == "|" {
			if len(current) > 0 {
				segments = append(segments, current)
				current = nil
			}
		} else {
			current = append(current, p)
		}
	}
	if len(current) > 0 {
		segments = append(segments, current)
	}
	return segments
}
