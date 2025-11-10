package admin

import (
	"SPG_ATM_Machine/internal/api"
	"SPG_ATM_Machine/internal/db"
	"SPG_ATM_Machine/utils"
	"fmt"
	"strings"
	"strconv"
)

func viewChoices() {
	fmt.Println("Enter 0 to view options again")
	fmt.Println("Enter 1 to Create New Customer Account")
	fmt.Println("Enter 2 to View Deposits/Withdrawals")
	fmt.Println("Enter 3 to Set Deposit/Withdrawal limits")
	fmt.Println("Enter 4 to Exit")
}

func createNewUser() {
	fmt.Println("Let's create a new account for you.")

	var (
		newUsername         string
		newPin              string
		newName             string
		newDateOfBirth      string
		floatStartingAmount float64
	)

	for {
		newUsername = utils.TypeInput("Please enter a username: ")
		break
	}
	for {
		newPin = utils.TypeInput("Please enter a 6-digit PIN: ")
		if utils.ValidatePIN(newPin) {
			break
		}
	}
	for {
		newName = utils.TypeInput("Please enter your Name: ")
		if utils.ValidateName(newName) {
			break
		}
	}
	for {
		newDateOfBirth = utils.TypeInput("Please enter your date of birth (MM/DD/YYYY): ")
		if utils.ValidateDate(newDateOfBirth) {
			break
		}
	}
	for {
		startingAmount := utils.TypeInput("Starting Amount: ")
		amount, ok := utils.ParseAmount(startingAmount)
		if ok {
			floatStartingAmount = amount
			break
		}
	}
	
	database, err := db.Connect()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}
	defer database.Close()

	err = api.CreateUser(database, newName, newDateOfBirth, newPin, floatStartingAmount, newUsername, "customer")
	if err != nil {
		fmt.Println("Error creating user:", err)
		return
	}

	// Summary
	fmt.Println("\nAccount created successfully!")
	fmt.Println("Username:", newUsername)
	fmt.Println("PIN:", newPin)
	fmt.Println("Name:", newName)
	fmt.Println("Date of Birth:", newDateOfBirth)
	fmt.Printf("Starting Amount: $%.2f\n\n", floatStartingAmount)

}

func Menu(username string) {
	fmt.Printf("Welcome, Admin %s! What would you like do to today?\n", username)
	database, err := db.Connect()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}
	defer database.Close()
	
	viewChoices()
	for {
		choice := utils.TypeInput("Enter your choice (0-4): ")

		switch choice {
		case "0":
			viewChoices()
		case "1":
			createNewUser()
		case "2":
			err = api.ShowTransactions(database)
			if err != nil {
				fmt.Println("Error:", err)
			}
		case "3":
			withdrawalLimit, depositLimit, err := api.GetATMLimits(database)
			if err != nil {
				fmt.Println("Error fetching limits:", err)
			} else {
				fmt.Printf("\nCurrent Withdrawal Limit: $%.2f\n", withdrawalLimit)
				fmt.Printf("Current Deposit Limit: $%.2f\n\n", depositLimit)
			}			
			limitChoice := strings.ToUpper(utils.TypeInput("Enter W to change withdrawal limit, D to change deposit limit, or S to skip: "))
			switch limitChoice {
			case "W":
				limitStr := utils.TypeInput("Enter new withdrawal limit: ")
				newLimit, err := strconv.ParseFloat(limitStr, 64)
				if err != nil {
					fmt.Println("Invalid number. Please try again.")
					break
				}
				err = api.UpdateWithdrawalLimit(database, newLimit)
				if err != nil {
					fmt.Println("Error updating withdrawal limit:", err)
				}
			case "D":
				limitStr := utils.TypeInput("Enter new deposit limit: ")
				newLimit, err := strconv.ParseFloat(limitStr, 64)
				if err != nil {
					fmt.Println("Invalid number. Please try again.")
					break
				}
				err = api.UpdateDepositLimit(database, newLimit)
				if err != nil {
					fmt.Println("Error updating deposit limit:", err)
				}
			case "S":
				// skip
			default:
				fmt.Println("Invalid choice. Please enter W, D, or S.")
			}

		case "4":
			fmt.Println("Thank you for banking with JP Goldman Stanley!")
			return
		default:
			fmt.Println("Invalid option, please try again.")
		}
	}
}
