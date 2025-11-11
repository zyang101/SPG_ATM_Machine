package handler

import (
	"SPG_ATM_Machine/internal/api"
	"SPG_ATM_Machine/internal/db"
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

	//connects database ask team if I should close the connection shortly after
	database, err := db.Connect()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}

	//cash handler operation
	viewChoices()
	for {
		choice := utils.TypeInput("Enter your choice (0-4): ")
		switch choice {
		case "0":
			viewChoices()

		case "1": //gets balance of ATM
			bal,err := api.GetATMBalance(database)

			if err != nil {
				fmt.Println("Could not get balance:", err)
				return
			}

			fmt.Printf("ATM Total balance is $%.2f\n",bal)


		case "2": //deposits balance
			//utils.Deposit(username) //not sure what username is so I wrote the code to not needed it (can be addjusted later) 
			
			// amount_dep := utils.TypeInput("Enter amount to Deposit into the ATM: ")
			// amount,_ := utils.ParseAmount(amount_dep) // shouold only need atio now since only ints not floats
			// err := api.DepositATM(database, amount) TODO update to copy customer deposit function
			
			//checks if deposit passed
			// if err != nil {
			// 	fmt.Println("ERROR: ", err)
			// 	return
			// }

			//makes a new balance check
			api.PrintNewATMBalance(database)

		case "3"://withdaw from atm

			amountStr := utils.TypeInput("Enter amount to Withdraw from the ATM: ")
			amount, _ := utils.ParseAmount(amountStr)

			fmt.Println("Enter bill breakdown for withdrawal:")
			nHundreds := utils.TypeInt("Hundreds: ")
			nFifties  := utils.TypeInt("Fifties: ")
			nTwenties := utils.TypeInt("Twenties: ")
			nTens     := utils.TypeInt("Tens: ")
			nFives    := utils.TypeInt("Fives: ")
			nOnes     := utils.TypeInt("Ones: ")

			err := api.WithdrawATM(database, amount, nHundreds, nFifties, nTwenties, nTens, nFives, nOnes)
			if err != nil {
				fmt.Println("ERROR:", err)
				continue
			}

		case "4":
			fmt.Println("Thank you for banking with JP Goldman Stanley!")
			return
		default:
			fmt.Println("Invalid option, please try again.")
		}
	}
}