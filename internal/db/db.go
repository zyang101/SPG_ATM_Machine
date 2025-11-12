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
        role TEXT,
		failed_attempts INTEGER DEFAULT 0,
    	locked INTEGER DEFAULT 0
    );`
	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	atmTable := `
    CREATE TABLE IF NOT EXISTS atm (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        balance REAL GENERATED ALWAYS AS (ones * 1 + fives * 5 + tens * 10 + twenties * 20 + fifties * 50 + hundreds * 100) STORED,
		withdrawal_limit REAL DEFAULT 0,
    	deposit_limit REAL DEFAULT 0,
		ones INTEGER DEFAULT 0,
		fives INTEGER DEFAULT 0,
		tens INTEGER DEFAULT 0,
		twenties INTEGER DEFAULT 0,
		fifties INTEGER DEFAULT 0,
		hundreds INTEGER DEFAULT 0
    );`

	_, err = db.Exec(atmTable)
	if err != nil {
		return nil, err
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM atm").Scan(&count)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		_, err = db.Exec(`
			INSERT INTO atm (withdrawal_limit, deposit_limit, ones, fives, tens, twenties, fifties, hundreds)
			VALUES (500, 1000, 0, 0, 0, 0, 0, 0);
		`)
		if err != nil {
			return nil, err
		}
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
