package database

import (
    "context"
    "database/sql"
)

type SystemStateRepository struct { db *sql.DB }

func NewSystemStateRepository(db *sql.DB) *SystemStateRepository { return &SystemStateRepository{db: db} }

func (r *SystemStateRepository) Get(ctx context.Context, key string) (string, bool, error) {
    row := r.db.QueryRowContext(ctx, `SELECT value FROM system_state WHERE key = ?`, key)
    var v string
    if err := row.Scan(&v); err != nil {
        if err == sql.ErrNoRows { return "", false, nil }
        return "", false, err
    }
    return v, true, nil
}

func (r *SystemStateRepository) Set(ctx context.Context, key, value string) error {
    _, err := r.db.ExecContext(ctx, `INSERT INTO system_state (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`, key, value)
    return err
}


