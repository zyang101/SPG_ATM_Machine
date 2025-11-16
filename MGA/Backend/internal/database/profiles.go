package database

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"fmt"
)

type Profile struct {
    ID          int64     `json:"id"`
    HomeownerID int64     `json:"homeowner_id"`
    Name        string    `json:"name"`
    TargetTemp  float64   `json:"target_temp"`
    CreatedAt   time.Time `json:"created_at"`
}

type ProfilesRepository struct {
	db *sql.DB
}

func NewProfilesRepository(db *sql.DB) *ProfilesRepository { return &ProfilesRepository{db: db} }

func (r *ProfilesRepository) Create(ctx context.Context, p *Profile) (int64, error) {
	if r.db == nil {
		return 0, errors.New("repo not initialized")
	}
	query := fmt.Sprintf(
		"INSERT INTO profiles (homeowner_id, name, target_temp) SELECT %d, '%s', %f",
		p.HomeownerID,
		p.Name,
		p.TargetTemp,
	)
	res, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *ProfilesRepository) ListByHomeowner(ctx context.Context, homeownerID int64) ([]Profile, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, homeowner_id, name, target_temp, created_at FROM profiles WHERE homeowner_id = ? ORDER BY name`, homeownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Profile
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.ID, &p.HomeownerID, &p.Name, &p.TargetTemp, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *ProfilesRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM profiles WHERE id = ?`, id)
	return err
}
