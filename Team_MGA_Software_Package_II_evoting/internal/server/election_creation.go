package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type CandidateInput struct {
	Name  string `json:"name"`
	Party string `json:"party"`
}

type PositionInput struct {
	Name       string           `json:"name"`
	Candidates []CandidateInput `json:"candidates"`
}

type ElectionInput struct {
	Name       string          `json:"election_name"`
	District   string          `json:"district_name"`
	OfficialID string          `json:"official_username"`
	Positions  []PositionInput `json:"positions"`
}

var (
	ErrTooFewPositions = errors.New("an election must have at least 3 positions")
	ErrTooFewCandidates = errors.New("each position must have at least 2 candidates")
)

func CreateElectionWithStructure(ctx context.Context, db *sql.DB, in ElectionInput) (int64, error) {
	// Basic validation
	if len(in.Positions) < 3 {
		return 0, ErrTooFewPositions
	}

	for _, p := range in.Positions {
		if strings.TrimSpace(p.Name) == "" {
			return 0, fmt.Errorf("position name required")
		}
		if len(p.Candidates) < 2 {
			return 0, fmt.Errorf("%w: %s", ErrTooFewCandidates, p.Name)
		}
		for _, c := range p.Candidates {
			if strings.TrimSpace(c.Name) == "" {
				return 0, fmt.Errorf("candidate name required -- position= %s", p.Name)
			}
			if strings.TrimSpace(c.Party) == "" {
				return 0, fmt.Errorf("candidate party required -- position= %s", p.Name)
			}
		}
	}

	// Tx for atomicity
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Insert election (FK requires official user exists in users.id)
	res, err := tx.ExecContext(ctx, `
		INSERT INTO elections (name, district, district_official_id, status)
		VALUES (?, ?, ?, 'not_started')`,
		in.Name, in.District, in.OfficialID)
	if err != nil {
		return 0, err
	}
	electionID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert positions, then candidates
	for _, p := range in.Positions {
		pr, err := tx.ExecContext(ctx, `
			INSERT INTO positions (name, election_id)
			VALUES (?, ?)`, p.Name, electionID)
		if err != nil {
			return 0, err
		}
		positionID, err := pr.LastInsertId()
		if err != nil {
			return 0, err
		}

		for _, c := range p.Candidates {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO candidates (name, position_id, party_name)
				VALUES (?, ?, ?)`, c.Name, positionID, c.Party); err != nil {
				return 0, err
			}
		}
	}

	// Commit
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return electionID, nil
}
