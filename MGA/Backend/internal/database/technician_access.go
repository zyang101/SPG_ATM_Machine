package database

import (
    "context"
    "database/sql"
    "time"
)

type TechnicianAccess struct {
    ID           int64
    HomeownerID  int64
    TechnicianID int64
    StartTime    time.Time
    EndTime      time.Time
}

type TechnicianAccessRepository struct { db *sql.DB }

func NewTechnicianAccessRepository(db *sql.DB) *TechnicianAccessRepository { return &TechnicianAccessRepository{db: db} }

func (r *TechnicianAccessRepository) Grant(ctx context.Context, t *TechnicianAccess) (int64, error) {
    res, err := r.db.ExecContext(ctx, `INSERT INTO technician_access (homeowner_id, technician_id, start_time, end_time) VALUES (?,?,?,?)`, t.HomeownerID, t.TechnicianID, t.StartTime, t.EndTime)
    if err != nil { return 0, err }
    return res.LastInsertId()
}

func (r *TechnicianAccessRepository) IsAllowedNow(ctx context.Context, homeownerID, technicianID int64, now time.Time) (bool, error) {
    // Ensure now is in UTC for consistent comparison with database times (stored in UTC)
    nowUTC := now.UTC()
    
    // Clean up expired records before checking (non-blocking - ignore errors)
    _ = r.DeleteExpired(ctx, nowUTC)
    
    row := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM technician_access WHERE homeowner_id = ? AND technician_id = ? AND start_time <= ? AND end_time >= ?`, homeownerID, technicianID, nowUTC, nowUTC)
    var count int
    if err := row.Scan(&count); err != nil { return false, err }
    return count > 0, nil
}

func (r *TechnicianAccessRepository) ListForHomeowner(ctx context.Context, homeownerID int64) ([]TechnicianAccess, error) {
    rows, err := r.db.QueryContext(ctx, `SELECT id, homeowner_id, technician_id, start_time, end_time FROM technician_access WHERE homeowner_id = ? ORDER BY start_time DESC`, homeownerID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var access []TechnicianAccess
    for rows.Next() {
        var a TechnicianAccess
        if err := rows.Scan(&a.ID, &a.HomeownerID, &a.TechnicianID, &a.StartTime, &a.EndTime); err != nil {
            return nil, err
        }
        access = append(access, a)
    }
    return access, rows.Err()
}

func (r *TechnicianAccessRepository) GetByID(ctx context.Context, id int64) (*TechnicianAccess, error) {
    row := r.db.QueryRowContext(ctx, `SELECT id, homeowner_id, technician_id, start_time, end_time FROM technician_access WHERE id = ?`, id)
    var a TechnicianAccess
    if err := row.Scan(&a.ID, &a.HomeownerID, &a.TechnicianID, &a.StartTime, &a.EndTime); err != nil {
        return nil, err
    }
    return &a, nil
}

func (r *TechnicianAccessRepository) Revoke(ctx context.Context, id int64) error {
    _, err := r.db.ExecContext(ctx, `DELETE FROM technician_access WHERE id = ?`, id)
    return err
}

// DeleteExpired removes all technician access records that have expired (end_time < now)
func (r *TechnicianAccessRepository) DeleteExpired(ctx context.Context, now time.Time) error {
    nowUTC := now.UTC()
    _, err := r.db.ExecContext(ctx, `DELETE FROM technician_access WHERE end_time < ?`, nowUTC)
    return err
}


