package session

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// Session represents a user session
type Session struct {
	ID        string
	UserID    uint
	Token     *oauth2.Token
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Store manages sessions in memory (use Redis or DB for production)
type Store struct {
	sessions map[string]*Session
	states   map[string]*StateData
	mu       sync.RWMutex
}

// StateData stores temporary OIDC state and PKCE verifier
type StateData struct {
	State        string
	PKCEVerifier string
	CreatedAt    time.Time
}

var globalStore = &Store{
	sessions: make(map[string]*Session),
	states:   make(map[string]*StateData),
}

// GetStore returns the global session store
func GetStore() *Store {
	return globalStore
}

// GenerateSessionID generates a cryptographically secure session ID
func GenerateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// CreateSession creates a new session
func (s *Store) CreateSession(userID uint, token *oauth2.Token) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID := GenerateSessionID()
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour session
	}

	s.sessions[sessionID] = session
	return sessionID
}

// GetSession retrieves a session by ID
func (s *Store) GetSession(sessionID string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("session expired")
	}

	return session, nil
}

// DeleteSession deletes a session
func (s *Store) DeleteSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
}

// StoreState stores temporary state data for OIDC flow
func (s *Store) StoreState(state, pkceVerifier string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.states[state] = &StateData{
		State:        state,
		PKCEVerifier: pkceVerifier,
		CreatedAt:    time.Now(),
	}
}

// GetState retrieves and deletes state data
func (s *Store) GetState(state string) (*StateData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stateData, exists := s.states[state]
	if !exists {
		return nil, errors.New("state not found")
	}

	// Check if state has expired (10 minutes)
	if time.Now().Sub(stateData.CreatedAt) > 10*time.Minute {
		delete(s.states, state)
		return nil, errors.New("state expired")
	}

	// Delete state after retrieval (one-time use)
	delete(s.states, state)

	return stateData, nil
}

// CleanupExpired removes expired sessions and states
func (s *Store) CleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cleanup expired sessions
	for id, session := range s.sessions {
		if time.Now().After(session.ExpiresAt) {
			delete(s.sessions, id)
		}
	}

	// Cleanup expired states (older than 10 minutes)
	for state, stateData := range s.states {
		if time.Now().Sub(stateData.CreatedAt) > 10*time.Minute {
			delete(s.states, state)
		}
	}
}

// StartCleanupRoutine starts a background goroutine to cleanup expired data
func (s *Store) StartCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			s.CleanupExpired()
		}
	}()
}
