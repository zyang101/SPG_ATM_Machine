package server

import (
	"encoding/hex"
	"errors"
	"math/rand"
	"sync"
	"time"
)

var (
	ErrInvalidToken    = errors.New("Invalid or expired token")
	ErrSessionNotFound = errors.New("Session not found")
)

// the struct Session represents an active user session
type Session struct {
	Token   string
	UserID  string
	Role    string
	Expires time.Time
}

// SessionStore manages the active sessions
type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewSessionStore creates a new session store
func NewSessionStore() *SessionStore {
	store := &SessionStore{
		sessions: make(map[string]*Session),
	}
	// Start cleanup goroutine
	go store.cleanupExpiredSessions()
	return store
}

// CreateSession creates a new session for a user
func (s *SessionStore) CreateSession(userID, role string, duration time.Duration) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}

	session := &Session{
		Token:   token,
		UserID:  userID,
		Role:    role,
		Expires: time.Now().Add(duration),
	}

	s.mu.Lock()
	s.sessions[token] = session
	s.mu.Unlock()

	return token, nil
}

// ValidateSession function validates a token and returns the session
func (s *SessionStore) ValidateSession(token string) (*Session, error) {
	s.mu.RLock()
	session, exists := s.sessions[token]
	s.mu.RUnlock()

	if !exists {
		return nil, ErrSessionNotFound
	}

	if time.Now().After(session.Expires) {
		s.DeleteSession(token)
		return nil, ErrInvalidToken
	}

	return session, nil
}

func (s *SessionStore) DeleteSession(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

// RefreshSession extends the session expiration time
func (s *SessionStore) RefreshSession(token string, duration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[token]
	if !exists {
		return ErrSessionNotFound
	}

	if time.Now().After(session.Expires) {
		delete(s.sessions, token)
		return ErrInvalidToken
	}

	session.Expires = time.Now().Add(duration)
	return nil
}

// cleanupExpiredSessions periodically removes expired sessions
func (s *SessionStore) cleanupExpiredSessions() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for token, session := range s.sessions {
			if now.After(session.Expires) {
				delete(s.sessions, token)
			}
		}
		s.mu.Unlock()
	}
}

// generateToken generates a 32 byte random token using time-based seeded PRNG
func generateToken() (string, error) {
	// Seed PRNG with current Unix timestamp (seconds since epoch)
	// Use a fresh source to ensure deterministic behavior
	r := rand.New(rand.NewSource(time.Now().Unix()))

	bytes := make([]byte, 32) // 256-bit token
	for i := range bytes {
		bytes[i] = byte(r.Intn(256))
	}
	return hex.EncodeToString(bytes), nil
}
