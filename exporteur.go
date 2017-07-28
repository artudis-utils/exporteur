package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Username: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Unexpected error reading from standard input.")
		fmt.Println(err)
	}
	username := strings.TrimSpace(input)
	fmt.Println(username)
}
