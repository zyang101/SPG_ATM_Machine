package client

import (
	"spc-evoting/internal/models"
	"spc-evoting/internal/schemes"
)

var sample_elections = []models.Election{
	{ElectionID: 1, DistrictOfficialID: "101", Name: "Election 1", District: "New York", IsActive: models.PRE_ELECTION},
	{ElectionID: 2, DistrictOfficialID: "102", Name: "Election 2", District: "New Jersey", IsActive: models.PRE_ELECTION},
	{ElectionID: 3, DistrictOfficialID: "103", Name: "Election 3", District: "Maryland", IsActive: models.ACTIVE_ELECTION},
	{ElectionID: 4, DistrictOfficialID: "103", Name: "Election 4", District: "Virginia", IsActive: models.POST_ELECTION},
}

var sample_positions = []models.Position{
	{PositionID: 201, PositionName: "Governor", ElectionID: 2, WinnerID: -1},
	{PositionID: 202, PositionName: "Lieutenant", ElectionID: 2, WinnerID: -1},
	{PositionID: 203, PositionName: "Senate", ElectionID: 2, WinnerID: -1},
}

var sample_candidates = []models.Candidate{
	{CandidateID: 301, UserID: "id41", PositionID: 201, Name: "Jonathan Baker", AffiliatedParty: "Democratic"},
	{CandidateID: 302, UserID: "id42", PositionID: 201, Name: "Jessica Addams", AffiliatedParty: "Republican"},
	{CandidateID: 303, UserID: "id43", PositionID: 202, Name: "Benjamin Jones", AffiliatedParty: "Democratic"},
	{CandidateID: 304, UserID: "id44", PositionID: 202, Name: "Beatrice Smith", AffiliatedParty: "Republican"},
	{CandidateID: 305, UserID: "id45", PositionID: 203, Name: "Carl Griffins", AffiliatedParty: "Democratic"},
	{CandidateID: 306, UserID: "id46", PositionID: 203, Name: "Cathy Skinner", AffiliatedParty: "Republican"},
}

var sample_users = []models.User{
	{UserID: "id41", FullName: "Gov Dem Cand"},
	{UserID: "id42", FullName: "Gov Rep Cand"},
	{UserID: "id43", FullName: "Lieu Dem Cand"},
	{UserID: "id44", FullName: "Lieu Rep Cand"},
	{UserID: "id45", FullName: "Sen Dem Cand"},
	{UserID: "id46", FullName: "Sen Rep Cand"},
}

var sample_ballot = models.Ballot{
	ElectionID: 1,
	UserID:     "id41",
	Selections: []models.BallotSelection{
		{
			Position: models.Position{PositionID: 201, PositionName: "Governor", ElectionID: 2},
			Candidates: []models.Candidate{
				{CandidateID: 301, UserID: "id41", PositionID: 201, Name: "Jonathan Baker", AffiliatedParty: "Democratic"},
				{CandidateID: 302, UserID: "id42", PositionID: 201, Name: "Jessica Addams", AffiliatedParty: "Republican"},
			},
			ChosenCandidateIdx: -1,
		},
		{
			Position: models.Position{PositionID: 202, PositionName: "Lieutenant", ElectionID: 2, WinnerID: -1},
			Candidates: []models.Candidate{
				{CandidateID: 303, UserID: "id43", PositionID: 202, Name: "Benjamin Jones", AffiliatedParty: "Democratic"},
				{CandidateID: 304, UserID: "id44", PositionID: 202, Name: "Beatrice Smith", AffiliatedParty: "Republican"},
			},
			ChosenCandidateIdx: -1,
		},
		{
			Position: models.Position{PositionID: 203, PositionName: "Senate", ElectionID: 2, WinnerID: -1},
			Candidates: []models.Candidate{
				{CandidateID: 305, UserID: "id45", PositionID: 203, Name: "Carl Griffins", AffiliatedParty: "Democratic"},
				{CandidateID: 306, UserID: "id46", PositionID: 203, Name: "Cathy Skinner", AffiliatedParty: "Republican"},
			},
			ChosenCandidateIdx: -1,
		},
	},
}

var sample_cand1 = schemes.CandidateResult{
	CandidateID: 211,
	Name:        "Ababab",
	Party:       "AAAAA",
	VoteCount:   34,
}

var sample_cand2 = schemes.CandidateResult{
	CandidateID: 213,
	Name:        "Cdcdcdcdc",
	Party:       "BBBBB",
	VoteCount:   56,
}
var sample_elec_res = schemes.ElectionResults{
	ElectionName: "Prezz",
	IsActive:     false,
	Positions: []schemes.PositionResult{
		{
			PositionID:   32332,
			PositionName: "Presdient",
			Candidates:   []schemes.CandidateResult{sample_cand1, sample_cand2},
			WinnerID:     &sample_cand2.CandidateID,
		},
	},
}
