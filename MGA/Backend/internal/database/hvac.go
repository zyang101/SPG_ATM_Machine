package database

import (
	"context"
	"database/sql"
	"time"
)

type HVACEvent struct {
	ID          int64
	HomeownerID int64
	OccurredAt  time.Time
	Mode        string
	State       string
	DurationSec sql.NullInt64
}

type HVACRepository struct{ db *sql.DB }

func NewHVACRepository(db *sql.DB) *HVACRepository { return &HVACRepository{db: db} }

func (r *HVACRepository) Insert(ctx context.Context, e *HVACEvent) (int64, error) {
	res, err := r.db.ExecContext(ctx, `INSERT INTO hvac_events (homeowner_id, mode, state, duration_sec) VALUES (?,?,?,?)`, e.HomeownerID, e.Mode, e.State, e.DurationSec)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *HVACRepository) ListRecent(ctx context.Context, homeownerID int64, limit int) ([]HVACEvent, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, homeowner_id, occurred_at, mode, state, duration_sec FROM hvac_events WHERE homeowner_id = ? ORDER BY occurred_at DESC LIMIT ?`, homeownerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []HVACEvent
	for rows.Next() {
		var e HVACEvent
		if err := rows.Scan(&e.ID, &e.HomeownerID, &e.OccurredAt, &e.Mode, &e.State, &e.DurationSec); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *HVACRepository) CountByHomeowner(ctx context.Context, homeownerID int64) (int64, error) {
	row := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM hvac_events WHERE homeowner_id = ?`, homeownerID)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
