package customer

import (
	"SPG_ATM_Machine/internal/api"
	"SPG_ATM_Machine/internal/db"
	"SPG_ATM_Machine/utils"
	"fmt"
	"strconv"
	"strings"
)

func viewChoices() {
	fmt.Println("Enter 0 to view options again")
	fmt.Println("Enter 1 to Check Balance")
	fmt.Println("Enter 2 to Deposit Money")
	fmt.Println("Enter 3 to Withdraw Money")
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
		choice := utils.TypeInput("Enter your choice (0-5): ")
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
			fmt.Printf("Enter the quantity of each denomination you're depositing in deposit.txt \n")
			fmt.Printf("Each line is the next higher denomination 1,5,10,20,50,100, e.g. a 3 on line 6 is $300 \n")
			utils.TypeInput("Press enter here when you are ready to continue:")

			input_denoms, err := utils.ParseDeposit("customer/deposit.txt")

			if err != nil {
				fmt.Println("Invalid Input:", err)
				continue
			}
			err = api.DepositATM(database, input_denoms)
			if err != nil {
				fmt.Println("Could not update ATM balance:", err)
				continue
			}

			denominations_values := []int{1, 5, 10, 20, 50, 100}
			deposit := 0
			for denom := range input_denoms {
				deposit += input_denoms[denom] * denominations_values[denom]
			}

			newBalance, err := api.DepositBalance(database, username, float64(deposit))
			if err != nil {
				fmt.Println("Could not update balance:", err)
				continue
			}
			fmt.Printf("Your new balance is $%.2f \n", newBalance)
		case "3":

			moneyWithdraw := utils.TypeInput("Enter how much money to withdraw: ")
			amount, err := strconv.ParseFloat(moneyWithdraw, 64)
			if err != nil {
				fmt.Println("Invalid Input:", err)
				continue
			}

			fmt.Println("Enter bill breakdown for withdrawal:")
			nHundreds := utils.TypeInt("Hundreds: ")
			nFifties := utils.TypeInt("Fifties: ")
			nTwenties := utils.TypeInt("Twenties: ")
			nTens := utils.TypeInt("Tens: ")
			nFives := utils.TypeInt("Fives: ")
			nOnes := utils.TypeInt("Ones: ")

			denoms := []int{nOnes, nFives, nTens, nTwenties, nFifties, nHundreds}

			err = api.WithdrawATM(database, amount, nHundreds, nFifties, nTwenties, nTens, nFives, nOnes)
			if err != nil {
				fmt.Println("ERROR:", err)
				continue
			}

			newBalance, err := api.WithdrawBalance(database, username, float64(amount))
			if err != nil {
				_ = api.DepositATM(database, denoms)
				fmt.Println("Transaction failed, withdrawal cancelled")
				fmt.Println("Could not update balance:", err)
				continue
			}
			fmt.Printf("Your new balance is $%.2f \n", newBalance)

		case "4":
			var transferTarget string
			var transferAmt float64
			for {
				transferTarget = utils.TypeInput("Enter username to transfer funds to: ")
				break
			}
			stmtCheck, err := database.Prepare("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)")
			if err != nil {
				fmt.Println("Specified username does not exist")
				continue
			}
			defer stmtCheck.Close()

			//Error handling to check if the username exists or not
			var exists bool
			if err := stmtCheck.QueryRow(transferTarget).Scan(&exists); err != nil {
				fmt.Printf("failed to check username: %v\n", err)
				continue
			}
			if !exists {
				fmt.Printf("user '%s' does not exist\n", transferTarget)
				continue
			}

			var role string
			err = database.QueryRow("SELECT role FROM users WHERE username = ?", transferTarget).Scan(&role)
			if err != nil {
				fmt.Printf("failed to check user role: %v\n", err)
				continue
			}
			if role != "customer" {
				fmt.Println("Specified username does not exist")
				continue
			}

			for {
				transferAmtStr := utils.TypeInput("Enter amount to transfer: ")
				amount, ok := utils.ParseAmount(transferAmtStr)
				if ok {
					transferAmt = amount
					break
				}
			}
			balance, err := api.GetUserBalance(database, username)
			if err != nil {
				fmt.Println("Could not get balance:", err)
				continue
			}

			if transferAmt > balance {
				fmt.Printf("Invalid transfer amount. Your current balance is: '%.2f'\n", balance)
				continue
			}

			for {
				answer := strings.ToUpper(utils.TypeInput(fmt.Sprintf("Confirm transfer of '%.2f' from '%s' to '%s'? (Y/N)", transferAmt, username, transferTarget)))
				if answer == "Y" {
					if err := api.TransferFunds(database, username, transferTarget, transferAmt); err != nil {
						fmt.Printf("Transfer failed: %v\n", err)
						continue
					}
					fmt.Println("Transfer success")
					break
				} else if answer == "N" {
					fmt.Println("Transfer cancelled.")
					break
				} else {
					fmt.Println("Please answer Y or N.")
				}
			}

		case "5":
			fmt.Println("Thank you for banking with JP Goldman Stanley!")
			return
		default:
			fmt.Println("Invalid option, please try again.")
		}
	}
}
