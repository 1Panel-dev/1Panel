package dto

import "time"

type SSHInfo struct {
	Port                   string `json:"port" validate:"required,number,max=65535,min=1"`
	ListenAddress          string `json:"listenAddress"`
	PasswordAuthentication string `json:"passwordAuthentication" validate:"required,oneof=yes no"`
	PubkeyAuthentication   string `json:"pubkeyAuthentication" validate:"required,oneof=yes no"`
	PermitRootLogin        string `json:"permitRootLogin" validate:"required,oneof=yes no without-password forced-commands-only"`
	UseDNS                 string `json:"useDNS" validate:"required,oneof=yes no"`
}

type GenerateSSH struct {
	EncryptionMode string `json:"encryptionMode" validate:"required,oneof=rsa ed25519 ecdsa dsa"`
	Password       string `json:"password"`
}

type GenerateLoad struct {
	EncryptionMode string `json:"encryptionMode" validate:"required,oneof=rsa ed25519 ecdsa dsa"`
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
	Belong   string    `json:"belong"`
	User     string    `json:"user"`
	AuthMode string    `json:"authMode"`
	Address  string    `json:"address"`
	Port     string    `json:"port"`
	Status   string    `json:"status"`
	Message  string    `json:"message"`
}
