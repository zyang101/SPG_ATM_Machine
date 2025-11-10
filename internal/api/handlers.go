package api

import (
	"SPG_ATM_Machine/internal/models"
	"database/sql"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Create a new user
func CreateUser(db *sql.DB, fullName, dob, pin string, startingBal float64, username, role string) error {
	stmtCheck, err := db.Prepare("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)")
	if err != nil {
		return err
	}
	defer stmtCheck.Close()

	var exists bool
	if err := stmtCheck.QueryRow(username).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check username: %v", err)
	}
	if exists {
		return fmt.Errorf("username '%s' already exists", username)
	}

	hashedPin, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash PIN: %v", err)
	}

	nextID, err := GetNextUserID(db)
	if err != nil {
		return err
	}

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

func GetNextUserID(db *sql.DB) (int, error) {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM users")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var count int
	err = stmt.QueryRow().Scan(&count)
	if err != nil {
		return 0, err
	}
	return count + 1, nil
}

func GetUserBalance(db *sql.DB, username string) (float64, error) {
	stmt, err := db.Prepare("SELECT starting_bal FROM users WHERE username = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var bal float64
	err = stmt.QueryRow(username).Scan(&bal)
	if err != nil {
		return 0, err
	}
	return bal, nil
}

func DepositBalance(db *sql.DB, username string, amount float64) (float64, error) {
	balance, err := GetUserBalance(db, username)
	if err != nil {
		return 0, fmt.Errorf("could not get balance: %v", err)
	}

	newBalance := balance + amount

	stmtUpdUser, err := db.Prepare("UPDATE users SET starting_bal = ? WHERE username = ?")
	if err != nil {
		return 0, err
	}
	defer stmtUpdUser.Close()

	if _, err = stmtUpdUser.Exec(newBalance, username); err != nil {
		return 0, fmt.Errorf("failed to update balance: %v", err)
	}

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

func WithdrawBalance(db *sql.DB, username string, amount float64) (float64, error) {
	balance, err := GetUserBalance(db, username)
	if err != nil {
		return 0, fmt.Errorf("could not get balance: %v", err)
	}

	newBalance := balance - amount
	if newBalance < 0 {
		return 0, fmt.Errorf("not enough in balance to withdraw. Current balance: $%.2f", balance)
	}

	stmtUpdUser, err := db.Prepare("UPDATE users SET starting_bal = ? WHERE username = ?")
	if err != nil {
		return 0, err
	}
	defer stmtUpdUser.Close()

	if _, err = stmtUpdUser.Exec(newBalance, username); err != nil {
		return 0, fmt.Errorf("failed to update balance: %v", err)
	}

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

func GetUserID(db *sql.DB, username string) (int, error) {
	stmt, err := db.Prepare("SELECT id FROM users WHERE username = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(username).Scan(&id)
	return id, err
}

func ListUsers(db *sql.DB) ([]models.User, error) {
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

func ShowTransactions(db *sql.DB) error {
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

func PrintNewATMBalance(db *sql.DB) {
	newBal, err := GetATMBalance(db)
	if err != nil {
		fmt.Println("Could not get balance:", err)
		return
	}
	fmt.Printf("New ATM balance: $%.2f\n", newBal)
}

func DepositATM(db *sql.DB, inc_amount float64) error {
	stmt, err := db.Prepare("UPDATE atm SET balance = balance + ? WHERE id = 1")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(inc_amount)
	return err
}

func WithdrawATM(db *sql.DB, dec_amount float64) error {
	bal, err := GetATMBalance(db)
	if err != nil {
		return fmt.Errorf("could not get ATM balance: %v", err)
	}
	if dec_amount > bal {
		return fmt.Errorf("ATM does not have enough cash. Current ATM balance: $%.2f", bal)
	}

	stmt, err := db.Prepare("UPDATE atm SET balance = balance - ? WHERE id = 1")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(dec_amount)
	return err
}
