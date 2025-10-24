package customer

import	(
	"fmt"
	"SPG_ATM_Machine/utils"
)

func viewChoices()  {
    fmt.Println("Enter 0 to view options again")
	fmt.Println("Enter 1 to Check Balance")
	fmt.Println("Enter 2 to Deposit Money")
	fmt.Println("Enter 3 to Withdrawal Money")
	fmt.Println("Enter 4 to Exit")
}

func Menu(username string)	{
	fmt.Printf("\nWelcome %s! What would you like do to today?\n", username)
	viewChoices()
	for {
		choice := utils.TypeInput("Enter your choice (0-4): ")
		switch choice {
		case "0":
			viewChoices()
		case "1":
			fmt.Println("Your balance is $1,234.56.")
		case "2":
			utils.Deposit(username)
		case "3":
			utils.Withdraw(username)
		case "4":
			fmt.Println("Thank you for banking with JP Goldman Stanley!")
			return
		default:
			fmt.Println("Invalid option, please try again.")
		}
	}
}