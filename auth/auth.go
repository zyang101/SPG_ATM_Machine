package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"golang.org/x/term"
)

func Login() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Enter password: ")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("\nError reading password:", err)
		return "", ""
	}
	password := strings.TrimSpace(string(bytePassword))
	fmt.Println()

	return username, password
}
