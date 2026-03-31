package auth

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

const (
	MFASessionTTL             = 5 * time.Minute
	MFASessionStoreMaxEntries = 1024
	MFASessionMaxFailures     = 5
)

var mfaSessions = newMFASessionStore()

func GetMFASessionStore() *mfaSessionStore {
	return mfaSessions
}

type mfaSession struct {
	Name      string
	Entrance  string
	IP        string
	Failures  int
	ExpiresAt time.Time
}

type mfaSessionStore struct {
	mu    sync.Mutex
	items map[string]mfaSession
}

func newMFASessionStore() *mfaSessionStore {
	return &mfaSessionStore{items: make(map[string]mfaSession)}
}

func (s *mfaSessionStore) Set(name, entrance, ip string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupExpiredLocked()
	if len(s.items) >= MFASessionStoreMaxEntries {
		s.removeOldestLocked()
	}

	sessionID := generateMFASessionID()
	s.items[sessionID] = mfaSession{
		Name:      name,
		Entrance:  entrance,
		IP:        ip,
		ExpiresAt: time.Now().Add(MFASessionTTL),
	}
	return sessionID
}

func (s *mfaSessionStore) Get(sessionID string) (mfaSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[sessionID]
	if !ok {
		return mfaSession{}, false
	}
	if time.Now().After(item.ExpiresAt) {
		delete(s.items, sessionID)
		return mfaSession{}, false
	}
	return item, true
}

func (s *mfaSessionStore) Delete(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, sessionID)
}

func (s *mfaSessionStore) RecordFailure(sessionID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[sessionID]
	if !ok {
		return false
	}
	if time.Now().After(item.ExpiresAt) {
		delete(s.items, sessionID)
		return false
	}

	item.Failures++
	if item.Failures >= MFASessionMaxFailures {
		delete(s.items, sessionID)
		return false
	}

	s.items[sessionID] = item
	return true
}

func (s *mfaSessionStore) cleanupExpiredLocked() {
	now := time.Now()
	for id, item := range s.items {
		if now.After(item.ExpiresAt) {
			delete(s.items, id)
		}
	}
}

func (s *mfaSessionStore) removeOldestLocked() {
	var oldestID string
	var oldestTime time.Time
	for id, item := range s.items {
		if oldestID == "" || item.ExpiresAt.Before(oldestTime) {
			oldestID = id
			oldestTime = item.ExpiresAt
		}
	}
	if oldestID != "" {
		delete(s.items, oldestID)
	}
}

func generateMFASessionID() string {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return generateFallbackMFASessionID()
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

func generateFallbackMFASessionID() string {
	return base64.RawURLEncoding.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
}
