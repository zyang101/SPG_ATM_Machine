package handler

import (
	"fmt"
	"SPG_ATM_Machine/utils"
)

func viewChoices()  {
    fmt.Println("Enter 0 to View Options Again")
    fmt.Println("Enter 1 to View Total ATM Cash")
	fmt.Println("Enter 2 to Deposit Cash to ATM")
	fmt.Println("Enter 3 to Withdraw Cash from ATM")
	fmt.Println("Enter 4 to Exit")
}

func Menu(username string) {
	fmt.Printf("\nWelcome Handler %s! What would you like do to today?\n", username)
	viewChoices()
	for {
		choice := utils.TypeInput("Enter your choice (0-4): ")
		switch choice {
		case "0":
			viewChoices()
		case "1":
			fmt.Println("ATM Total Balance is $1,234.56.")
		case "2":
			fmt.Println("Deposit feature coming soon!")
		case "3":
			fmt.Println("Withdrawal feature coming soon!")
		case "4":
			fmt.Println("Thank you for banking with JP Goldman Stanley!")
			return
		default:
			fmt.Println("Invalid option, please try again.")
		}
	}
}