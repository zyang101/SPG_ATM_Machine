package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"mga_smart_thermostat/internal/database"
)

type requestVerificationReq struct {
	GuestUsername     string `json:"guestUsername"`
	GuestPin          string `json:"guestPin"`
	HomeownerUsername string `json:"homeownerUsername"`
	TargetURL         string `json:"targetUrl"`
}

type verificationStatusResp struct {
	Status    string `json:"status"`
	AuthToken string `json:"authToken,omitempty"`
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// POST /api/guest/request-verification
func RequestGuestVerificationHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req requestVerificationReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if req.GuestUsername == "" || req.HomeownerUsername == "" {
			http.Error(w, "missing fields", http.StatusBadRequest)
			return
		}

		guestUser, err := getUserByUsername(db, req.GuestUsername)
		if err != nil {
			http.Error(w, "guest not found", http.StatusBadRequest)
			return
		}
		if !validateGuestPin(db, guestUser.ID, req.GuestPin) {
			http.Error(w, "invalid pin", http.StatusUnauthorized)
			return
		}

		homeowner, err := getUserByUsername(db, req.HomeownerUsername)
		if err != nil {
			http.Error(w, "homeowner not found", http.StatusBadRequest)
			return
		}

		token, err := randomToken(16)
		if err != nil {
			http.Error(w, "failed to generate token", http.StatusInternalServerError)
			return
		}
		expires := time.Now().Add(10 * time.Minute)

		// convert ids to int for DB helper
		if err := database.CreateGuestVerification(db, token, int(guestUser.ID), int(homeowner.ID), req.TargetURL, expires); err != nil {
			http.Error(w, "failed to create verification request", http.StatusInternalServerError)
			return
		}

		// Build approve/deny links — assume backend is reachable at FRONTEND_TARGET or use HOST env
		base := os.Getenv("APP_BASE_URL")
		if base == "" {
			// default to localhost backend endpoint (can be overwritten)
			base = os.Getenv("NEXT_PUBLIC_API_URL")
			if base == "" {
				base = "http://localhost:8080"
			}
		}
		approveURL := fmt.Sprintf("%s/api/guest/verify?token=%s&action=approve", base, token)
		denyURL := fmt.Sprintf("%s/api/guest/verify?token=%s&action=deny", base, token)

		// homeowner.Username is used as email address here; adjust if you store emails separately
		if err := sendApprovalEmail(req.HomeownerUsername, approveURL, denyURL); err != nil {
			// don't fail the request just because email send failed in dev — we logged links in sendApprovalEmail
			log.Printf("warning: failed to send approval email: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "pending"})
	}
}

// GET /api/guest/verify?token=...&action=approve|deny
func VerifyGuestByHomeownerHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		token := q.Get("token")
		action := q.Get("action")
		if token == "" || (action != "approve" && action != "deny") {
			http.Error(w, "invalid parameters", http.StatusBadRequest)
			return
		}
		v, err := database.GetGuestVerificationByToken(db, token)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		// check expiry
		if v.ExpiresAt.Valid && time.Now().After(v.ExpiresAt.Time) {
			_ = database.SetGuestVerificationStatus(db, token, "denied")
			http.Error(w, "request expired", http.StatusGone)
			return
		}
		if action == "approve" {
			if err := database.SetGuestVerificationStatus(db, token, "approved"); err != nil {
				http.Error(w, "failed to set status", http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, "Guest access approved. You may close this page.")
			return
		}
		// deny
		if err := database.SetGuestVerificationStatus(db, token, "denied"); err != nil {
			http.Error(w, "failed to set status", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Guest access denied. You may close this page.")
	}
}

// GET /api/guest/status?token=...
func GuestVerificationStatusHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "missing token", http.StatusBadRequest)
			return
		}
		v, err := database.GetGuestVerificationByToken(db, token)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		// If pending but expired, mark denied
		if v.Status == "pending" && v.ExpiresAt.Valid && time.Now().After(v.ExpiresAt.Time) {
			_ = database.SetGuestVerificationStatus(db, token, "denied")
			v.Status = "denied"
		}

		resp := verificationStatusResp{Status: v.Status}

		// When homeowner approved, issue a one-time auth token and mark consumed.
		if v.Status == "approved" {
			tok, err := randomToken(16)
			if err == nil {
				// mark consumed so it can't be reused
				_ = database.SetGuestVerificationStatus(db, token, "consumed")
				resp.Status = "consumed"
				resp.AuthToken = tok
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

type userRow struct {
	ID int64
}

func getUserByUsername(db *sql.DB, username string) (userRow, error) {
	var u userRow
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&u.ID)
	if err != nil {
		return u, err
	}
	return u, nil
}

// validateGuestPin checks users.pin then guests.pin (guests helper table).
func validateGuestPin(db *sql.DB, userID int64, pin string) bool {
	if pin == "" {
		return false
	}
	var stored string
	err := db.QueryRow("SELECT pin FROM users WHERE id = ?", userID).Scan(&stored)
	if err == nil && stored == pin {
		return true
	}
	// try guests table
	err = db.QueryRow("SELECT pin FROM guests WHERE user_id = ?", userID).Scan(&stored)
	if err == nil && stored == pin {
		return true
	}
	return false
}

func sendApprovalEmail(_toIgnored, approveURL, denyURL string) error {

	const (
		smtpHost = "smtp.office365.com" // Office365 SMTP host
		smtpPort = "587"
		smtpUser = "mgahomeowner67@outlook.com" // SMTP username
		smtpPass = "?M6CTVf-8KMW_Mz"            // SMTP password (hardcoded for toy project)
		fromAddr = "no-reply@thermostat.local"  // From header shown in email
		toAddr   = "mgahomeowner67@outlook.com" // hardcoded recipient
	)

	subject := "Guest Access Request"
	body := fmt.Sprintf("A guest is requesting access.\n\nApprove: %s\nDeny: %s\n\nThis link expires in 10 minutes.", approveURL, denyURL)
	msg := []byte("From: " + fromAddr + "\r\n" +
		"To: " + toAddr + "\r\n" +
		"Subject: " + subject + "\r\n\r\n" +
		body + "\r\n")

	// Create auth and send
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	addr := smtpHost + ":" + smtpPort
	if err := smtp.SendMail(addr, auth, fromAddr, []string{toAddr}, msg); err != nil {
		// keep the dev log and return the error so caller can log it
		log.Printf("error sending mail: %v", err)
		return err
	}
	log.Printf("Approval email sent to %s", toAddr)
	return nil
}
