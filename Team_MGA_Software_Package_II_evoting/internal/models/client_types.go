package models

import "time"

type ActiveEnum int

const (
	PRE_ELECTION ActiveEnum = iota
	ACTIVE_ELECTION
	POST_ELECTION
)

type User struct {
	UserID      string
	FullName    string
	DateOfBirth time.Time
	CreateTime  time.Time
	UpdateTime  time.Time
}

type Candidate struct {
	CandidateID     int64
	UserID          string
	PositionID      int64
	Name            string
	AffiliatedParty string
}

type Election struct {
	ElectionID         int64
	DistrictOfficialID string
	Name               string
	District           string
	IsActive           ActiveEnum
}

type Position struct {
	PositionID   int64
	PositionName string
	ElectionID   int64
	WinnerID     int64
}

type BallotSelection struct {
	Position           Position
	Candidates         []Candidate
	ChosenCandidateIdx int // Might be bad design, must be instantiated as -1
}

type Ballot struct {
	ElectionID int64
	UserID     string
	Selections []BallotSelection
}
