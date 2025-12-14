package server

import (
	"context"
	"database/sql"
)


type CandidateForBallot struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Party string `json:"party"`
}

type PositionForBallot struct {
	PositionID   int64                  `json:"position_id"`
	PositionName string                 `json:"position_name"`
	Candidates   []CandidateForBallot   `json:"candidates"`
}

type ElectionForBallot struct {
	ElectionID   int64                 `json:"election_id"`
	ElectionName string                `json:"election_name"`
	Positions    []PositionForBallot   `json:"positions"`
}

func GetElectionForBallot(ctx context.Context, db *sql.DB, electionID int64) (*ElectionForBallot, error) {
	// fetch basic election info
	var out ElectionForBallot
	if err := db.QueryRowContext(ctx,
		`SELECT id, name FROM elections WHERE id = ?`, electionID).
		Scan(&out.ElectionID, &out.ElectionName); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrElectionNotFound
		}
		return nil, err
	}

	// fetch positions + candidates
	rows, err := db.QueryContext(ctx, `
		SELECT
			p.id   AS position_id,
			p.name AS position_name,
			c.id   AS candidate_id,
			COALESCE(c.name, '')        AS candidate_name,
			COALESCE(c.party_name, '')  AS candidate_party
		FROM positions p
		LEFT JOIN candidates c ON c.position_id = p.id
		WHERE p.election_id = ?
		ORDER BY p.id, c.id`, electionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []PositionForBallot
	var cur PositionForBallot
	var haveCur bool

	for rows.Next() {
		var posID int64
		var posName string
		var candID sql.NullInt64
		var candName, candParty sql.NullString

		if err := rows.Scan(&posID, &posName, &candID, &candName, &candParty); err != nil {
			return nil, err
		}

		// new position boundary
		if !haveCur || cur.PositionID != posID {
			// push previous
			if haveCur {
				positions = append(positions, cur)
			}
			cur = PositionForBallot{
				PositionID:   posID,
				PositionName: posName,
				Candidates:   []CandidateForBallot{},
			}
			haveCur = true
		}

		
		if candID.Valid {
			cur.Candidates = append(cur.Candidates, CandidateForBallot{
				ID:    candID.Int64,
				Name:  candName.String,
				Party: candParty.String,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if haveCur {
		positions = append(positions, cur)
	}

	out.Positions = positions
	return &out, nil
}
