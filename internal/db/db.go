package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func Connect() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./data.db")
	if err != nil {
		return nil, err
	}

	query := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        full_name TEXT NOT NULL,
        dob TEXT,
        pin TEXT,
        starting_bal REAL,
        username TEXT UNIQUE,
        role TEXT
    );`
	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	atmTable := `
    CREATE TABLE IF NOT EXISTS atm (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        balance REAL
    );`
	_, err = db.Exec(atmTable)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`INSERT INTO atm (balance)
                      SELECT 0 WHERE NOT EXISTS (SELECT 1 FROM atm);`)
	if err != nil {
		return nil, err
	}

	transactions := `
	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id int,
		date TEXT NOT NULL,
		balance REAL
	);`

	_, err = db.Exec(transactions)
	if err != nil {
		return nil, err
	}

	return db, nil
}
