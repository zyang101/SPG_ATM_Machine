package database

import (
    "context"
    "database/sql"
    "time"
)

type OutdoorWeather struct {
    ID          int64
    HomeownerID int64
    RecordedAt  time.Time
    Temp        sql.NullFloat64
    Humidity    sql.NullFloat64
    PrecipMM    sql.NullFloat64
}

type WeatherRepository struct { db *sql.DB }

func NewWeatherRepository(db *sql.DB) *WeatherRepository { return &WeatherRepository{db: db} }

func (r *WeatherRepository) Insert(ctx context.Context, w *OutdoorWeather) (int64, error) {
    res, err := r.db.ExecContext(ctx, `INSERT INTO outdoor_weather (homeowner_id, temp, humidity, precipitation_mm) VALUES (?,?,?,?)`, w.HomeownerID, w.Temp, w.Humidity, w.PrecipMM)
    if err != nil { return 0, err }
    return res.LastInsertId()
}

func (r *WeatherRepository) ListRecent(ctx context.Context, homeownerID int64, limit int) ([]OutdoorWeather, error) {
    rows, err := r.db.QueryContext(ctx, `SELECT id, homeowner_id, recorded_at, temp, humidity, precipitation_mm FROM outdoor_weather WHERE homeowner_id = ? ORDER BY recorded_at DESC LIMIT ?`, homeownerID, limit)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []OutdoorWeather
    for rows.Next() {
        var w OutdoorWeather
        if err := rows.Scan(&w.ID, &w.HomeownerID, &w.RecordedAt, &w.Temp, &w.Humidity, &w.PrecipMM); err != nil { return nil, err }
        out = append(out, w)
    }
    return out, rows.Err()
}


