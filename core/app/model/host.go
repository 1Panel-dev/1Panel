package model

type Host struct {
	BaseModel

	GroupID          uint   `gorm:"not null" json:"group_id"`
	Name             string `gorm:"not null" json:"name"`
	Addr             string `gorm:"not null" json:"addr"`
	Port             int    `gorm:"not null" json:"port"`
	User             string `gorm:"not null" json:"user"`
	AuthMode         string `gorm:"not null" json:"authMode"`
	Password         string `json:"password"`
	PrivateKey       string `json:"privateKey"`
	PassPhrase       string `json:"passPhrase"`
	RememberPassword bool   `json:"rememberPassword"`

	Description string `json:"description"`
}
