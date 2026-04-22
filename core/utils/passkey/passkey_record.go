package passkey

import "github.com/go-webauthn/webauthn/webauthn"

type PasskeyCredentialRecord struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	CreatedAt  string              `json:"createdAt"`
	LastUsedAt string              `json:"lastUsedAt"`
	FlagsValue uint8               `json:"flagsValue"`
	Credential webauthn.Credential `json:"credential"`
}
