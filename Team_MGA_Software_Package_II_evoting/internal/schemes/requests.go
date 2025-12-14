package schemes

// Authentication requests
type LoginReq struct {
	UserID   string `json:"username"`
	Password string `json:"password"`
}

// User creation requests
type CreateUserReq struct {
	UserID    string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	DOB       string `json:"date_of_birth"` // Format: YYYY-MM-DD (yyyy-mm-dd)
	Password  string `json:"password"`
	Role      string `json:"role"` // "admin", "official", or "voter"
	// Note: AdminID is derived from session token, not sent in request
}

// Voting requests
type VoteReq struct {
	VoterUserID string `json:"voterUserId"`
	PositionID  int64  `json:"positionId"`
	CandidateID int64  `json:"candidateId"`
}

// Types for casted ballot submission
type CastedBallotReq struct {
	VoterUserID string          `json:"username"`
	ElectionID  int64           `json:"election_id"`
	Ballot      []BallotVoteReq `json:"ballot"`
}

type BallotVoteReq struct {
	PositionID  int64 `json:"position_id"`
	CandidateID int64 `json:"candidate_id"`
}
