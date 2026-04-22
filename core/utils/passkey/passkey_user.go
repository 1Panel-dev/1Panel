package passkey

import "github.com/go-webauthn/webauthn/webauthn"

type PasskeyUser struct {
	ID          []byte
	Name        string
	DisplayName string
	Credentials []webauthn.Credential
}

func (u PasskeyUser) WebAuthnID() []byte {
	return u.ID
}

func (u PasskeyUser) WebAuthnName() string {
	return u.Name
}

func (u PasskeyUser) WebAuthnDisplayName() string {
	return u.DisplayName
}

func (u PasskeyUser) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}
