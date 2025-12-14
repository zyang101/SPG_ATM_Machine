package schemes

//201 Created
//409 Foreign Key Constraint
//400 missing candidates or positions
//400 Bad json

//error type is included for each

// Authentication response
type LoginRes struct {
	Token   string `json:"token"`
	Expires string `json:"expires"` // ISO 8601 format
	Role    string `json:"role"`
}

// User creation response
type CreateUserRes struct {
	UserID  string `json:"userId"`
	Message string `json:"message"`
}

// Generic ID response
type IDResponse struct {
	ID int64 `json:"id"` // ID of the created election or vote
}

// ErrorResponse is a simple JSON error format.
type ErrorResponse struct {
	Error string `json:"error"`
}

// PositionResult represents the vote tally for a position
type PositionResult struct {
	PositionID   int64             `json:"position_id"`
	PositionName string            `json:"position_name"`
	Candidates   []CandidateResult `json:"candidates"`
	WinnerID     *int64            `json:"winner_id,omitempty"`
}

// CandidateResult represents a candidate's vote count
type CandidateResult struct {
	CandidateID int64  `json:"candidate_id"`
	Name        string `json:"name"`
	Party       string `json:"party"`
	VoteCount   int64  `json:"vote_count"`
}

// What the server sends from GET /elections/{id}/results
type ElectionResults struct {
	ElectionName string           `json:"election_name"`
	IsActive     bool             `json:"is_active"`
	Positions    []PositionResult `json:"positions"`
}

type ElectionRes struct {
	ElectionID   int64  `json:"election_id"`
	OfficialID   string `json:"official_id"`
	ElectionName string `json:"election_name"`
	DistrictName string `json:"district_name"`
	IsActive     int64  `json:"is_active"`
}

// What the server sends from GET /elections
type ElectionsListRes struct {
	Elections []ElectionRes `json:"elections"`
}

type CandidateBallotRes struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Party string `json:"party"`
}

type PositionBallotRes struct {
	PositionID   int64                `json:"position_id"`
	PositionName string               `json:"position_name"`
	Candidates   []CandidateBallotRes `json:"candidates"`
}

// What the server sends from GET /elections/{id}
type ElectionBallotRes struct {
	ElectionID int64               `json:"election_id"`
	Positions  []PositionBallotRes `json:"positions"`
}
