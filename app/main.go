package main

import (
	"bufio"
	"fmt"
	_ "log"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("$ ")

		line, err := reader.ReadString('\n')
		if err != nil {
			os.Exit(0)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)

		switch parts[0] {
		case "exit":
			return

		case "echo":
			if len(parts) > 1 {
				fmt.Println(strings.Join(parts[1:], " "))
			} else {
				fmt.Println()
			}

		case "type":
			if len(parts) < 2 {
				continue
			}

			switch parts[1] {
			case "echo", "exit", "type":
				fmt.Printf("%s is a shell builtin\n", parts[1])
			default:
				fmt.Printf("%s: not found\n", parts[1])
			}

		default:
			fmt.Println(parts[0] + ": command not found")
		}
	}
}
