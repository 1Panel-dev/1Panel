package dto

type Fail2banBaseInfo struct {
	IsEnable bool   `json:"isEnable"`
	IsActive bool   `json:"isActive"`
	Version  string `json:"version"`

	Port      int    `json:"port"`
	MaxRetry  int    `json:"maxRetry"`
	BanTime   string `json:"banTime"`
	FindTime  string `json:"findTime"`
	BanAction string `json:"banAction"`
	LogPath   string `json:"logPath"`
}

type Fail2banSearch struct {
	PageInfo
	Status string `json:"status" validate:"required,oneof=banned ignore"`
}

type Fail2banUpdate struct {
	Key   string `json:"key" validate:"required,oneof=port banTime findTime maxRetry banAction action logPath"`
	Value string `json:"value"`
}

type Fail2banSet struct {
	IPs     []string `json:"ips"`
	Operate string   `json:"status"  validate:"required,oneof=banned unbanned ignore"`
}
