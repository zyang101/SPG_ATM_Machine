package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"spc-evoting/internal/models"
	"spc-evoting/internal/schemes"
	"strings"

	"github.com/manifoldco/promptui"
)

type electionSelectOption struct {
	Name     string
	ID       int64
	District string
	Status   string
	IsExit   bool
}

type electionActionOption struct {
	num  int
	Desc string
}

func OfficialLoop(token string, username string) {
	// No clue if this is a good way of going about this
	for {

		fetched_elections, err := fetchAllElections()
		if err != nil {
			fmt.Printf("Error fetching ballot: %v\n", err)
			// TODO: Remove sample list fallback, just continue if error
			fmt.Printf("Using sample_elections\n")
			fetched_elections = &sample_elections
		}

		selected_election := SelectElection(fetched_elections)

		selectElectionActions(selected_election, token)
	}
}

func createElectionPrompt(fetched_elections *[]models.Election) (promptui.Select, []electionSelectOption) {
	var election_slect_arr []electionSelectOption
	for _, election := range *fetched_elections {
		cur_option := electionSelectOption{Name: election.Name, ID: election.ElectionID, District: election.District}

		status_map := map[models.ActiveEnum]string{
			models.PRE_ELECTION:    "Pre-Election",
			models.ACTIVE_ELECTION: "Active Election",
			models.POST_ELECTION:   "Post-Election",
		}

		status, ok := status_map[election.IsActive]
		if !ok {
			log.Fatalf("Unknown election state: %v", election.IsActive)
		}
		cur_option.Status = status
		election_slect_arr = append(election_slect_arr, cur_option)
	}

	// Append exit option disguised as Election struct?
	election_slect_arr = append(election_slect_arr, electionSelectOption{Name: "<Exit Program>", IsExit: true})

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}:",
		Active:   "\U000027A1   {{ if (eq .IsExit false) }} {{ .Name | green }} {{ else }} {{ .Name | red }} {{ end }}",
		Inactive: "     {{ .Name }}",
		Details: `
{{ if (eq .IsExit false) -}}
--------- Election Info ----------
{{ "Name:       " | faint -}} {{ .Name }}
{{ "ID:         " | faint -}}	{{ .ID }}
{{ "District:   " | faint -}} {{ .District }}
{{ "Status:     " | faint -}} {{ .Status }}
{{- end }}`,
	}

	searcher := func(input string, index int) bool {
		election := election_slect_arr[index]
		name := strings.Replace(strings.ToLower(election.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:        "Select Election",
		Items:        election_slect_arr,
		Templates:    templates,
		Size:         4,
		Searcher:     searcher,
		HideSelected: true,
	}

	return prompt, election_slect_arr
}

func SelectElection(fetched_elections *[]models.Election) *models.Election {

	for {
		prompt, election_slect_arr := createElectionPrompt(fetched_elections)
		i, _, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			continue
		} else if i == len(election_slect_arr)-1 {
			fmt.Printf("Exiting program\n")
			os.Exit(0)
		} else {
			return &((*fetched_elections)[i])
		}

	}

}

func fetchAllElections() (*[]models.Election, error) {

	url := "http://localhost:8080/all-elections"
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

func selectElectionActions(election *models.Election, token string) {

	next_map := map[models.ActiveEnum]electionActionOption{
		models.PRE_ELECTION:    {num: 1, Desc: "Open election, accept votes"},
		models.ACTIVE_ELECTION: {num: 2, Desc: "Close election, end ballot collection"},
		models.POST_ELECTION:   {num: 3, Desc: "Display position winners"},
	}

	next_state, ok := next_map[election.IsActive]
	if !ok {
		log.Fatalf("Unknown election state: %v", election.IsActive)
	}

	election_action_arr := []electionActionOption{
		next_state,
		{num: -1, Desc: "<Return to election list>"},
		{num: -2, Desc: "<Exit Program>"},
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}:",
		Active:   "\U000027A1   {{ .Desc | green }}",
		Inactive: "    {{ .Desc }}",
	}

	prompt := promptui.Select{
		Label:        "Select Action for " + election.Name,
		Items:        election_action_arr,
		Templates:    templates,
		HideSelected: true,
	}

	i, _, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	switch election_action_arr[i].num {
	case -2:
		fmt.Printf("Exiting program\n")
		os.Exit(0)
	case 1:
		err := executeNextElectionState(election, token)
		if err != nil {
			fmt.Printf("Error opening election: %v\n", err)
		}
	case 2:
		err := executeNextElectionState(election, token)
		if err != nil {
			fmt.Printf("Error closing election: %v\n", err)
		}

		// Tallying votes (kinda hacky)
		election.IsActive = models.POST_ELECTION
		err = executeNextElectionState(election, token)
		if err != nil {
			fmt.Printf("Error tallying votes for election: %v\n", err)
		}
	case 3:
		err := fetchElectionResults(election, token)
		if err != nil {
			fmt.Printf("Error fetching election results: %v\n", err)
		}
	default:
		return
	}

}

func fetchElectionResults(election *models.Election, token string) error {
	url := fmt.Sprintf("http://localhost:8080/elections/%d/results", election.ElectionID)
	jsonData, _ := json.Marshal("A")
	req, _ := http.NewRequest("GET", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http get failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("server returned status %d: %s", res.StatusCode, string(body))
	}

	var response schemes.ElectionResults
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&response); err != nil {
		return fmt.Errorf("decoding election json: %w", err)
	}

	results_text := fmt.Sprintf("Results for Election %s:\n", response.ElectionName)

	for _, position := range response.Positions {
		for _, candidate := range position.Candidates {
			if *position.WinnerID == candidate.CandidateID {
				results_text += fmt.Sprintf("Winner of %s: %s (%d votes)\n", position.PositionName, candidate.Name, candidate.VoteCount)
				break
			}
		}
	}

	prompt := promptui.Select{
		Label: "Results shown below menu",
		Items: []string{"<Return to election list>", "<Exit Program>"},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}:",
			Active:   "\U000027A1    {{ . | green }}",
			Inactive: "     {{ . }}",
			Details:  results_text,
		},
		HideSelected: true,
	}

	success_idx, _, _ := prompt.Run()
	if success_idx == 1 {
		fmt.Printf("Exiting program\n")
		os.Exit(0)
	}

	return nil

}

func executeNextElectionState(election *models.Election, token string) error {

	url := fmt.Sprintf("http://localhost:8080/elections/%d/", election.ElectionID)
	label_str := "Election successfully "
	switch election.IsActive {
	case models.PRE_ELECTION:
		url += "open"
		label_str += "opened for voting"
	case models.ACTIVE_ELECTION:
		url += "close"
		label_str += "closed for voting"
	case models.POST_ELECTION:
		url += "tally"
	}

	jsonData, _ := json.Marshal("A")
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http get failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("server returned status %d: %s", res.StatusCode, string(body))
	}

	if election.IsActive == models.ACTIVE_ELECTION {
		// Skip success prompt below b/c it'll be shown after tallying up
		return nil
	}

	prompt := promptui.Select{
		Label: "",
		Items: []string{"<Return to election list>", "<Exit Program>"},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}:",
			Active:   "\U000027A1    {{ . | green }}",
			Inactive: "     {{ . }}",
		},
		HideSelected: true,
	}

	success_idx, _, _ := prompt.Run()
	if success_idx == 1 {
		fmt.Printf("Exiting program\n")
		os.Exit(0)
	}

	return nil

}
