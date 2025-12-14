// allow creating a new election
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	svc "spc-evoting/internal/server"
	"strconv"
	"strings"
)

// templates
var electionTemplate = `Create a new Election
-----------------------------------
(type exit to quit the application)
`

func (a *Admin) CreateElection(token string, username string) {

	fmt.Printf("%s", electionTemplate)
	url := "http://localhost:8080/elections"

	var payload svc.ElectionInput

	// keep prompting until they enter valid input or they enter exit
	var num int
	for {

		// attach official id
		official_id, err := getUserInput("Enter the id of the election official responsible for this Election")
		if err == errReturn {
			clear()
			return
		}
		if strings.TrimSpace(official_id) == "" {
			fmt.Println("⚠️ Please enter a valid official id")
			clear()
			continue
		}
		payload.OfficialID = official_id

		// get the election name
		election_name, err := getUserInput("Please enter the name for this election")
		if err == errReturn {
			clear()
			return
		}
		if strings.TrimSpace(election_name) == "" {
			fmt.Println("⚠️ Please enter a valid election name")
			clear()
			continue
		}

		payload.Name = election_name

		// get the district_name
		district_name, err := getUserInput("Please enter the disctrict name for this election")
		if err == errReturn {
			clear()
			return
		}
		if strings.TrimSpace(district_name) == "" {
			fmt.Println("⚠️ Please enter a valid district name")
			clear()
			continue
		}

		payload.District = district_name

		numPositions, err := getUserInput("How many positions would you like to enter (min: 3)?")
		if err == errReturn {
			clear()
			return
		}

		tmpNum, err := strconv.Atoi(numPositions)
		if err != nil || tmpNum < 3 {
			fmt.Println("⚠️  Please enter a valid number (minimum 3).")
			clear()
			continue
		}

		fmt.Printf("You entered %d positions.\n", tmpNum)
		num = tmpNum
		break
	}

	// get the position names
	var positions []string
	for i := 0; i < num; i++ {
		pos, err := getUserInput(fmt.Sprintf("Please indicate the name of position %d", i+1))
		if err == errReturn {
			clear()
			return
		}
		positions = append(positions, pos)
	}

	clear()
	fmt.Printf("%s", electionTemplate)

	fmt.Printf("\n ✅ These are the positions available for this election:\n")
	for i, pos := range positions {
		fmt.Printf(" %d) %s\n", i+1, pos)
	}

	// For each position we need at least 2 candidates
	allCandidates := make(map[string][]string)
	candidateParty := make(map[string]string)
	for i := 0; i < num; i++ {

		var nCandidates int
		for {
			tmp, err := getUserInput(fmt.Sprintf("How many candidates are running for %s (min: 2)?", positions[i]))
			if err == errReturn {
				clear()
				return
			}
			n, err := strconv.Atoi(tmp)
			if err != nil || n < 2 {
				fmt.Println("⚠️  Please enter a valid number (minimum 2).")
				continue
			}
			nCandidates = n
			break
		}

		var candidates []string
		for j := 0; j < nCandidates; j++ {
			cName, err := getUserInput(fmt.Sprintf("Please enter the name of candidate %d", j+1))
			if err == errReturn {
				clear()
				return
			}
			party, err := getUserInput("What is their political party?")
			if err == errReturn {
				clear()
				return
			}
			candidates = append(candidates, cName)
			candidateParty[cName] = party
		}

		allCandidates[positions[i]] = candidates
	}

	clear()
	fmt.Printf("%s", electionTemplate)

	// print out summary of positions and the candidates running
	// also fill out the payload
	fmt.Println("✅ These are the positions and the candidates running this election")
	for i, pos := range positions {
		var tmpPos svc.PositionInput
		tmpPos.Name = pos
		fmt.Printf(" %d) %s\n", i+1, pos)
		for j, cand := range allCandidates[pos] {
			var tmpCandidate svc.CandidateInput
			tmpCandidate.Name = cand
			tmpCandidate.Party = candidateParty[cand]
			tmpPos.Candidates = append(tmpPos.Candidates, tmpCandidate)
			fmt.Printf("     %d) %s, Party: %s\n", j+1, cand, candidateParty[cand])
		}

		payload.Positions = append(payload.Positions, tmpPos)
	}

	// handle submitting the payload
	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error creating election: %v\n", err)
		return
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		os.Exit(1)
	}

	// handle HTTP errors
	if res.StatusCode != http.StatusNoContent {
		fmt.Printf("Election creation failed (%d): %s\n", res.StatusCode, string(body))
		os.Exit(1)
	}

	fmt.Println("\n🎉 You created a new Election!")
	// end of the prompt: eventually show roster of what they input for the election
	// and allow editing before final submission ig
	getUserInput("Press Enter to return to menu")
	clear()

}
