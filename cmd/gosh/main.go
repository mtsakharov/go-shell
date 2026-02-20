package main

import (
	"fmt"
	"os"

	"github.com/mtsakharov/go-shell/internal/shell"
)

func main() {
	sh, err := shell.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "shell: init error:", err)
		os.Exit(1)
	}

	os.Exit(sh.Run())
}
