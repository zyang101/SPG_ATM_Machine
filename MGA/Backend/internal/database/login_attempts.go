package database

import (
	"context"
	"database/sql"
	"time"
)

type LoginAttempt struct {
	ID          int64
	Username    string
	Password    string
	Role        string
	Success     bool
	AttemptedAt time.Time
}

type LoginAttemptsRepository struct{ db *sql.DB }

func NewLoginAttemptsRepository(db *sql.DB) *LoginAttemptsRepository {
	return &LoginAttemptsRepository{db: db}
}

func (r *LoginAttemptsRepository) Insert(ctx context.Context, e *LoginAttempt) (int64, error) {
	res, err := r.db.ExecContext(ctx, `INSERT INTO login_attempts (username, password, role, success, attempted_at) VALUES (?,?,?,?)`, e.Username, e.Password, e.Role, e.Success, e.AttemptedAt)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
