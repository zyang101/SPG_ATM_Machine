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
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid username or password")
)

func AuthenticateUser(ctx context.Context, db *sql.DB, username, password string) (string, error) {
	// Input validation
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return "", ErrInvalidCredentials
	}

	var hash, role string
	err := db.QueryRowContext(ctx, "SELECT password_hash, role FROM credentials WHERE user_id = ?", username).Scan(&hash, &role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("database error: %w", err)
	}

	//Use Sha256
	hasher := sha256.New()
	hasher.Write([]byte(password))
	passwordHash := hex.EncodeToString(hasher.Sum(nil))

	if passwordHash != hash {
		return "", ErrInvalidCredentials
	}

	return role, nil
}

func HashPassword(password string) (string, error) {
	matched, _ := regexp.MatchString(`^\d{5}$`, password)
	if !matched {
		return "", errors.New("password must be a 5 digit pin")
	}
	hasher := sha256.New()
	hasher.Write([]byte(password))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return hash, nil
}
