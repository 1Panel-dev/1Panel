package dto

type SSHInfo struct {
	Port                   string `json:"port"`
	ListenAddress          string `json:"listenAddress"`
	PasswordAuthentication string `json:"passwordAuthentication"`
	PubkeyAuthentication   string `json:"pubkeyAuthentication"`
	PermitRootLogin        string `json:"permitRootLogin"`
	UseDNS                 string `json:"useDNS"`
}

type GenerateSSH struct {
	EncryptionMode string `json:"encryptionMode"`
	Password       string `json:"password"`
}
