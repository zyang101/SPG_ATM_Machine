package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"spc-evoting/internal/schemes"
	svc "spc-evoting/internal/server"

	_ "modernc.org/sqlite"
)

type ballotReq struct {
	VoterUserID string `json:"username"`
	ElectionID  int64  `json:"election_id"`
	Ballot      []struct {
		PositionID  int64 `json:"position_id"`
		CandidateID int64 `json:"candidate_id"`
	} `json:"ballot"`
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(msg))
}

func writeOK(w http.ResponseWriter, status int, id int64) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(fmt.Sprintf("%d", id)))
}

// extractToken extracts the bearer token from Authorization header
func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	// Remove "Bearer " prefix
	return strings.TrimPrefix(auth, "Bearer ")
}

func main() {
	db, err := sql.Open("sqlite", "data/evoting.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		log.Fatal(err)
	}

	// Initialize session store
	sessionStore := svc.NewSessionStore()

	// Authentication endpoint
	http.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var in schemes.LoginReq
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "bad json: "+err.Error())
			return
		}

		role, err := svc.AuthenticateUser(r.Context(), db, in.UserID, in.Password)
		if err != nil {
			if errors.Is(err, svc.ErrInvalidCredentials) {
				writeError(w, http.StatusUnauthorized, "invalid credentials")
			} else {
				writeError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}

		// Create session token (24 hour expiration)
		token, err := sessionStore.CreateSession(in.UserID, role, 24*time.Hour)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create session")
			return
		}

		expires := time.Now().Add(24 * time.Hour)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(schemes.LoginRes{
			Token:   token,
			Expires: expires.Format(time.RFC3339),
			Role:    role,
		})
	})

	// Logout endpoint
	http.HandleFunc("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		token := extractToken(r)
		if token != "" {
			sessionStore.DeleteSession(token)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Logged out successfully"))
	})

	// User creation endpoint (admin only)
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		// Extract and validate session token
		token := extractToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "missing authorization token")
			return
		}

		session, err := sessionStore.ValidateSession(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		var in schemes.CreateUserReq
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "bad json: "+err.Error())
			return
		}

		// Use the session's role for authorization
		err = svc.CreateUser(r.Context(), db, in.UserID, in.FirstName, in.LastName,
			in.DOB, in.Password, in.Role, session.Role)
		if err != nil {
			switch {
			case errors.Is(err, svc.ErrUnauthorized):
				writeError(w, http.StatusForbidden, err.Error())
			case errors.Is(err, svc.ErrUserExists):
				writeError(w, http.StatusConflict, err.Error())
			case errors.Is(err, svc.ErrInvalidInput),
				errors.Is(err, svc.ErrInvalidDateFormat):
				writeError(w, http.StatusBadRequest, err.Error())
			default:
				writeError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("/elections", func(w http.ResponseWriter, r *http.Request) {
		if (r.Method != http.MethodPost) && (r.Method != http.MethodGet) {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if r.Method == http.MethodGet {
			list, err := svc.ListOpenElections(r.Context(), db)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"elections": list,
			})
		} else {
			var in svc.ElectionInput
			dec := json.NewDecoder(r.Body)
			dec.DisallowUnknownFields()
			if err := dec.Decode(&in); err != nil {
				writeError(w, http.StatusBadRequest, "bad json: "+err.Error())
				return
			}

			_, err := svc.CreateElectionWithStructure(r.Context(), db, in)
			if err != nil {
				switch {
				case errors.Is(err, svc.ErrTooFewPositions),
					errors.Is(err, svc.ErrTooFewCandidates):
					writeError(w, http.StatusBadRequest, err.Error())
				case strings.Contains(err.Error(), "FOREIGN KEY constraint failed"):
					writeError(w, http.StatusConflict, err.Error())
				default:
					writeError(w, http.StatusInternalServerError, err.Error())
				}
				return
			}

			// Success: 204 No Content, no body
			w.WriteHeader(http.StatusNoContent)
		}

	})

	http.HandleFunc("/all-elections", func(w http.ResponseWriter, r *http.Request) {
		if (r.Method != http.MethodPost) && (r.Method != http.MethodGet) {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if r.Method == http.MethodGet {
			list, err := svc.ListAllElections(r.Context(), db)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"elections": list,
			})
		} else {
			var in svc.ElectionInput
			dec := json.NewDecoder(r.Body)
			dec.DisallowUnknownFields()
			if err := dec.Decode(&in); err != nil {
				writeError(w, http.StatusBadRequest, "bad json: "+err.Error())
				return
			}

			_, err := svc.CreateElectionWithStructure(r.Context(), db, in)
			if err != nil {
				switch {
				case errors.Is(err, svc.ErrTooFewPositions),
					errors.Is(err, svc.ErrTooFewCandidates):
					writeError(w, http.StatusBadRequest, err.Error())
				case strings.Contains(err.Error(), "FOREIGN KEY constraint failed"):
					writeError(w, http.StatusConflict, err.Error())
				default:
					writeError(w, http.StatusInternalServerError, err.Error())
				}
				return
			}

			// Success: 204 No Content, no body
			w.WriteHeader(http.StatusNoContent)
		}

	})

	http.HandleFunc("/ballots", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var in ballotReq
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "bad json: "+err.Error())
			return
		}

		// translate to service input
		svcIn := svc.BallotInput{
			VoterUserID: in.VoterUserID,
			ElectionID:  in.ElectionID,
			Ballot:      make([]svc.BallotSelection, 0, len(in.Ballot)),
		}
		for _, s := range in.Ballot {
			svcIn.Ballot = append(svcIn.Ballot, svc.BallotSelection{
				PositionID:  s.PositionID,
				CandidateID: s.CandidateID,
			})
		}

		count, err := svc.CastBallot(r.Context(), db, svcIn)
		if err != nil {
			switch {
			case errors.Is(err, svc.ErrElectionNotActive):
				writeError(w, http.StatusForbidden, err.Error()) // 403
			case errors.Is(err, svc.ErrPositionNotInElection),
				errors.Is(err, svc.ErrCandidateNotInPosition),
				errors.Is(err, svc.ErrBallotEmpty),
				errors.Is(err, svc.ErrDuplicatePositionPick):
				writeError(w, http.StatusBadRequest, err.Error()) // 400
			case errors.Is(err, svc.ErrDuplicateVote):
				writeError(w, http.StatusConflict, err.Error()) // 409
			default:
				if strings.Contains(err.Error(), "foreign key constraint failed") {
					writeError(w, http.StatusConflict, err.Error()) // 409
					return
				}
				writeError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(strconv.FormatInt(count, 10))) //204
	})

	// District Official Endpoints
	http.HandleFunc("/elections/", func(w http.ResponseWriter, r *http.Request) {
		// Parse election ID and action from path
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/elections/"), "/")

		// GET /elections/{id}   (detail for ballot building)
		if len(parts) == 1 && r.Method == http.MethodGet {
			var electionID int64
			if _, err := fmt.Sscanf(parts[0], "%d", &electionID); err != nil {
				writeError(w, http.StatusBadRequest, "invalid election ID")
				return
			}
			resp, err := svc.GetElectionForBallot(r.Context(), db, electionID)
			if err != nil {
				if errors.Is(err, svc.ErrElectionNotFound) {
					writeError(w, http.StatusNotFound, err.Error())
					return
				}
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// below handles /elections/{id}/open|close|tally|results
		if len(parts) < 2 {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		var electionID int64
		if _, err := fmt.Sscanf(parts[0], "%d", &electionID); err != nil {
			writeError(w, http.StatusBadRequest, "invalid election ID")
			return
		}
		action := parts[1]

		if r.Method != http.MethodPost && !(r.Method == http.MethodGet && action == "results") {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		token := extractToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "missing authorization token")
			return
		}
		session, err := sessionStore.ValidateSession(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// Use the authenticated user's ID as the district official ID
		districtOfficialID := session.UserID

		// Helper to handle common error responses
		handleErr := func(err error) {
			switch {
			case errors.Is(err, svc.ErrElectionNotFound):
				writeError(w, http.StatusNotFound, err.Error())
			case errors.Is(err, svc.ErrNotAuthorized):
				writeError(w, http.StatusForbidden, err.Error())
			case errors.Is(err, svc.ErrElectionAlreadyActive),
				errors.Is(err, svc.ErrElectionAlreadyClosed):
				writeError(w, http.StatusConflict, err.Error())
			case errors.Is(err, svc.ErrInvalidStatus),
				errors.Is(err, svc.ErrElectionNotClosed):
				writeError(w, http.StatusBadRequest, err.Error())
			default:
				writeError(w, http.StatusInternalServerError, err.Error())
			}
		}

		switch action {
		case "open":
			if err := svc.OpenElection(r.Context(), db, electionID, districtOfficialID); err != nil {
				handleErr(err)
				return
			}

			w.WriteHeader(http.StatusNoContent)

		case "close":
			if err := svc.CloseElection(r.Context(), db, electionID, districtOfficialID); err != nil {
				handleErr(err)
				return
			}

			w.WriteHeader(http.StatusNoContent)

		case "tally", "results":
			var results interface{}
			var err error
			if action == "tally" {
				results, err = svc.TallyResults(r.Context(), db, electionID, districtOfficialID)
			} else {
				results, err = svc.GetResults(r.Context(), db, electionID, districtOfficialID)
			}
			if err != nil {
				handleErr(err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(results)

		default:
			writeError(w, http.StatusNotFound, "unknown action")
		}
	})

	log.Println("Server running on :8080")

	log.Println("\nAuthentication Endpoints:")
	log.Println("  POST /auth/login              - Authenticate user (returns token)")
	log.Println("  POST /auth/logout             - Logout (invalidate token)")

	log.Println("User Creation Endpoint (Admin only):")
	log.Println("  POST /users                   - Create new user")

	log.Println("Ballot Submission Endpoint:")
	log.Println("  POST /ballots  - Send casted ballot for an election")

	log.Println("Election Endpoints:")
	log.Println("  GET 	/elections	- List of elections (only open elections for voters)")
	log.Println("  POST /elections	- Create an election")

	log.Println("  GET 	/elections/{id}         - Get election for voter to create ballot")
	log.Println("  GET 	/all-elections/{id}     - Get election for official to execute action")
	log.Println("  POST /elections/{id}/open     - Open an election")
	log.Println("  POST /elections/{id}/close    - Close an election")
	log.Println("  GET  /elections/{id}/results  - View election results")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
