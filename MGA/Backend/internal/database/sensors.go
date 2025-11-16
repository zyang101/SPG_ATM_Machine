package database

import (
    "context"
    "database/sql"
    "time"
)

type SensorReading struct {
    ID          int64     `json:"id"`
    HomeownerID int64     `json:"homeowner_id"`
    RecordedAt  time.Time `json:"recorded_at"`
    IndoorTemp  float64   `json:"indoor_temp"`
    Humidity    float64   `json:"humidity"`
    COPPM       float64   `json:"co_ppm"`
}

type SensorsRepository struct { db *sql.DB }

func NewSensorsRepository(db *sql.DB) *SensorsRepository { return &SensorsRepository{db: db} }

func (r *SensorsRepository) Insert(ctx context.Context, s *SensorReading) (int64, error) {
    res, err := r.db.ExecContext(ctx, `INSERT INTO sensor_readings (homeowner_id, indoor_temp, humidity, co_ppm) VALUES (?,?,?,?)`, s.HomeownerID, s.IndoorTemp, s.Humidity, s.COPPM)
    if err != nil { return 0, err }
    return res.LastInsertId()
}

func (r *SensorsRepository) UpdateByID(ctx context.Context, id int64, temp, humidity, coPPM float64) error {
    _, err := r.db.ExecContext(ctx, `UPDATE sensor_readings SET indoor_temp = ?, humidity = ?, co_ppm = ?, recorded_at = CURRENT_TIMESTAMP WHERE id = ?`, temp, humidity, coPPM, id)
    return err
}

func (r *SensorsRepository) ListRecent(ctx context.Context, homeownerID int64, limit int) ([]SensorReading, error) {
    rows, err := r.db.QueryContext(ctx, `SELECT id, homeowner_id, recorded_at, indoor_temp, humidity, co_ppm FROM sensor_readings WHERE homeowner_id = ? ORDER BY recorded_at DESC LIMIT ?`, homeownerID, limit)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []SensorReading
    for rows.Next() {
        var s SensorReading
        if err := rows.Scan(&s.ID, &s.HomeownerID, &s.RecordedAt, &s.IndoorTemp, &s.Humidity, &s.COPPM); err != nil { return nil, err }
        out = append(out, s)
    }
    return out, rows.Err()
}

func (r *SensorsRepository) GetLatest(ctx context.Context, homeownerID int64) (*SensorReading, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, homeowner_id, recorded_at, indoor_temp, humidity, co_ppm FROM sensor_readings WHERE homeowner_id = ? ORDER BY recorded_at DESC LIMIT 1`, homeownerID)
	var s SensorReading
	err := row.Scan(&s.ID, &s.HomeownerID, &s.RecordedAt, &s.IndoorTemp, &s.Humidity, &s.COPPM)
	if err != nil {
		return nil, err
	}
	return &s, nil
}




