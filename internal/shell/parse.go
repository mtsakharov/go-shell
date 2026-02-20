package shell

import "strings"

// parseArgs splits a command line into arguments, respecting single quotes,
// double quotes, and backslash escapes.
func parseArgs(line string) []string {
	var args []string
	var cur strings.Builder
	inSingle, inDouble := false, false

	flush := func() {
		if cur.Len() > 0 {
			args = append(args, cur.String())
			cur.Reset()
		}
	}

	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case inSingle:
			if c == '\'' {
				inSingle = false
			} else {
				cur.WriteByte(c)
			}
		case inDouble:
			if c == '"' {
				inDouble = false
			} else if c == '\\' && i+1 < len(line) {
				next := line[i+1]
				if next == '"' || next == '\\' || next == '$' {
					cur.WriteByte(next)
					i++
				} else {
					cur.WriteByte(c)
				}
			} else {
				cur.WriteByte(c)
			}
		case c == '\'':
			inSingle = true
		case c == '"':
			inDouble = true
		case c == ' ' || c == '\t':
			flush()
		case c == '\\' && i+1 < len(line):
			cur.WriteByte(line[i+1])
			i++
		default:
			cur.WriteByte(c)
		}
	}
	flush()
	return args
}

// splitPipeline divides a parsed argument list on "|" tokens into segments.
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
