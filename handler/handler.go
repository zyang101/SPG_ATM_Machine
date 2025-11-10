package handler

import (
	//"SPG_ATM_Machine/internal/api"
	"database/sql"
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

//CHECK WITH TEAM TO SEE IF THEY WANT TO PUT THIS IN HANDALER/ IF THE DB QUERY WAS DONE CORRECTLY
func GetATMBalance(db *sql.DB) (float64, error) {

	var bal float64
	err := db.QueryRow("SELECT balance FROM atm WHERE id = 1").Scan(&bal)

	if err != nil {
		return 0, err
	}

	return bal, nil

}

func PrintNewATMBalance(db *sql.DB) { //prints balance after operation

	new_bal,err := GetATMBalance(db)
	if err != nil {
		fmt.Println("Could not get balance:", err)
		return
	}

	fmt.Printf("New ATM balance: $%.2f\n", new_bal)
}

func DepositATM(db *sql.DB, inc_amount float64) error {

	_, err := db.Exec("UPDATE atm SET balance = balance + ? WHERE id = 1", inc_amount)

	if err != nil {
		return fmt.Errorf("failed to update ATM balance: %v", err)
	}

	return nil
}

func WithdrawATM (db *sql.DB, dec_anount float64) error {

	bal, err := GetATMBalance(db)

	if err != nil {
		return fmt.Errorf("could not get ATM balance: %v", err)
	}

	if dec_anount > bal {
		return fmt.Errorf("ATM does not have enough cash. Current ATM balance: $%.2f", bal)
	}

	_, err = db.Exec("UPDATE atm SET balance = balance - ? WHERE id = 1", dec_anount)
	if err != nil {
		return fmt.Errorf("failed to withdraw from ATM balance: %v", err)
	}

	return nil
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
			bal,err := GetATMBalance(database)

			if err != nil {
				fmt.Println("Could not get balance:", err)
				return
			}

			fmt.Printf("ATM Total balance is $%.2f\n",bal)


		case "2": //deposits balance
			//utils.Deposit(username) //not sure what username is so I wrote the code to not needed it (can be addjusted later) 
			
			amount_dep := utils.TypeInput("Enter amount to Deposit into the ATM: ")
			amount,_ := utils.ParseAmount(amount_dep)
			err := DepositATM(database, amount)
			
			//checks if deposit passed
			if err != nil {
				fmt.Println("ERROR: ", err)
				return
			}

			//makes a new balance check
			PrintNewATMBalance(database)

		case "3"://withdaw from atm

			amount_with := utils.TypeInput("Enter amount to Withdraw from the ATM: ")
			amount,_ := utils.ParseAmount(amount_with)

			err := WithdrawATM(database, amount)

			//checks if withdraw passed
			if err != nil {
				fmt.Println("ERROR: ", err)
				return
			}

			//makes a new balance check
			PrintNewATMBalance(database)

		case "4":
			fmt.Println("Thank you for banking with JP Goldman Stanley!")
			return
		default:
			fmt.Println("Invalid option, please try again.")
		}
	}
}