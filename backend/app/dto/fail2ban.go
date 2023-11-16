package dto

type Fail2BanBaseInfo struct {
	IsEnable bool   `json:"isEnable"`
	IsActive bool   `json:"isActive"`
	IsExist  bool   `json:"isExist"`
	Version  string `json:"version"`

	MaxRetry  int    `json:"maxRetry"`
	BanTime   string `json:"banTime"`
	FindTime  string `json:"findTime"`
	BanAction string `json:"banAction"`
}

type Fail2BanSearch struct {
	Status string `json:"status" validate:"required,oneof=banned ignore"`
}

type Fail2BanUpdate struct {
	Key   string `json:"key" validate:"required,oneof=port bantime findtime maxretry banaction"`
	Value string `json:"value"`
}

type Fail2BanSet struct {
	IPs     []string `json:"ips"`
	Operate string   `json:"operate"  validate:"required,oneof=banned ignore"`
}
