package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/1Panel-dev/1Panel/core/utils/ttlstore"
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
	Name                    string
	Entrance                string
	IP                      string
	AuthSource              string
	AuthSourceID            uint
	AuthSourceConfigVersion uint64
	Failures                int
	ExpiresAt               time.Time
}

type mfaSessionStore struct {
	store *ttlstore.Store[mfaSession]
}

func newMFASessionStore() *mfaSessionStore {
	return &mfaSessionStore{
		store: ttlstore.New[mfaSession](MFASessionTTL, MFASessionStoreMaxEntries, generateMFASessionID),
	}
}

func (s *mfaSessionStore) Set(name, entrance, ip string) string {
	return s.store.Set(mfaSession{
		Name:     name,
		Entrance: entrance,
		IP:       ip,
	})
}

func (s *mfaSessionStore) SetWithAuthSource(
	name, entrance, ip, authSource string,
	authSourceID uint,
	authSourceConfigVersion uint64,
) string {
	return s.store.Set(mfaSession{
		Name:                    name,
		Entrance:                entrance,
		IP:                      ip,
		AuthSource:              authSource,
		AuthSourceID:            authSourceID,
		AuthSourceConfigVersion: authSourceConfigVersion,
	})
}

func (s *mfaSessionStore) Get(sessionID string) (mfaSession, bool) {
	return s.store.Get(sessionID)
}

func (s *mfaSessionStore) Delete(sessionID string) {
	s.store.Delete(sessionID)
}

func (s *mfaSessionStore) RecordFailure(sessionID string) int {
	failures := 0
	ok := s.store.Update(sessionID, func(item *mfaSession) bool {
		item.Failures++
		failures = item.Failures
		return item.Failures < MFASessionMaxFailures
	})
	if !ok {
		return 0
	}
	return failures
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
