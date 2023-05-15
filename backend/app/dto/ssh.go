package dto

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
