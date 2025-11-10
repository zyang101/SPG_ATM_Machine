package customer

import (
	"SPG_ATM_Machine/internal/api"
	"SPG_ATM_Machine/internal/db"
	"SPG_ATM_Machine/utils"
	"fmt"
	"strconv"
)

func viewChoices() {
	fmt.Println("Enter 0 to view options again")
	fmt.Println("Enter 1 to Check Balance")
	fmt.Println("Enter 2 to Deposit Money")
	fmt.Println("Enter 3 to Withdrawal Money")
	fmt.Println("Enter 4 to Transfer Funds")
	fmt.Println("Enter 5 to Exit")
}

func Menu(username string) {
	fmt.Printf("\nWelcome %s! What would you like do to today?\n", username)
	database, err := db.Connect()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}
	viewChoices()
	for {
		choice := utils.TypeInput("Enter your choice (0-4): ")
		switch choice {
		case "0":
			viewChoices()
		case "1":
			balance, err := api.GetUserBalance(database, username)
			if err != nil {
				fmt.Println("Could not get balance:", err)
				continue
			}
			fmt.Printf("Your balance is $%.2f \n", balance)
		case "2":
			moneyDeposit := utils.TypeInput("Enter how much money to deposit: ")
			val, err := strconv.ParseFloat(moneyDeposit, 64)
			if err != nil {
				fmt.Println("Invalid Input:", err)
				continue
			}
			newBalance, err := api.DepositBalance(database, username, float64(val))
			if err != nil {
				fmt.Println("Could not update balance:", err)
				continue
			}
			fmt.Printf("Your new blanace is $%.2f \n", newBalance)
		case "3":
			moneyWithdraw := utils.TypeInput("Enter how much money to withdraw: ")
			val, err := strconv.ParseFloat(moneyWithdraw, 64)
			if err != nil {
				fmt.Println("Invalid Input:", err)
				continue
			}
			newBalance, err := api.WithdrawBalance(database, username, float64(val))
			if err != nil {
				fmt.Println("Could not update balance:", err)
				continue
			}
			fmt.Printf("Your new balance is $%.2f \n", newBalance)
		case "4":
			for {
				transferTarget := utils.TypeInput("Enter username to transfer funds to: ")
				if utils.ValidateName(transferTarget) {
					break
				}
			}
			stmtCheck, err := db.Prepare("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)")
			if err != nil {
				return err
			}
			defer stmtCheck.Close()

			//Error handling to check if the username exists or not
			var exists bool
			if err := stmtCheck.QueryRow(username).Scan(&exists); err != nil {
				return fmt.Errorf("failed to check username: %v", err)
			}
			if exists {
				return fmt.Errorf("username '%s' already exists", username)
			}

		case "5":
			fmt.Println("Thank you for banking with JP Goldman Stanley!")
			return
		default:
			fmt.Println("Invalid option, please try again.")
		}
	}
}
