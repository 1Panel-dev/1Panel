package passkey

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/1Panel-dev/1Panel/core/utils/ttlstore"
	"github.com/go-webauthn/webauthn/webauthn"
)

const (
	PasskeyUserIDSettingKey       = "PasskeyUserID"
	PasskeyCredentialSettingKey   = "PasskeyCredentials"
	PasskeyMaxCredentials         = 5
	PasskeySessionTTL             = 5 * time.Minute
	PasskeySessionKindLogin       = "login"
	PasskeySessionKindRegister    = "register"
	PasskeyCredentialNameDefault  = "Passkey"
	PasskeySessionStoreMaxEntries = 1024
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
	store *ttlstore.Store[passkeySession]
}

func newPasskeySessionStore() *passkeySessionStore {
	return &passkeySessionStore{
		store: ttlstore.New[passkeySession](PasskeySessionTTL, PasskeySessionStoreMaxEntries, generatePasskeySessionID),
	}
}

func (s *passkeySessionStore) Set(kind, name string, session webauthn.SessionData) string {
	return s.store.Set(passkeySession{
		Kind:    kind,
		Name:    name,
		Session: session,
	})
}

func (s *passkeySessionStore) Get(sessionID string) (passkeySession, bool) {
	return s.store.Get(sessionID)
}

func (s *passkeySessionStore) Delete(sessionID string) {
	s.store.Delete(sessionID)
}

func generatePasskeySessionID() string {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return common.RandStr(32)
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}
