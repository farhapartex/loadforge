package web

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Session struct {
	Username  string
	ExpiresAt time.Time
}

type SessionStore struct {
	m   sync.Map
	ttl time.Duration
}

func newSessionStore(ttl time.Duration) *SessionStore {
	return &SessionStore{ttl: ttl}
}

func (s *SessionStore) create(username string) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}
	s.m.Store(token, Session{
		Username:  username,
		ExpiresAt: time.Now().Add(s.ttl),
	})
	return token, nil
}

func (s *SessionStore) get(token string) (Session, bool) {
	val, ok := s.m.Load(token)
	if !ok {
		return Session{}, false
	}
	sess := val.(Session)
	if time.Now().After(sess.ExpiresAt) {
		s.m.Delete(token)
		return Session{}, false
	}
	return sess, true
}

func (s *SessionStore) delete(token string) {
	s.m.Delete(token)
}

func (s *SessionStore) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.purgeExpired()
		case <-ctx.Done():
			return
		}
	}
}

func (s *SessionStore) purgeExpired() {
	now := time.Now()
	s.m.Range(func(key, value any) bool {
		if sess, ok := value.(Session); ok && now.After(sess.ExpiresAt) {
			s.m.Delete(key)
		}
		return true
	})
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
