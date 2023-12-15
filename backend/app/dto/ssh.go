package dto

import "time"

type SSHUpdate struct {
	Key      string `json:"key" validate:"required"`
	OldValue string `json:"oldValue"`
	NewValue string `json:"newValue"`
}

type SSHInfo struct {
	AutoStart              bool   `json:"autoStart"`
	Status                 string `json:"status"`
	Message                string `json:"message"`
	Port                   string `json:"port"`
	ListenAddress          string `json:"listenAddress"`
	PasswordAuthentication string `json:"passwordAuthentication"`
	PubkeyAuthentication   string `json:"pubkeyAuthentication"`
	PermitRootLogin        string `json:"permitRootLogin"`
	UseDNS                 string `json:"useDNS"`
}

type GenerateSSH struct {
	EncryptionMode string `json:"encryptionMode" validate:"required,oneof=rsa ed25519 ecdsa dsa"`
	Password       string `json:"password"`
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
type SSHLog struct {
	Logs            []SSHHistory `json:"logs"`
	TotalCount      int          `json:"totalCount"`
	SuccessfulCount int          `json:"successfulCount"`
	FailedCount     int          `json:"failedCount"`
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
