package main

import (
	"bufio"
	"fmt"
	_ "log"
	"os"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	fmt.Print("$ ")
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("$ ")
		// Wait for user input
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		fmt.Println(command[:len(command)-1] + ": command not found")
	}
}
