package api

import (
	"SPG_ATM_Machine/internal/models"
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"

)

// Create a new user
func CreateUser(db *sql.DB, fullName, dob, pin string, startingBal float64, username, role string) error {
	var exists bool
    err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username).Scan(&exists)
    if err != nil {
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

	_, err = db.Exec(`
        INSERT INTO users (id, full_name, dob, pin, starting_bal, username, role)
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
		nextID, fullName, dob, string(hashedPin), startingBal, username, role)
	return err
}

func GetNextUserID(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count + 1, nil
}

func GetUserBalance(db *sql.DB, username string) (float64, error) {
	var bal float64
	err := db.QueryRow("SELECT starting_bal FROM users WHERE username = ?", username).Scan(&bal)
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

	_, err = db.Exec("UPDATE users SET starting_bal = ? WHERE username = ?", newBalance, username)
	if err != nil {
		return 0, fmt.Errorf("failed to update balance: %v", err)
	}

	_, err = db.Exec("UPDATE atm SET balance = balance + ? WHERE id = 1", amount)
	if err != nil {
		return 0, fmt.Errorf("failed to update ATM balance: %v", err)
	}

	userID, err := GetUserID(db, username)
	if err != nil {
		return newBalance, fmt.Errorf("could not get user id: %v", err)
	}

	_, err = db.Exec(`
        INSERT INTO transactions (user_id, date, balance)
        VALUES (?, datetime('now', 'localtime'), ?)`,
		userID, amount,
	)
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
		return 0, fmt.Errorf("not enough in balance to withdraw.  Current balance: $%.2f", balance)
	}

	_, err = db.Exec("UPDATE users SET starting_bal = ? WHERE username = ?", newBalance, username)
	if err != nil {
		return 0, fmt.Errorf("failed to update balance: %v", err)
	}

	_, err = db.Exec("UPDATE atm SET balance = balance - ? WHERE id = 1", amount)
	if err != nil {
		return 0, fmt.Errorf("failed to update ATM balance: %v", err)
	}

	userID, err := GetUserID(db, username)
	if err != nil {
		return newBalance, fmt.Errorf("could not get user id: %v", err)
	}

	_, err = db.Exec(`
        INSERT INTO transactions (user_id, date, balance)
        VALUES (?, datetime('now', 'localtime'), ?)`,
		userID, -amount,
	)
	if err != nil {
		return newBalance, fmt.Errorf("failed to log transaction: %v", err)
	}
	return newBalance, err
}
func GetUserID(db *sql.DB, username string) (int, error) {
	var id int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// List all users
func ListUsers(db *sql.DB) ([]models.User, error) {
	rows, err := db.Query(`SELECT id, full_name, dob, pin, starting_bal, username, role FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.FullName, &u.DOB, &u.PIN, &u.StartingBal, &u.Username, &u.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
