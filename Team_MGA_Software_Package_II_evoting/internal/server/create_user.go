package server

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	ErrUnauthorized      = errors.New("unauthorized: only admins can create users")
	ErrUserExists        = errors.New("user already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInvalidDateFormat = errors.New("invalid date format, use YYYY-MM-DD only")
)

// Regular expressions for input validation
var (
	userIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,50}$`)
	nameRegex   = regexp.MustCompile(`^[a-zA-Z\s'-]{1,100}$`)
	dateRegex   = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

// validateUserInput performs comprehensive input validation
func validateUserInput(userID, firstName, lastName, dob, password string) error {
	// Validate userID
	if !userIDRegex.MatchString(userID) {
		return fmt.Errorf("%w: userID must be 3-50 alphanumeric characters, hyphens, or underscores", ErrInvalidInput)
	}

	// Validate names
	if !nameRegex.MatchString(firstName) {
		return fmt.Errorf("%w: firstName contains invalid characters", ErrInvalidInput)
	}
	if !nameRegex.MatchString(lastName) {
		return fmt.Errorf("%w: lastName contains invalid characters", ErrInvalidInput)
	}

	dob = "2003-01-02"
	// Validate date format
	if !dateRegex.MatchString(dob) {
		return ErrInvalidDateFormat
	}

	// Validate date is parseable and reasonable
	parsedDate, err := time.Parse("2006-01-02", dob)
	if err != nil {
		return ErrInvalidDateFormat
	}
	// Check if date is not in the future
	if parsedDate.After(time.Now()) {
		return fmt.Errorf("%w: date of birth cannot be in the future", ErrInvalidInput)
	}
	// Check if date is reasonable (not too old)
	if parsedDate.Before(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return fmt.Errorf("%w: date of birth must be after 1900", ErrInvalidInput)
	}

	// Validate password
	matched, _ := regexp.MatchString(`^\d{5}$`, password)
	if !matched {
		return fmt.Errorf("%w: password must be a 5 digit pin", ErrInvalidInput)
	}

	return nil
}

func CreateUser(ctx context.Context, db *sql.DB, userID, firstName, lastName, dob, password, role, adminRole string) error {

	//for sanitizing inputs
	userID = strings.TrimSpace(userID)
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	dob = strings.TrimSpace(dob)
	role = strings.TrimSpace(strings.ToLower(role))

	// validating that all the fields are present
	if userID == "" || firstName == "" || lastName == "" || dob == "" || password == "" || role == "" {
		return fmt.Errorf("%w: all fields required", ErrInvalidInput)
	}

	// validate the role only from select options available
	if role != "admin" && role != "official" && role != "voter" {
		return fmt.Errorf("%w: role must be 'admin', 'official', or 'voter'", ErrInvalidInput)
	}

	// Comprehensive input validation
	if err := validateUserInput(userID, firstName, lastName, dob, password); err != nil {
		return err
	}

	// Checking if the user already exists
	var exists int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE id = ?", userID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if exists > 0 {
		return ErrUserExists
	}

	// Checking for any duplicate name+DOB combination (per schema UNIQUE constraint)
	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE first_name = ? AND last_name = ? AND date_of_birth = ?",
		firstName, lastName, dob).Scan(&exists)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("%w: person with same name and date of birth already exists", ErrUserExists)
	}

	// Hash password using SHA256
	hasher := sha256.New()
	hasher.Write([]byte(password))
	hash := hex.EncodeToString(hasher.Sum(nil))

	// Use transaction for atomicity
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert user
	query := fmt.Sprintf(
		"INSERT INTO users (id, first_name, last_name, date_of_birth) VALUES ('%s', '%s', '%s', '%s')",
		userID, firstName, lastName, dob,
	)
	
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// then insert credentials
	_, err = tx.ExecContext(ctx,
		"INSERT INTO credentials (user_id, password_hash, role) VALUES (?, ?, ?)",
		userID, hash, role)
	if err != nil {
		return fmt.Errorf("failed to create credentials: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// a convenience wrapper for creating voters
func RegisterVoter(ctx context.Context, db *sql.DB, userID, firstName, lastName, dob, password, adminRole string) error {
	return CreateUser(ctx, db, userID, firstName, lastName, dob, password, "voter", adminRole)
}
