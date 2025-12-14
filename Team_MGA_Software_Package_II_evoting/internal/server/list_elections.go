package server

import (
	"context"
	"database/sql"
)

type ElectionListRow struct {
	ElectionID   int64  `json:"election_id"`
	OfficialID   string `json:"official_id"`
	ElectionName string `json:"election_name"`
	DistrictName string `json:"district_name"`
	IsActive     int    `json:"is_active"` // 1 for active (open)
}

func ListOpenElections(ctx context.Context, db *sql.DB) ([]ElectionListRow, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			id,
			district_official_id,
			name,
			district,
			CASE WHEN status = 'active' THEN 1 ELSE 0 END AS is_active
		FROM elections
		WHERE status = 'active'
		ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ElectionListRow
	for rows.Next() {
		var r ElectionListRow
		if err := rows.Scan(&r.ElectionID, &r.OfficialID, &r.ElectionName, &r.DistrictName, &r.IsActive); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func ListAllElections(ctx context.Context, db *sql.DB) ([]ElectionListRow, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			id,
			district_official_id,
			name,
			district,
			CASE WHEN status = 'active' THEN 1 WHEN status = 'not_active' THEN 0 ELSE 2 END AS activation_status
		FROM elections
		ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ElectionListRow
	for rows.Next() {
		var r ElectionListRow
		if err := rows.Scan(&r.ElectionID, &r.OfficialID, &r.ElectionName, &r.DistrictName, &r.IsActive); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
