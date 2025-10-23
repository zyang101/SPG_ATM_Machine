package auth

import (
	"SPG_ATM_Machine/admin"
	"SPG_ATM_Machine/customer"
	"SPG_ATM_Machine/handler"
	"bufio"
	"fmt"
	"os"
	"strings"
	"golang.org/x/term"
)

func Login() (bool, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Enter username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Println("Enter PIN: ")
	bytePin, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("\nError reading password:", err)
		return false, ""
	}
	var success bool
	pin := strings.TrimSpace(string(bytePin))
	//hard codded for now
	if pin == "password"	{
		success = true
	}	else	{
		success = false
	}
	return success, username
}

func RouteUser(username string)	{
	//hard coded for now
	userType := "admin"
	switch userType {
	case "admin":
		admin.Menu(username)
	case "customer":
		customer.Menu(username)
	case "cash handler":
		handler.Menu(username)
	default:
		fmt.Println("Error validating user type")
	}
}