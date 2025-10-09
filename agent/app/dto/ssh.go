package dto

import "time"

type SSHUpdate struct {
	Key      string `json:"key" validate:"required"`
	OldValue string `json:"oldValue"`
	NewValue string `json:"newValue"`
}

type SSHInfo struct {
	AutoStart              bool   `json:"autoStart"`
	IsExist                bool   `json:"isExist"`
	IsActive               bool   `json:"isActive"`
	Message                string `json:"message"`
	Port                   string `json:"port"`
	ListenAddress          string `json:"listenAddress"`
	PasswordAuthentication string `json:"passwordAuthentication"`
	PubkeyAuthentication   string `json:"pubkeyAuthentication"`
	PermitRootLogin        string `json:"permitRootLogin"`
	UseDNS                 string `json:"useDNS"`
	CurrentUser            string `json:"currentUser"`
}

type RootCertOperate struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	Mode           string `json:"mode"`
	EncryptionMode string `json:"encryptionMode" validate:"required,oneof=rsa ed25519 ecdsa dsa"`
	PassPhrase     string `json:"passPhrase"`
	PublicKey      string `json:"publicKey"`
	PrivateKey     string `json:"privateKey"`
	Description    string `json:"description"`
}
type RootCert struct {
	ID             uint      `json:"id"`
	CreatedAt      time.Time `json:"createdAt"`
	Name           string    `json:"name"`
	EncryptionMode string    `json:"encryptionMode"`
	PassPhrase     string    `json:"passPhrase"`
	PublicKey      string    `json:"publicKey"`
	PrivateKey     string    `json:"privateKey"`
	Description    string    `json:"description"`
}

type GenerateLoad struct {
	EncryptionMode string `json:"encryptionMode" validate:"required,oneof=rsa ed25519 ecdsa dsa"`
}

type SSHConf struct {
	File string `json:"file"`
}
type SearchSSHLog struct {
	PageInfo
	Info   string `json:"info"`
	Status string `json:"Status" validate:"required,oneof=Success Failed All"`
}

type SSHHistory struct {
	Date     time.Time `json:"date"`
	DateStr  string    `json:"dateStr"`
	Area     string    `json:"area"`
	User     string    `json:"user"`
	AuthMode string    `json:"authMode"`
	Address  string    `json:"address"`
	Port     string    `json:"port"`
	Status   string    `json:"status"`
	Message  string    `json:"message"`
}
