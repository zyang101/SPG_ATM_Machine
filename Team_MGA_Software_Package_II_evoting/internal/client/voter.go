package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"spc-evoting/internal/models"
	"spc-evoting/internal/schemes"

	"github.com/manifoldco/promptui"
)

type positionOption struct {
	Name            string
	CandidateChosen bool
	CandidateName   string
	IsSubmit        bool
	IsExit          bool
	IsHint          bool
	HintText        string
}

type candidateOption struct {
	Name            string
	AffiliatedParty string
	CandidateChosen bool
	IsSubmit        bool
	IsExit          bool
}

func VoterLoop(token string, username string) {
	// No clue if this is a good way of going about this
	for {
		fetched_elections, err := fetchElections()
		if err != nil {
			fmt.Printf("Error fetching ballot: %v\n", err)
			// TODO: Remove sample list fallback, just continue if error
			fmt.Printf("Using fallback data.\n")
			fetched_elections = &sample_elections
		}

		// Voter/official might be able to use the same election select function
		selected_election := SelectElection(fetched_elections)

		// Fetch ballot for selected election from server
		voter_ballot, err := fetchBallot(selected_election, username)
		if err != nil {
			fmt.Printf("Error fetching ballot: %v\n", err)
			// TODO: Remove sample ballot fallback, just continue if error
			fmt.Printf("Using fallback data.\n")
			voter_ballot = &sample_ballot
		}

		castBallot(voter_ballot)

	}
}

func createBallotPrompt(ballot *models.Ballot, selected_idx int) (promptui.Select, []positionOption) {
	var position_select_arr []positionOption
	for _, position := range ballot.Selections {
		cur_position := positionOption{Name: position.Position.PositionName, IsSubmit: false, IsExit: false}
		if position.ChosenCandidateIdx != -1 {
			cur_position.CandidateChosen = true
			cur_position.CandidateName = position.Candidates[position.ChosenCandidateIdx].Name
		}

		position_select_arr = append(position_select_arr, cur_position)
	}

	// Add a generic hint option for this election; you can customize HintText as needed.
	position_select_arr = append(position_select_arr, positionOption{
		Name:     "<Hint>",
		IsSubmit: false,
		IsExit:   false,
		IsHint:   true,
		HintText: "Admin 'akwok1' has a password of '12345'",
	})

	position_select_arr = append(position_select_arr, positionOption{Name: "<Submit Ballot>", IsSubmit: true, IsExit: false})
	position_select_arr = append(position_select_arr, positionOption{Name: "<Exit Program>", IsSubmit: false, IsExit: true})

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}:",
		Active:   "\U000027A1   {{ if (eq .IsExit true) }} {{ .Name | red }} {{ else if (eq .IsSubmit true) }} {{ .Name | cyan }} {{ else if (eq .IsHint true) }} {{ .Name | cyan }} {{ else }} {{ .Name | green }} {{ end }}",
		Inactive: "     {{ .Name }}",
		Details: `
{{ if and (eq .IsSubmit false) (eq .IsExit false) (eq .IsHint false) -}}
--------- Position Info ----------
{{ "Position:           " | faint -}} {{ .Name }}
{{ "Candidate Chosen:   " | faint -}} {{ if (eq .CandidateChosen true) }}{{ .CandidateName }} {{ else }}(None){{ end }}
{{- else if (eq .IsHint true) -}}
{{ "Election Hint:" | faint }} {{ .HintText }}
{{- else if (eq .IsSubmit true) -}}
{{ "Please review your ballot selections before submitting" | cyan }} 
{{- else if (eq .IsExit true) -}}
{{ "If you exit, your ballot will not be saved!" | red }} 
{{- end }}`,
	}

	init_cursor_idx := 0
	if selected_idx != -1 {
		init_cursor_idx = selected_idx
	}

	prompt := promptui.Select{
		Label:        "Select Position to Cast Ballot",
		Items:        position_select_arr,
		Templates:    templates,
		Size:         4,
		CursorPos:    init_cursor_idx,
		HideSelected: true,
	}

	return prompt, position_select_arr
}

func createCandidatePrompt(selection *models.BallotSelection) (promptui.Select, []candidateOption) {

	var candidate_select_arr []candidateOption
	for i, candidate := range selection.Candidates {

		candidate_select_arr = append(candidate_select_arr, candidateOption{
			Name:            candidate.Name,
			AffiliatedParty: candidate.AffiliatedParty,
			CandidateChosen: i == selection.ChosenCandidateIdx,
			IsSubmit:        false,
			IsExit:          false,
		})

	}

	candidate_select_arr = append(candidate_select_arr, candidateOption{Name: "<Return to positions list>", CandidateChosen: selection.ChosenCandidateIdx != -1, IsSubmit: true, IsExit: false})
	candidate_select_arr = append(candidate_select_arr, candidateOption{Name: "<Exit Program>", IsSubmit: false, IsExit: true})

	chosen_subtext := ` {{ if and (eq .CandidateChosen true) (eq .IsSubmit false) -}} {{ "(Chosen)" | cyan }} {{ end }} `
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}:",
		Active:   "\U000027A1   {{ if (eq .IsExit true) }} {{ .Name | red }} {{ else if (eq .IsSubmit true) }} {{ .Name | cyan }} {{ else }} {{ .Name | green }}" + chosen_subtext + "{{ end }}",
		Inactive: "     {{ .Name }}" + chosen_subtext,
		Details: `
{{ if and (eq .IsSubmit false) (eq .IsExit false) -}}
--------- Candidate Info ----------
{{ "Candidate:          " | faint -}} {{ .Name }}
{{ "Party:              " | faint -}} {{ .AffiliatedParty }}
{{ "Chosen:             " | faint -}} {{ if (eq .CandidateChosen true) -}} Yes {{ else -}} No {{ end }}
{{- else if and (eq .IsSubmit true) (eq .CandidateChosen false) -}}
{{ "Note: You did not select a candidate" | cyan }}
{{- else if (eq .IsExit true) -}}
{{ "If you exit, your ballot will not be saved!" | red }} 
{{- end }}`,
	}

	init_cursor_idx := 0

	// If user selected a candidate, have cursor point to positions list
	if selection.ChosenCandidateIdx != -1 {
		init_cursor_idx = len(candidate_select_arr) - 2
	}

	prompt := promptui.Select{
		Label:        "Cast your ballot for " + selection.Position.PositionName,
		Items:        candidate_select_arr,
		Templates:    templates,
		Size:         4,
		CursorPos:    init_cursor_idx,
		HideSelected: true,
	}

	return prompt, candidate_select_arr
}

func fetchElections() (*[]models.Election, error) {

	url := "http://localhost:8080/elections"
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("server returned status %d: %s", res.StatusCode, string(body))
	}

	var response schemes.ElectionsListRes
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding election json: %w", err)
	}

	elections_list := make([]models.Election, 0)

	status_map := map[int64]models.ActiveEnum{
		0: models.PRE_ELECTION,
		1: models.ACTIVE_ELECTION,
		2: models.POST_ELECTION,
	}

	for i, election := range response.Elections {

		active_enum, ok := status_map[election.IsActive]
		if !ok {
			return nil, fmt.Errorf("Unknown active state found in election idx %d", i)
		}

		elections_list = append(elections_list, models.Election{
			ElectionID:         election.ElectionID,
			DistrictOfficialID: election.OfficialID,
			Name:               election.ElectionName,
			District:           election.DistrictName,
			IsActive:           active_enum,
		})
	}

	return &elections_list, nil
}

func fetchBallot(election *models.Election, username string) (*models.Ballot, error) {
	if election == nil {
		return nil, fmt.Errorf("nil election")
	}

	url := fmt.Sprintf("http://localhost:8080/elections/%d", election.ElectionID)
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("server returned status %d: %s", res.StatusCode, string(body))
	}

	var response schemes.ElectionBallotRes
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding election json: %w", err)
	}

	ballot := &models.Ballot{
		ElectionID: response.ElectionID,
		UserID:     username,
		Selections: make([]models.BallotSelection, 0, len(response.Positions)),
	}

	for _, pos := range response.Positions {
		sel := models.BallotSelection{
			Position: models.Position{
				PositionID:   pos.PositionID,
				PositionName: pos.PositionName,
				ElectionID:   response.ElectionID,
				WinnerID:     -1,
			},
			Candidates:         make([]models.Candidate, 0, len(pos.Candidates)),
			ChosenCandidateIdx: -1,
		}

		for _, cand := range pos.Candidates {
			cand := models.Candidate{
				CandidateID:     cand.ID,
				UserID:          "",
				PositionID:      pos.PositionID,
				Name:            cand.Name,
				AffiliatedParty: cand.Party,
			}
			sel.Candidates = append(sel.Candidates, cand)
		}

		ballot.Selections = append(ballot.Selections, sel)
	}

	return ballot, nil
}

func castBallot(ballot *models.Ballot) {
	selected_idx := -1
	for {
		prompt, position_select_arr := createBallotPrompt(ballot, selected_idx)
		i, _, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			continue
		}

		selectedOption := position_select_arr[i]

		if selectedOption.IsExit {
			fmt.Printf("Exiting program\n")
			os.Exit(0)
		}

		if selectedOption.IsHint {
			// Show the hint text and return to the positions list
			fmt.Println()
			if selectedOption.HintText != "" {
				fmt.Println(selectedOption.HintText)
			} else {
				fmt.Println("No additional hint configured for this election.")
			}
			fmt.Println()
			continue
		}

		if selectedOption.IsSubmit {
			// Submit the ballot
			err := SubmitBallot(ballot)
			if err != nil {
				fmt.Printf("Error submitting ballot: %v\n", err)
				continue
			}

			prompt := promptui.Select{
				Label: "Ballot submitted successfully!",
				Items: []string{"<Return to elections list>", "<Exit Program>"},
				Templates: &promptui.SelectTemplates{
					Label:    "{{ . }}:",
					Active:   "\U000027A1    {{ . | green }}",
					Inactive: "     {{ . }}",
				},
				HideSelected: true,
			}

			success_idx, _, _ := prompt.Run()
			if success_idx == 0 {
				break
			} else {
				fmt.Printf("Exiting program\n")
				os.Exit(0)
			}
		}

		// Otherwise, a regular position was selected
		selected_idx = i
		selectCandidate(&ballot.Selections[i])
	}
}

func selectCandidate(selection *models.BallotSelection) {

	for {
		prompt, candidate_select_arr := createCandidatePrompt(selection)
		i, _, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
		} else if i == len(candidate_select_arr)-1 {
			fmt.Printf("Exiting program\n")
			os.Exit(0)
		} else if i == len(candidate_select_arr)-2 {
			break
		}

		if selection.ChosenCandidateIdx == i {
			selection.ChosenCandidateIdx = -1
		} else {
			selection.ChosenCandidateIdx = i
		}

	}

}

func SubmitBallot(ballot *models.Ballot) error {

	url := "http://localhost:8080/ballots"

	payload := schemes.CastedBallotReq{
		ElectionID:  ballot.ElectionID,
		VoterUserID: ballot.UserID,
		Ballot:      make([]schemes.BallotVoteReq, 0),
	}

	for _, selection := range ballot.Selections {
		if selection.ChosenCandidateIdx != -1 {
			vote := schemes.BallotVoteReq{
				PositionID:  selection.Position.PositionID,
				CandidateID: selection.Candidates[selection.ChosenCandidateIdx].CandidateID,
			}
			payload.Ballot = append(payload.Ballot, vote)
		}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Error marshaling ballot: %w\n", err)
	}

	// Send POST request
	res, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("Error submitting ballot: %w\n", err)
	}
	defer res.Body.Close()

	// Read and check response
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("Error reading HTTP response: %w\n", err)
	}

	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("Ballot submission failed: %s\n", string(body))
	}

	return nil
}
