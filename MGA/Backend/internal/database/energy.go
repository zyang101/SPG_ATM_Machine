package database

import (
    "context"
    "database/sql"
    "time"
)

type EnergyUsage struct {
    ID          int64
    HomeownerID int64
    RecordedAt  time.Time
    KWh         float64
    Cost        sql.NullFloat64
}

type EnergyRepository struct { db *sql.DB }

func NewEnergyRepository(db *sql.DB) *EnergyRepository { return &EnergyRepository{db: db} }

func (r *EnergyRepository) Insert(ctx context.Context, e *EnergyUsage) (int64, error) {
    res, err := r.db.ExecContext(ctx, `INSERT INTO energy_usage (homeowner_id, kwh, cost) VALUES (?,?,?)`, e.HomeownerID, e.KWh, e.Cost)
    if err != nil { return 0, err }
    return res.LastInsertId()
}

func (r *EnergyRepository) ListRange(ctx context.Context, homeownerID int64, start, end time.Time) ([]EnergyUsage, error) {
    rows, err := r.db.QueryContext(ctx, `SELECT id, homeowner_id, recorded_at, kwh, cost FROM energy_usage WHERE homeowner_id = ? AND recorded_at BETWEEN ? AND ? ORDER BY recorded_at`, homeownerID, start, end)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []EnergyUsage
    for rows.Next() {
        var e EnergyUsage
        if err := rows.Scan(&e.ID, &e.HomeownerID, &e.RecordedAt, &e.KWh, &e.Cost); err != nil { return nil, err }
        out = append(out, e)
    }
    return out, rows.Err()
}


