package model

type Firewall struct {
	BaseModel

	Type        string `gorm:"not null" json:"type"`
	Port        string `gorm:"not null" json:"port"`
	Protocol    string `gorm:"not null" json:"protocol"`
	Address     string `gorm:"not null" json:"address"`
	Strategy    string `gorm:"not null" json:"strategy"`
	Description string `gorm:"not null" json:"description"`
}

type Forward struct {
	BaseModel

	Protocol   string `gorm:"not null" json:"protocol"`
	Port       string `gorm:"not null" json:"port"`
	TargetIP   string `gorm:"not null" json:"targetIP"`
	TargetPort string `gorm:"not null" json:"targetPort"`
	Interface  string `json:"interface"`
}

type IptablesRule struct {
	BaseModel

	RuleType string `gorm:"not null" json:"ruleType"` // port, address
	Protocol string `json:"protocol"`
	Port     string `json:"port"`
	Strategy string `gorm:"not null" json:"strategy"` // accept, drop, reject
	Address  string `json:"address"`
	Family   string `json:"family"` // ipv4, ipv6
}
