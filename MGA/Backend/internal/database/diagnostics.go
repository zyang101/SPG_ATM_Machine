package database

import (
    "context"
    "database/sql"
    "time"
)

type DiagnosticLog struct {
    ID          int64
    HomeownerID sql.NullInt64
    LoggedAt    time.Time
    Level       string
    Message     string
}

type DiagnosticsRepository struct { db *sql.DB }

func NewDiagnosticsRepository(db *sql.DB) *DiagnosticsRepository { return &DiagnosticsRepository{db: db} }

func (r *DiagnosticsRepository) Insert(ctx context.Context, d *DiagnosticLog) (int64, error) {
    res, err := r.db.ExecContext(ctx, `INSERT INTO diagnostics_logs (homeowner_id, level, message) VALUES (?,?,?)`, d.HomeownerID, d.Level, d.Message)
    if err != nil { return 0, err }
    return res.LastInsertId()
}

func (r *DiagnosticsRepository) ListRecent(ctx context.Context, homeownerID sql.NullInt64, limit int) ([]DiagnosticLog, error) {
    var rows *sql.Rows
    var err error
    if homeownerID.Valid {
        rows, err = r.db.QueryContext(ctx, `SELECT id, homeowner_id, logged_at, level, message FROM diagnostics_logs WHERE homeowner_id = ? ORDER BY logged_at DESC LIMIT ?`, homeownerID.Int64, limit)
    } else {
        rows, err = r.db.QueryContext(ctx, `SELECT id, homeowner_id, logged_at, level, message FROM diagnostics_logs ORDER BY logged_at DESC LIMIT ?`, limit)
    }
    if err != nil { return nil, err }
    defer rows.Close()
    var out []DiagnosticLog
    for rows.Next() {
        var d DiagnosticLog
        if err := rows.Scan(&d.ID, &d.HomeownerID, &d.LoggedAt, &d.Level, &d.Message); err != nil { return nil, err }
        out = append(out, d)
    }
    return out, rows.Err()
}


