package database

import (
    "context"
    "database/sql"
    "errors"
    "time"
)

type Schedule struct {
    ID          int64     `json:"id"`
    HomeownerID int64     `json:"homeowner_id"`
    Name        string    `json:"name"`
    StartTime   string    `json:"start_time"`
    TargetTemp  float64   `json:"target_temp"`
    CreatedAt   time.Time `json:"created_at"`
}

type SchedulesRepository struct { db *sql.DB }

func NewSchedulesRepository(db *sql.DB) *SchedulesRepository { return &SchedulesRepository{db: db} }

func (r *SchedulesRepository) Create(ctx context.Context, s *Schedule) (int64, error) {
    if r.db == nil { return 0, errors.New("repo not initialized") }
    res, err := r.db.ExecContext(ctx,
        `INSERT INTO schedules (homeowner_id, name, start_time, target_temp) VALUES (?,?,?,?)`,
        s.HomeownerID, s.Name, s.StartTime, s.TargetTemp,
    )
    if err != nil { return 0, err }
    return res.LastInsertId()
}

func (r *SchedulesRepository) ListAll(ctx context.Context) ([]Schedule, error) {
    rows, err := r.db.QueryContext(ctx, `SELECT id, homeowner_id, name, start_time, target_temp, created_at FROM schedules ORDER BY start_time`)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []Schedule
    for rows.Next() {
        var s Schedule
        if err := rows.Scan(&s.ID, &s.HomeownerID, &s.Name, &s.StartTime, &s.TargetTemp, &s.CreatedAt); err != nil { return nil, err }
        out = append(out, s)
    }
    return out, rows.Err()
}

func (r *SchedulesRepository) Delete(ctx context.Context, id int64) error {
    _, err := r.db.ExecContext(ctx, `DELETE FROM schedules WHERE id = ?`, id)
    return err
}


