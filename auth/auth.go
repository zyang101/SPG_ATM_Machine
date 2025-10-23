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
	bytePassword, _ := term.ReadPassword(int(os.Stdin.Fd()))
	password := string(bytePassword)
	fmt.Println()

	return username, password
}
