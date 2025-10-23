package customer

import	(
	"fmt"
	"strings"

)

func typeInput(prompt string) string {
	var input string
	fmt.Print(prompt)
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}

func Menu(username string)	{
	fmt.Printf("\nWelcome %s! What would you like do to today?\n", username)
	fmt.Println("Enter 1 to Check Balance")
	fmt.Println("Enter 2 to Deposit Money")
	fmt.Println("Enter 3 to Withdrawal Money")
	fmt.Println("Enter 4 to Exit")
	for {
		choice := typeInput("Enter your choice (1-4): ")

		switch choice {
		case "1":
			fmt.Println("Your balance is $1,234.56.")
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