package shell

import (
	"fmt"
	"io"
	"os"
)

type redirect struct {
	outFile   string
	outAppend bool
	errFile   string
	errAppend bool
}

// extractRedirect separates redirection tokens from command arguments.
func extractRedirect(parts []string) (args []string, r redirect) {
	for i := 0; i < len(parts); i++ {
		tok := parts[i]
		hasNext := i+1 < len(parts)
		switch {
		case (tok == ">>" || tok == "1>>") && hasNext:
			r.outFile, r.outAppend = parts[i+1], true
			i++
		case (tok == ">" || tok == "1>") && hasNext:
			r.outFile, r.outAppend = parts[i+1], false
			i++
		case tok == "2>>" && hasNext:
			r.errFile, r.errAppend = parts[i+1], true
			i++
		case tok == "2>" && hasNext:
			r.errFile, r.errAppend = parts[i+1], false
			i++
		default:
			args = append(args, tok)
		}
	}
	return
}

func openOutput(path string, appendMode bool) (*os.File, error) {
	if appendMode {
		return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}
	return os.Create(path)
}

// resolveStreams returns stdout and stderr writers based on redirection config.
func resolveStreams(r redirect) (stdout, stderr io.Writer) {
	stdout = os.Stdout
	stderr = os.Stderr
	if r.outFile != "" {
		f, err := openOutput(r.outFile, r.outAppend)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot open %s: %v\n", r.outFile, err)
		} else {
			stdout = f
		}
	}
	if r.errFile != "" {
		f, err := openOutput(r.errFile, r.errAppend)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot open %s: %v\n", r.errFile, err)
		} else {
			stderr = f
		}
	}
	return
}
