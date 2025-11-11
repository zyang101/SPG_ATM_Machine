package api

import (
	"database/sql"
	"fmt"
	"strings"
)

type UserAuthInfo struct {
	PINHash        string
	FailedAttempts int
	Locked         bool
	Role           string
}

func FetchUserRole(conn *sql.DB, username string) (string, error) {
	var role string
	stmt, err := conn.Prepare("SELECT role FROM users WHERE username = ?")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&role)
	return strings.ToLower(role), err
}

func GetUserAuth(db *sql.DB, username string) (*UserAuthInfo, error) {
	stmt, err := db.Prepare(`
		SELECT pin, failed_attempts, locked, role
		FROM users
		WHERE username = ?
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	var info UserAuthInfo
	var lockedInt int

	err = stmt.QueryRow(username).Scan(&info.PINHash, &info.FailedAttempts, &lockedInt, &info.Role)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	} else if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}

	info.Locked = lockedInt == 1

	return &info, nil
}

// Increments the failed attempts for a user.
func IncrementFailedAttempts(db *sql.DB, username string) (int, bool, error) {
    stmt, err := db.Prepare("SELECT failed_attempts FROM users WHERE username = ?")
    if err != nil {
        return 0, false, err
    }
    defer stmt.Close()

    var attempts int
    if err := stmt.QueryRow(username).Scan(&attempts); err != nil {
        return 0, false, err
    }

    attempts++
    locked := false
    if attempts >= 3 {
        locked = true
        _, err = db.Exec("UPDATE users SET failed_attempts = ?, locked = 1 WHERE username = ?", attempts, username)
    } else {
        _, err = db.Exec("UPDATE users SET failed_attempts = ? WHERE username = ?", attempts, username)
    }
    return attempts, locked, err
}

// Resets failed attempts to 0 for the user.
func ResetFailedAttempts(db *sql.DB, username string) error {
    _, err := db.Exec("UPDATE users SET failed_attempts = 0 WHERE username = ?", username)
    return err
}


// Resets failed attempts and unlocks a user's account
func UnlockAccount(db *sql.DB, username string) error {
	// Check if user exists
	var exists int
	err := db.QueryRow("SELECT COUNT(1) FROM users WHERE username = ?", username).Scan(&exists)
	if err != nil {
		return fmt.Errorf("database error checking user existence: %v", err)
	}
	if exists == 0 {
		return fmt.Errorf("no user found with username '%s'", username)
	}

	// Reset failed attempts and unlock
	_, err = db.Exec("UPDATE users SET failed_attempts = 0, locked = 0 WHERE username = ?", username)
	if err != nil {
		return fmt.Errorf("failed to unlock account for '%s': %v", username, err)
	}

	return nil
}
