package main

import (
	"fmt"
	"strings"
	"SPG_ATM_Machine/auth"
)

func typeInput(prompt string) string {
	var input string
	fmt.Print(prompt)
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}

func main() {
	fmt.Println("Welcome to JP Goldman Stanley ATM!")

	for {
		answer := strings.ToUpper(typeInput("Are you an existing customer? (Y/N): "))

		if answer == "Y" {
			fmt.Println("Great! Please log in.")
			username, password := auth.Login()
			fmt.Println("You entered:", username, password)
			break

		} else if answer == "N" {
			fmt.Println("Let's create a new account for you.")

			var newUsername string
			for {
				newUsername = typeInput("Please enter a username: ")

				exists := false
				if exists {
					fmt.Println("That username already exists. Please choose another.")
				} else {
					break
				}
			}

			newPassword := typeInput("Please enter a password: ")
			fmt.Println("Account created successfully!")
			fmt.Println("Username:", newUsername)
			fmt.Println("Password:", newPassword)

			fmt.Println("\nPlease log in to your new account:")
			username, password := auth.Login()
			fmt.Println("You entered:", username, password)
			break

		} else {
			fmt.Println("Please answer Y or N.")
		}
	}
}
