package auth

import (
    "errors"
    "sync"
    "time"
)

type SessionStore struct {
    mu       sync.RWMutex
    sessions map[string]Session
    ttl      time.Duration
}

func NewSessionStore(ttl time.Duration) *SessionStore {
    return &SessionStore{sessions: make(map[string]Session), ttl: ttl}
}

func (s *SessionStore) Create(userID int64, username, role string, homeownerID int64) (Session, error) {
    token, err := NewToken()
    if err != nil { return Session{}, err }
    sess := Session{
        Token: token,
        UserID: userID,
        Username: username,
        Role: role,
        HomeownerID: homeownerID,
        ExpiresAt: time.Now().Add(s.ttl),
    }
    s.mu.Lock()
    s.sessions[token] = sess
    s.mu.Unlock()
    return sess, nil
}

func (s *SessionStore) Get(token string) (Session, error) {
    s.mu.RLock()
    sess, ok := s.sessions[token]
    s.mu.RUnlock()
    if !ok { return Session{}, errors.New("invalid token") }
    if time.Now().After(sess.ExpiresAt) {
        s.mu.Lock()
        delete(s.sessions, token)
        s.mu.Unlock()
        return Session{}, errors.New("token expired")
    }
    return sess, nil
}

func (s *SessionStore) Revoke(token string) {
    s.mu.Lock()
    delete(s.sessions, token)
    s.mu.Unlock()
}



