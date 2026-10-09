package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

const SessionCookieName = "session"

var ErrInvalidSession = errors.New("invalid or expired session")

type Session struct {
	ExecID    int
	ExpiresAt time.Time
}

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[[sha256.Size]byte]Session
	lifetime time.Duration
	now      func() time.Time
}

func NewSessionStore(lifetime time.Duration) *SessionStore {
	return &SessionStore{
		sessions: make(map[[sha256.Size]byte]Session),
		lifetime: lifetime,
		now:      time.Now,
	}
}

func (s *SessionStore) Create(execID int) (string, time.Time, error) {
	if execID <= 0 {
		return "", time.Time{}, errors.New("invalid exec ID")
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", time.Time{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	tokenHash := sha256.Sum256([]byte(token))
	now := s.now()
	expiresAt := now.Add(s.lifetime)

	s.mu.Lock()
	for existingTokenHash, session := range s.sessions {
		if !now.Before(session.ExpiresAt) {
			delete(s.sessions, existingTokenHash)
		}
	}
	s.sessions[tokenHash] = Session{ExecID: execID, ExpiresAt: expiresAt}
	s.mu.Unlock()

	return token, expiresAt, nil
}

func (s *SessionStore) Get(token string) (Session, error) {
	if token == "" {
		return Session{}, ErrInvalidSession
	}
	tokenHash := sha256.Sum256([]byte(token))

	s.mu.RLock()
	session, found := s.sessions[tokenHash]
	s.mu.RUnlock()
	if !found {
		return Session{}, ErrInvalidSession
	}
	if !s.now().Before(session.ExpiresAt) {
		s.mu.Lock()
		delete(s.sessions, tokenHash)
		s.mu.Unlock()
		return Session{}, ErrInvalidSession
	}
	return session, nil
}

func (s *SessionStore) Delete(token string) {
	if token == "" {
		return
	}
	tokenHash := sha256.Sum256([]byte(token))
	s.mu.Lock()
	delete(s.sessions, tokenHash)
	s.mu.Unlock()
}

var DefaultSessionStore = NewSessionStore(24 * time.Hour)
