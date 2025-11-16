package database

import (
	"database/sql"
	"time"
)

type GuestVerification struct {
	ID              int
	Token           string
	GuestUserID     int
	HomeownerUserID int
	TargetURL       string
	Status          string
	CreatedAt       time.Time
	ExpiresAt       sql.NullTime
}

func CreateGuestVerification(db *sql.DB, token string, guestID, homeownerID int, target string, expires time.Time) error {
	_, err := db.Exec(`INSERT INTO guest_verifications(token, guest_user_id, homeowner_user_id, target_url, status, expires_at) VALUES (?, ?, ?, ?, 'pending', ?)`,
		token, guestID, homeownerID, target, expires)
	return err
}

func GetGuestVerificationByToken(db *sql.DB, token string) (*GuestVerification, error) {
	v := &GuestVerification{}
	row := db.QueryRow(`SELECT id, token, guest_user_id, homeowner_user_id, target_url, status, created_at, expires_at FROM guest_verifications WHERE token = ?`, token)
	var expires sql.NullTime
	if err := row.Scan(&v.ID, &v.Token, &v.GuestUserID, &v.HomeownerUserID, &v.TargetURL, &v.Status, &v.CreatedAt, &expires); err != nil {
		return nil, err
	}
	v.ExpiresAt = expires
	return v, nil
}

func SetGuestVerificationStatus(db *sql.DB, token, status string) error {
	_, err := db.Exec(`UPDATE guest_verifications SET status = ? WHERE token = ?`, status, token)
	return err
}
