package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)


type BallotSelection struct {
	PositionID  int64
	CandidateID int64
}

type BallotInput struct {
	VoterUserID string
	ElectionID int64
	Ballot []BallotSelection
}

var (
	ErrBallotEmpty = errors.New("ballot must include at least one selection")
	ErrDuplicatePositionPick = errors.New("duplicate position in ballot")
	ErrPositionNotInElection = errors.New("position does not belong to election")
	ErrElectionNotActive = errors.New("election is not open")
	ErrCandidateNotInPosition = errors.New("candidate does not belong to position")
	ErrDuplicateVote = errors.New("voter already voted for this position")
)

func CastBallot(ctx context.Context, db *sql.DB, in BallotInput) (int64, error) {
	if in.ElectionID == 0 || in.VoterUserID == "" {
		return 0, errors.New("voterUserID and election_id are required")
	}
	if len(in.Ballot) == 0 {
		return 0, ErrBallotEmpty
	}

	// Election must be active
	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM elections WHERE id = ?`, in.ElectionID).Scan(&status); err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("election not found")
		}
		return 0, err
	}
	if status != "active" {
		return 0, ErrElectionNotActive
	}

	// Build a set of positions that belong to this election
	rows, err := db.QueryContext(ctx, `SELECT id FROM positions WHERE election_id = ?`, in.ElectionID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	posInElection := make(map[int64]struct{})
	for rows.Next() {
		var pid int64
		if err := rows.Scan(&pid); err != nil {
			return 0, err
		}
		posInElection[pid] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	// Validate ballot selections before inserting
	seenPos := make(map[int64]struct{})
	for _, sel := range in.Ballot {
		if _, ok := posInElection[sel.PositionID]; !ok {
			return 0, ErrPositionNotInElection
		}
		if _, dup := seenPos[sel.PositionID]; dup {
			return 0, ErrDuplicatePositionPick
		}
		seenPos[sel.PositionID] = struct{}{}

		// candidate must belong to that position
		var ok int
		if err := db.QueryRowContext(ctx, `
			SELECT 1 FROM candidates WHERE id = ? AND position_id = ?`,
			sel.CandidateID, sel.PositionID).Scan(&ok); err != nil {
			if err == sql.ErrNoRows {
				return 0, ErrCandidateNotInPosition
			}
			return 0, err
		}
	}

	// Insert all votes atomically
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var inserted int64
	for _, sel := range in.Ballot {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO votes (voter_user_id, position_id, candidate_id)
			VALUES (?, ?, ?)`, in.VoterUserID, sel.PositionID, sel.CandidateID); err != nil {
			msg := err.Error()
			if strings.Contains(msg, "UNIQUE constraint failed: votes.voter_user_id, votes.position_id") {
				err = ErrDuplicateVote
				return 0, err
			}
			if strings.Contains(msg, "FOREIGN KEY constraint failed") {
				err = fmt.Errorf("foreign key constraint failed")
				return 0, err
			}
			return 0, err
		}
		inserted++
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return inserted, nil
}