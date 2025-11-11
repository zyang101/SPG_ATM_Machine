package api

import (
	"SPG_ATM_Machine/internal/models"
	"database/sql"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Create the user
func CreateUser(db *sql.DB, fullName, dob, pin string, startingBal float64, username, role string) error {
	//Check database to see if it exist (use prepare statement to separate code and data)
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

	//Hashes pin to store in database
	hashedPin, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash PIN: %v", err)
	}

	//Update id number for each user
	nextID, err := GetNextUserID(db)
	if err != nil {
		return err
	}

	//Upload all USER metadata into database
	stmtInsert, err := db.Prepare(`
		INSERT INTO users (id, full_name, dob, pin, starting_bal, username, role)
		VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmtInsert.Close()

	_, err = stmtInsert.Exec(nextID, fullName, dob, string(hashedPin), startingBal, username, role)
	return err
}

// Increments the userID by one each time a new user is created
func GetNextUserID(db *sql.DB) (int, error) {
	//Gets most recent userID number
	stmt, err := db.Prepare("SELECT COUNT(*) FROM users")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	//Increments the userID by one and returns it
	var count int
	err = stmt.QueryRow().Scan(&count)
	if err != nil {
		return 0, err
	}
	return count + 1, nil
}

// Gets the User's current balance
func GetUserBalance(db *sql.DB, username string) (float64, error) {
	//Retrieve the current balance of the user
	stmt, err := db.Prepare("SELECT starting_bal FROM users WHERE username = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	//Error handling then returns the balance
	var bal float64
	err = stmt.QueryRow(username).Scan(&bal)
	if err != nil {
		return 0, err
	}
	return bal, nil
}

// Depost money to the user's account
func DepositBalance(db *sql.DB, username string, amount float64) (float64, error) {
	//Retrieve the user's current balance
	balance, err := GetUserBalance(db, username)
	if err != nil {
		return 0, fmt.Errorf("could not get balance: %v", err)
	}

	//Gets the withdraw and deposit limits and do error handling
	_, depositLimit, err := GetATMLimits(db)
	if err != nil {
		fmt.Println("Error fetching limits:", err)
	} else {
		if amount > depositLimit {
			return 0, fmt.Errorf("your deposit amount %f is over the deposit limit: %f", amount, depositLimit)
		}
	}

	//Find new balance
	newBalance := balance + amount

	//Update the user's balance with the new balance
	stmtUpdUser, err := db.Prepare("UPDATE users SET starting_bal = ? WHERE username = ?")
	if err != nil {
		return 0, err
	}
	defer stmtUpdUser.Close()

	if _, err = stmtUpdUser.Exec(newBalance, username); err != nil {
		return 0, fmt.Errorf("failed to update balance: %v", err)
	}

	//Updates the atm's balance with the newly deposited money
	stmtUpdATM, err := db.Prepare("UPDATE atm SET balance = balance + ? WHERE id = 1")
	if err != nil {
		return 0, err
	}
	defer stmtUpdATM.Close()

	if _, err = stmtUpdATM.Exec(amount); err != nil {
		return 0, fmt.Errorf("failed to update ATM balance: %v", err)
	}

	userID, err := GetUserID(db, username)
	if err != nil {
		return newBalance, fmt.Errorf("could not get user id: %v", err)
	}

	//Update transaction log
	stmtTrans, err := db.Prepare(`
		INSERT INTO transactions (user_id, date, balance)
		VALUES (?, datetime('now', 'localtime'), ?)`)
	if err != nil {
		return newBalance, err
	}
	defer stmtTrans.Close()

	_, err = stmtTrans.Exec(userID, amount)
	if err != nil {
		return newBalance, fmt.Errorf("failed to log transaction: %v", err)
	}

	return newBalance, err
}

// Withdraw money from the user's account
func WithdrawBalance(db *sql.DB, username string, amount float64) (float64, error) {
	//Get the user's current balance
	balance, err := GetUserBalance(db, username)
	if err != nil {
		return 0, fmt.Errorf("could not get balance: %v", err)
	}

	//Check if the user has enough money to withdraw the amount
	withdrawLimit, _, err := GetATMLimits(db)
	if err != nil {
		fmt.Println("Error fetching limits:", err)
	} else {
		if amount > withdrawLimit {
			return 0, fmt.Errorf("your withdrawl amount %f is over the withdrawl limit: %f", amount, withdrawLimit)
		}
	}

	bal, err := GetATMBalance(db)
	if err != nil {
		fmt.Println("Error checking total cash in ATM:", err)
	} else {
		if amount > bal {
			return 0, fmt.Errorf("your withdrawl amount %f is over the ATM balance: %f", amount, bal)
		}
	}

	//Find the new balance after withdraw amount
	newBalance := balance - amount
	if newBalance < 0 {
		return 0, fmt.Errorf("not enough in balance to withdraw. Current balance: $%.2f", balance)
	}

	//Update the user's balance
	stmtUpdUser, err := db.Prepare("UPDATE users SET starting_bal = ? WHERE username = ?")
	if err != nil {
		return 0, err
	}
	defer stmtUpdUser.Close()

	if _, err = stmtUpdUser.Exec(newBalance, username); err != nil {
		return 0, fmt.Errorf("failed to update balance: %v", err)
	}

	//Update the atm's balance
	stmtUpdATM, err := db.Prepare("UPDATE atm SET balance = balance - ? WHERE id = 1")
	if err != nil {
		return 0, err
	}
	defer stmtUpdATM.Close()

	if _, err = stmtUpdATM.Exec(amount); err != nil {
		return 0, fmt.Errorf("failed to update ATM balance: %v", err)
	}

	userID, err := GetUserID(db, username)
	if err != nil {
		return newBalance, fmt.Errorf("could not get user id: %v", err)
	}

	//Update transaction log
	stmtTrans, err := db.Prepare(`
		INSERT INTO transactions (user_id, date, balance)
		VALUES (?, datetime('now', 'localtime'), ?)`)
	if err != nil {
		return newBalance, err
	}
	defer stmtTrans.Close()

	_, err = stmtTrans.Exec(userID, -amount)
	if err != nil {
		return newBalance, fmt.Errorf("failed to log transaction: %v", err)
	}

	return newBalance, err
}

// Get the user's ID based on username
func GetUserID(db *sql.DB, username string) (int, error) {
	//Pull ID from database with username as input
	stmt, err := db.Prepare("SELECT id FROM users WHERE username = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(username).Scan(&id)
	return id, err
}

// Transfer funds from source user to target user
func TransferFunds(db *sql.DB, sourceUser string, targetUser string, amount float64) error {
	//Get the source user's current balance
	sourceBalance, err := GetUserBalance(db, sourceUser)
	if err != nil {
		return fmt.Errorf("could not get source balance: %v", err)
	}

	//Get the target user's current balance
	targetBalance, err := GetUserBalance(db, targetUser)
	if err != nil {
		return fmt.Errorf("could not get target balance: %v", err)
	}

	//Find the new source balance after withdraw amount
	newSourceBalance := sourceBalance - amount
	if newSourceBalance < 0 {
		return fmt.Errorf("not enough in balance to withdraw. Current balance: $%.2f", sourceBalance)
	}

	//Find the new target balance after withdraw amount
	newTargetBalance := targetBalance + amount

	// Start a transaction to update both balances or none
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback() // Will rollback if we exit the function early

	stmtUpdUser, err := tx.Prepare("UPDATE users SET starting_bal = ? WHERE username = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare transfer transaction: %v", err)
	}
	defer stmtUpdUser.Close()

	if _, err = stmtUpdUser.Exec(newSourceBalance, sourceUser); err != nil {
		return fmt.Errorf("failed to update source user balance: %v", err)
	}

	if _, err = stmtUpdUser.Exec(newTargetBalance, targetUser); err != nil {
		return fmt.Errorf("failed to update target user balance: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// List all the current users inside the databases
func ListUsers(db *sql.DB) ([]models.User, error) {
	//Pull entire list of users in Database
	stmt, err := db.Prepare(`SELECT id, full_name, dob, pin, starting_bal, username, role FROM users`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//Add each user to a list and returns it
	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.FullName, &u.DOB, &u.PIN, &u.StartingBal, &u.Username, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// Show the list of transactions
func ShowTransactions(db *sql.DB) error {
	//Retrieves all transcations
	stmt, err := db.Prepare(`
		SELECT t.id, u.username, t.date, t.balance
		FROM transactions t
		LEFT JOIN users u ON t.user_id = u.id
		ORDER BY t.id ASC`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return fmt.Errorf("failed to query transactions: %v", err)
	}
	defer rows.Close()

	//Print out each of the transactions
	fmt.Println("\n===== TRANSACTION HISTORY =====")
	fmt.Printf("%-5s | %-15s | %-20s | %-10s\n", "ID", "Username", "Date", "Amount ($)")
	fmt.Println(strings.Repeat("-", 60))

	for rows.Next() {
		var id int
		var username, date string
		var amount float64
		if err := rows.Scan(&id, &username, &date, &amount); err != nil {
			return fmt.Errorf("failed to scan transaction: %v", err)
		}
		fmt.Printf("%-5d | %-15s | %-20s | %10.2f\n", id, username, date, amount)
	}

	return rows.Err()
}

// Update the withdrawal limit
func UpdateWithdrawalLimit(db *sql.DB, newLimit float64) error {
	if newLimit < 0 {
		return fmt.Errorf("withdrawal limit cannot be negative")
	}
	stmt, err := db.Prepare("UPDATE atm SET withdrawal_limit = ? WHERE id = 1")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(newLimit)
	return err
}

// Update the deposit limit
func UpdateDepositLimit(db *sql.DB, newLimit float64) error {
	if newLimit < 0 {
		return fmt.Errorf("deposit limit cannot be negative")
	}
	stmt, err := db.Prepare("UPDATE atm SET deposit_limit = ? WHERE id = 1")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(newLimit)
	return err
}

// Retrieve the ATM deposit and withdrawal limits
func GetATMLimits(db *sql.DB) (float64, float64, error) {
	stmt, err := db.Prepare(`SELECT withdrawal_limit, deposit_limit FROM atm WHERE id = 1`)
	if err != nil {
		return 0, 0, err
	}
	defer stmt.Close()

	var withdrawalLimit, depositLimit float64
	if err := stmt.QueryRow().Scan(&withdrawalLimit, &depositLimit); err != nil {
		return 0, 0, err
	}
	return withdrawalLimit, depositLimit, nil
}

// Get the ATM's current balance
func GetATMBalance(db *sql.DB) (float64, error) {
	stmt, err := db.Prepare("SELECT balance FROM atm WHERE id = 1")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var bal float64
	err = stmt.QueryRow().Scan(&bal)
	return bal, err
}

// Print the ATM's new balance after updating it
func PrintNewATMBalance(db *sql.DB) {
	newBal, err := GetATMBalance(db)
	if err != nil {
		fmt.Println("Could not get balance:", err)
		return
	}
	fmt.Printf("New ATM balance: $%.2f\n", newBal)
}

// Withdraw money from the atm from the Cash Handler
func WithdrawATM(db *sql.DB, dec_amount float64, nHundreds, nFifties, nTwenties, nTens, nFives, nOnes int) error {
	bal, err := GetATMBalance(db)
	if err != nil {
		return fmt.Errorf("could not get ATM balance: %v", err)
	}

	if dec_amount > bal {
		return fmt.Errorf("ATM does not have enough cash. Current ATM balance: $%.2f", bal)
	}

	// Get current bill counts
	row := db.QueryRow(`
		SELECT ones, fives, tens, twenties, fifties, hundreds
		FROM atm WHERE id = 1;
	`)

	var ones, fives, tens, twenties, fifties, hundreds int
	err = row.Scan(&ones, &fives, &tens, &twenties, &fifties, &hundreds)
	if err != nil {
		return fmt.Errorf("failed fetching ATM denominations: %v", err)
	}

	// Total from input
	withdrawTotal := (nHundreds * 100) + (nFifties * 50) + (nTwenties * 20) +
		(nTens * 10) + (nFives * 5) + (nOnes * 1)

	if withdrawTotal != int(dec_amount) {
		return fmt.Errorf("bills selected ($%d) do not match withdrawal amount $%.2f", withdrawTotal, dec_amount)
	}

	// Check availability
	if nHundreds > hundreds || nFifties > fifties || nTwenties > twenties ||
		nTens > tens || nFives > fives || nOnes > ones {
		return fmt.Errorf("ATM does not have enough of one or more bill denominations")
	}

	// Deduct bills
	hundreds -= nHundreds
	fifties -= nFifties
	twenties -= nTwenties
	tens -= nTens
	fives -= nFives
	ones -= nOnes

	// Update DB
	stmt, err := db.Prepare(`
		UPDATE atm
		SET balance = balance - ?,
		    ones = ?, fives = ?, tens = ?, twenties = ?, fifties = ?, hundreds = ?
		WHERE id = 1`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(dec_amount, ones, fives, tens, twenties, fifties, hundreds)
	if err != nil {
		return fmt.Errorf("failed to withdraw bills: %v", err)
	}

	return nil
}

// Withdraw money from the atm from the Cash Handler
func DepositATM(db *sql.DB, denoms []int) error {

	// Get current bill counts
	row := db.QueryRow(`
		SELECT ones, fives, tens, twenties, fifties, hundreds
		FROM atm WHERE id = 1;
	`)

	var ones, fives, tens, twenties, fifties, hundreds int
	err := row.Scan(&ones, &fives, &tens, &twenties, &fifties, &hundreds)
	if err != nil {
		return fmt.Errorf("failed fetching ATM denominations: %v", err)
	}

	// Add bills
	hundreds += denoms[5]
	fifties += denoms[4]
	twenties += denoms[3]
	tens += denoms[2]
	fives += denoms[1]
	ones += denoms[0]

	depositTotal := (denoms[5] * 100) + (denoms[4] * 50) + (denoms[3] * 20) + (denoms[2] * 10) + (denoms[1] * 5) + (denoms[0] * 1)

	// Update DB
	stmt, err := db.Prepare(`
		UPDATE atm
		SET balance = balance + ?,
		    ones = ?, fives = ?, tens = ?, twenties = ?, fifties = ?, hundreds = ?
		WHERE id = 1`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(depositTotal, ones, fives, tens, twenties, fifties, hundreds)
	if err != nil {
		return fmt.Errorf("failed to deposit bills: %v", err)
	}

	return nil
}
