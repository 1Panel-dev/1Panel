package model

type RootCert struct {
	BaseModel
	Name           string `json:"name" gorm:"not null;"`
	EncryptionMode string `json:"encryptionMode"`
	PassPhrase     string `json:"passPhrase"`
	PublicKeyPath  string `json:"publicKeyPath"`
	PrivateKeyPath string `json:"privateKeyPath"`
	Description    string `json:"description"`
}
