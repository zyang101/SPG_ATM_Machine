package admin

import (
    "fmt"
    "SPG_ATM_Machine/utils"
)


func viewChoices()  {
    fmt.Println("Enter 0 to view options again")
    fmt.Println("Enter 1 to Create New Customer Account")
	fmt.Println("Enter 2 to View Deposits/Withdrawals")
	fmt.Println("Enter 3 to Set Deposit/Withdrawal limits")
	fmt.Println("Enter 4 to Exit")
}

func createNewUser()    {
    fmt.Println("Let's create a new account for you.")
    var newUsername string
    for {
        newUsername = utils.TypeInput("Please enter a username: ")
        //hard coded for now, need to query db for this
        exists := false
        if exists {
            fmt.Println("That username already exists. Please choose another.")
        } else {
            break
        }
    }
    newPin := utils.TypeInput("Please enter a PIN: ")
    newName := utils.TypeInput("Please enter your Name: ")
    newDateOfBirth := utils.TypeInput("Please enter your date of birth (mm/dd/yr): ")
    startingAmount := utils.TypeInput("Starting Amount: ")

    fmt.Println("Account created successfully!")
    fmt.Println("Username: ", newUsername)
    fmt.Println("PIN: ", newPin)
    fmt.Println("Name: ", newName)
    fmt.Println("Date of Birth: ", newDateOfBirth)
    fmt.Println("Starting Amount: ", startingAmount)
    // update DB
}

func Menu(username string) {
    fmt.Printf("Welcome, Admin %s! What would you like do to today?\n", username)
    viewChoices()
	for {
		choice := utils.TypeInput("Enter your choice (0-4): ")

		switch choice {
        case "0":
            viewChoices()
		case "1":
            createNewUser()
		case "2":
			fmt.Println("View Deposit/Withdrawl feature coming soon!")
		case "3":
			fmt.Println("Set Deposit/Withdrawal feature coming soon!")
		case "4":
			fmt.Println("Thank you for banking with JP Goldman Stanley!")
			return
		default:
			fmt.Println("Invalid option, please try again.")
		}
	}
}