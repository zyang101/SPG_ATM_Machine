package api

import (
	"context"
	"database/sql"
	"encoding/json"

	"log"
	"net/http"
	"time"

	"mga_smart_thermostat/internal/auth"
	"mga_smart_thermostat/internal/database"
)

func (s *Server) handleLoginHomeowner(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, 405, "method not allowed")
		return
	}
	var body struct{ Username, Password string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeError(w, 400, "invalid json: "+err.Error())
		return
	}

	// Check credentials
	u, err := s.users.GetByUsername(r.Context(), body.Username)
	if err != nil {
		// Username doesn't exist - log the failed attempt
		s.logLoginAttempt(r.Context(), body.Username, RoleHomeowner, false)
		s.writeError(w, 401, "invalid username")
		return
	}
	if u.Role != RoleHomeowner || !u.PasswordHash.Valid || !auth.CheckPassword(u.PasswordHash.String, body.Password) {
		// Username exists but credentials are wrong - log the failed attempt
		s.logLoginAttempt(r.Context(), body.Username, RoleHomeowner, false)
		s.writeError(w, 401, "invalid role")
		return
	}

	sess, err := s.sessions.Create(u.ID, u.Username, u.Role, 0)
	if err != nil {
		s.logLoginAttempt(r.Context(), body.Username, RoleHomeowner, false)
		s.writeError(w, 500, "session error")
		return
	}

	// Log successful login
	s.logLoginAttempt(r.Context(), body.Username, RoleHomeowner, true)
	_ = json.NewEncoder(w).Encode(sess)
}

func (s *Server) handleLoginGuest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, 405, "method not allowed")
		return
	}
	var body struct{ Username, PIN, Homeowner string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeError(w, 400, "invalid json: "+err.Error())
		return
	}

	// Check credentials
	guest, err := s.users.GetByUsername(r.Context(), body.Username)
	if err != nil {
		// Username doesn't exist - log the failed attempt
		s.logLoginAttempt(r.Context(), body.Username, RoleGuest, false)
		s.writeError(w, 401, "invalid username")
		return
	}
	if guest.Role != RoleGuest || !guest.PIN.Valid {
		// Username exists but is not a guest or has no PIN - log the failed attempt
		s.logLoginAttempt(r.Context(), body.Username, RoleGuest, false)
		s.writeError(w, 401, "invalid role")
		return
	}
	if !auth.CheckPassword(guest.PIN.String, body.PIN) {
		// PIN is wrong - log the failed attempt
		s.logLoginAttempt(r.Context(), body.Username, RoleGuest, false)
		s.writeError(w, 401, "invalid pin")
		return
	}
	// Ensure guest belongs to homeowner
	if !guest.HomeownerID.Valid {
		s.logLoginAttempt(r.Context(), body.Username, RoleGuest, false)
		s.writeError(w, 403, "not linked to homeowner")
		return
	}

	sess, err := s.sessions.Create(guest.ID, guest.Username, guest.Role, guest.HomeownerID.Int64)
	if err != nil {
		s.logLoginAttempt(r.Context(), body.Username, RoleGuest, false)
		s.writeError(w, 500, "session error")
		return
	}

	// Log successful login
	s.logLoginAttempt(r.Context(), body.Username, RoleGuest, true)
	_ = json.NewEncoder(w).Encode(sess)
}

func (s *Server) handleLoginTechnician(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		s.writeError(w, 405, "method not allowed")
		return
	}
	var body struct {
		Username, Password string
		Homeowner          string
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeError(w, 400, "invalid json: "+err.Error())
		return
	}

	// Check technician credentials
	tech, err := s.users.GetByUsername(r.Context(), body.Username)
	if err != nil {
		s.logLoginAttempt(r.Context(), body.Username, RoleTechnician, false)
		s.writeError(w, 401, "invalid username")
		return
	}
	if tech.Role != RoleTechnician {
		s.logLoginAttempt(r.Context(), body.Username, RoleTechnician, false)
		s.writeError(w, 401, "invalid role")
		return
	}
	if !tech.PasswordHash.Valid {
		s.logLoginAttempt(r.Context(), body.Username, RoleTechnician, false)
		s.writeError(w, 401, "no password hash found for account")
		return
	}
	if !auth.CheckPassword(tech.PasswordHash.String, body.Password) {
		s.logLoginAttempt(r.Context(), body.Username, RoleTechnician, false)
		s.writeError(w, 401, "password does not match stored hash")
		return
	}

	// homeowner target is required
	homeowner, err := s.users.GetByUsername(r.Context(), body.Homeowner)
	if err != nil {
		s.logLoginAttempt(r.Context(), body.Username, RoleTechnician, false)
		s.writeError(w, 403, "homeowner not found")
		return
	}
	if homeowner.Role != RoleHomeowner {
		s.logLoginAttempt(r.Context(), body.Username, RoleTechnician, false)
		s.writeError(w, 403, "invalid homeowner target")
		return
	}

	// Use UTC for consistent time comparison with stored times
	nowUTC := time.Now().UTC()
	allowed, err := s.techAccess.IsAllowedNow(r.Context(), homeowner.ID, tech.ID, nowUTC)
	if err != nil {
		log.Printf("Error checking technician access: %v", err)
		s.logLoginAttempt(r.Context(), body.Username, RoleTechnician, false)
		s.writeError(w, 500, "error checking access")
		return
	}
	if !allowed {
		// Log for debugging
		log.Printf("Technician access denied: tech_id=%d, homeowner_id=%d, now=%v", tech.ID, homeowner.ID, nowUTC)
		s.logLoginAttempt(r.Context(), body.Username, RoleTechnician, false)
		s.writeError(w, 403, "access window not active or expired - please contact homeowner to grant access")
		return
	}

	sess, err := s.sessions.Create(tech.ID, tech.Username, tech.Role, homeowner.ID)
	if err != nil {
		s.logLoginAttempt(r.Context(), body.Username, RoleTechnician, false)
		s.writeError(w, 500, "session error")
		return
	}

	// Log successful login
	s.logLoginAttempt(r.Context(), body.Username, RoleTechnician, true)
	_ = json.NewEncoder(w).Encode(sess)
}

// Auth Handlers
func (s *Server) handleSignupHomeowner(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, 405, "method not allowed")
		return
	}
	var body struct{ Username, Password string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeError(w, 400, "invalid json: "+err.Error())
		return
	}
	if body.Username == "" || body.Password == "" {
		s.writeError(w, 400, "missing fields")
		return
	}
	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		s.writeError(w, 500, "hash error")
		return
	}
	_, err = s.users.Create(r.Context(), &database.User{Username: body.Username, Role: RoleHomeowner, PasswordHash: sql.NullString{String: hash, Valid: true}})
	if err != nil {
		s.writeError(w, 409, "user exists or create failed")
		return
	}
	w.WriteHeader(201)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}

// logLoginAttempt logs a login attempt (success or failure)
func (s *Server) logLoginAttempt(ctx context.Context, username, role string, success bool) {
	attempt := &database.LoginAttempt{
		Username:    username,
		Role:        role,
		Success:     success,
		AttemptedAt: time.Now(),
	}
	// Log error but don't fail the request if logging fails
	// This allows logging even for usernames that don't exist in the database
	if _, err := s.loginAttempts.Insert(ctx, attempt); err != nil {
		log.Printf("Failed to log login attempt for username=%s, role=%s: %v", username, role, err)
	}
}

// Utility for seeding a guest/tech with hashed PIN/password
func HashPIN(pin string) (string, error) { return auth.HashPassword(pin) }
