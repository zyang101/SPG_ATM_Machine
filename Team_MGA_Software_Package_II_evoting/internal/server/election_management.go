package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"spc-evoting/internal/schemes"
)

var (
	ErrElectionNotFound      = errors.New("election not found")
	ErrNotAuthorized         = errors.New("not authorized to modify this election")
	ErrInvalidStatus         = errors.New("invalid election status for this operation")
	ErrElectionAlreadyActive = errors.New("election is already active")
	ErrElectionAlreadyClosed = errors.New("election is already closed")
	ErrElectionNotClosed     = errors.New("election must be closed before tallying results")
)

// getElectionInfo retrieves election details and verifies ownership
func getElectionInfo(ctx context.Context, db *sql.DB, electionID int64, officialID string) (name, status string, err error) {
	var electionOfficialID string
	err = db.QueryRowContext(ctx, `
		SELECT name, status, district_official_id 
		FROM elections 
		WHERE id = ?`, electionID).Scan(&name, &status, &electionOfficialID)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", ErrElectionNotFound
		}
		return "", "", fmt.Errorf("failed to query election: %w", err)
	}

	if electionOfficialID != officialID {
		return "", "", ErrNotAuthorized
	}

	return name, status, nil
}

// OpenElection changes election status from 'not_started' to 'active'
// Only the district official who created the election can open it
func OpenElection(ctx context.Context, db *sql.DB, electionID int64, officialID string) error {
	_, currentStatus, err := getElectionInfo(ctx, db, electionID, officialID)
	if err != nil {
		return err
	}

	switch currentStatus {
	case "active":
		return ErrElectionAlreadyActive
	case "closed":
		return fmt.Errorf("%w: cannot reopen a closed election", ErrInvalidStatus)
	case "not_started":
		_, err := db.ExecContext(ctx, `
			UPDATE elections 
			SET status = 'active', start_date = CURRENT_TIMESTAMP 
			WHERE id = ?`, electionID)
		return err
	default:
		return fmt.Errorf("%w: unknown status '%s'", ErrInvalidStatus, currentStatus)
	}
}

// CloseElection changes election status from 'active' to 'closed'
// Only the district official who created the election can close it
func CloseElection(ctx context.Context, db *sql.DB, electionID int64, officialID string) error {
	_, currentStatus, err := getElectionInfo(ctx, db, electionID, officialID)
	if err != nil {
		return err
	}

	switch currentStatus {
	case "closed":
		return ErrElectionAlreadyClosed
	case "not_started":
		return fmt.Errorf("%w: cannot close an election that was never opened", ErrInvalidStatus)
	case "active":
		_, err := db.ExecContext(ctx, `
			UPDATE elections 
			SET status = 'closed', end_date = CURRENT_TIMESTAMP 
			WHERE id = ?`, electionID)
		return err
	default:
		return fmt.Errorf("%w: unknown status '%s'", ErrInvalidStatus, currentStatus)
	}
}

// getPositions retrieves all positions for an election
func getPositions(ctx context.Context, db *sql.DB, electionID int64) ([]struct {
	id   int64
	name string
}, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, name 
		FROM positions 
		WHERE election_id = ? 
		ORDER BY id`, electionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer rows.Close()

	var positions []struct {
		id   int64
		name string
	}
	for rows.Next() {
		var p struct {
			id   int64
			name string
		}
		if err := rows.Scan(&p.id, &p.name); err != nil {
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}
		positions = append(positions, p)
	}
	return positions, rows.Err()
}

// TODO: improve this function
// queryCandidateVotes retrieves vote counts for all candidates in a position
// This is a helper function used by both TallyResults and GetResults
func queryCandidateVotes(ctx context.Context, querier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}, positionID int64) ([]schemes.CandidateResult, error) {
	rows, err := querier.QueryContext(ctx, `
		SELECT 
			c.id,
			c.name,
			c.party_name,
			COALESCE(COUNT(v.id), 0) as vote_count
		FROM candidates c
		LEFT JOIN votes v ON v.candidate_id = c.id
		WHERE c.position_id = ?
		GROUP BY c.id, c.name, c.party_name
		ORDER BY vote_count DESC, c.id`, positionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query candidate votes: %w", err)
	}
	defer rows.Close()

	var candidates []schemes.CandidateResult
	for rows.Next() {
		var cr schemes.CandidateResult
		if err := rows.Scan(&cr.CandidateID, &cr.Name, &cr.Party, &cr.VoteCount); err != nil {
			return nil, fmt.Errorf("failed to scan candidate result: %w", err)
		}
		candidates = append(candidates, cr)
	}

	return candidates, rows.Err()
}

// TallyResults counts votes and determines winners for each position
// This also updates the winner_candidate_id in the positions table
func TallyResults(ctx context.Context, db *sql.DB, electionID int64, officialID string) (*schemes.ElectionResults, error) {
	// Verify ownership and get election info
	electionName, status, err := getElectionInfo(ctx, db, electionID, officialID)
	if err != nil {
		return nil, err
	}

	// Election must be closed to tally results
	if status != "closed" {
		return nil, ErrElectionNotClosed
	}

	results := &schemes.ElectionResults{
		ElectionName: electionName,
		IsActive:     false,
	}

	// Get all positions - read into memory to avoid connection pool issues
	positions, err := getPositions(ctx, db, electionID)
	if err != nil {
		return nil, err
	}

	// Begin transaction for updating winners
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// For each position, get vote counts and update winner
	for _, pos := range positions {
		candidates, err := queryCandidateVotes(ctx, tx, pos.id)
		if err != nil {
			return nil, err
		}

		positionResult := schemes.PositionResult{
			PositionID:   pos.id,
			PositionName: pos.name,
			Candidates:   candidates,
		}

		// Update winner if there are votes
		if len(candidates) > 0 && candidates[0].VoteCount > 0 {
			winnerID := candidates[0].CandidateID
			_, err = tx.ExecContext(ctx, `
				UPDATE positions 
				SET winner_candidate_id = ? 
				WHERE id = ?`, winnerID, pos.id)

			if err != nil {
				return nil, fmt.Errorf("failed to update position winner: %w", err)
			}
			positionResult.WinnerID = &winnerID
		}

		results.Positions = append(results.Positions, positionResult)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return results, nil
}

// GetResults retrieves the current results for an election without modifying the database
func GetResults(ctx context.Context, db *sql.DB, electionID int64, officialID string) (*schemes.ElectionResults, error) {
	// Verify ownership and get election info
	electionName, status, err := getElectionInfo(ctx, db, electionID, officialID)
	if err != nil {
		return nil, err
	}

	results := &schemes.ElectionResults{
		ElectionName: electionName,
		IsActive:     status == "active",
	}

	// Get all positions with winners - read into memory to avoid connection pool issues
	type posData struct {
		id, winnerID int64
		name         string
		hasWinner    bool
	}

	rows, err := db.QueryContext(ctx, `
		SELECT id, name, winner_candidate_id 
		FROM positions 
		WHERE election_id = ? 
		ORDER BY id`, electionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}

	var positions []posData
	for rows.Next() {
		var p posData
		var winnerID sql.NullInt64
		if err := rows.Scan(&p.id, &p.name, &winnerID); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}
		if winnerID.Valid {
			p.winnerID = winnerID.Int64
			p.hasWinner = true
		}
		positions = append(positions, p)
	}
	rows.Close()

	// Now query candidates for each position
	for _, pos := range positions {
		candidates, err := queryCandidateVotes(ctx, db, pos.id)
		if err != nil {
			return nil, err
		}

		posResult := schemes.PositionResult{
			PositionID:   pos.id,
			PositionName: pos.name,
			Candidates:   candidates,
		}
		if pos.hasWinner {
			posResult.WinnerID = &pos.winnerID
		}
		results.Positions = append(results.Positions, posResult)
	}

	return results, nil
}
