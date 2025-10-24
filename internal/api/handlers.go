package api

import (
	"database/sql"
	"SPG_ATM_Machine/internal/models"
)

// Create a new user
func CreateUser(db *sql.DB, u models.User) error {
	_, err := db.Exec(`
        INSERT INTO users (full_name, dob, pin, starting_bal, username, role)
        VALUES (?, ?, ?, ?, ?, ?)`,
		u.FullName, u.DOB, u.PIN, u.StartingBal, u.Username, u.Role,
	)
	return err
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
