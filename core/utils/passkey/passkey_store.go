package passkey

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/go-webauthn/webauthn/webauthn"
)

const (
	PasskeyUserIDSettingKey      = "PasskeyUserID"
	PasskeyCredentialSettingKey  = "PasskeyCredentials"
	PasskeyMaxCredentials        = 5
	PasskeySessionTTL            = 5 * time.Minute
	PasskeySessionKindLogin      = "login"
	PasskeySessionKindRegister   = "register"
	PasskeyCredentialNameDefault = "Passkey"
)

var passkeySessions = newPasskeySessionStore()

func GetPasskeySessionStore() *passkeySessionStore {
	return passkeySessions
}

type passkeySession struct {
	Kind      string
	Name      string
	Session   webauthn.SessionData
	ExpiresAt time.Time
}

type passkeySessionStore struct {
	mu    sync.Mutex
	items map[string]passkeySession
}

func newPasskeySessionStore() *passkeySessionStore {
	return &passkeySessionStore{items: make(map[string]passkeySession)}
}

func (s *passkeySessionStore) Set(kind, name string, session webauthn.SessionData) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID := generatePasskeySessionID()
	s.items[sessionID] = passkeySession{
		Kind:      kind,
		Name:      name,
		Session:   session,
		ExpiresAt: time.Now().Add(PasskeySessionTTL),
	}
	return sessionID
}

func (s *passkeySessionStore) Get(sessionID string) (passkeySession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[sessionID]
	if !ok {
		return passkeySession{}, false
	}
	if time.Now().After(item.ExpiresAt) {
		delete(s.items, sessionID)
		return passkeySession{}, false
	}
	return item, true
}

func (s *passkeySessionStore) Delete(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, sessionID)
}

func generatePasskeySessionID() string {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return common.RandStr(32)
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}
