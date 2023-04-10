package model

type Host struct {
	BaseModel

	GroupID          uint   `gorm:"type:decimal;not null" json:"group_id"`
	Name             string `gorm:"type:varchar(64);not null" json:"name"`
	Addr             string `gorm:"type:varchar(16);not null" json:"addr"`
	Port             int    `gorm:"type:decimal;not null" json:"port"`
	User             string `gorm:"type:varchar(64);not null" json:"user"`
	AuthMode         string `gorm:"type:varchar(16);not null" json:"authMode"`
	Password         string `gorm:"type:varchar(64)" json:"password"`
	PrivateKey       string `gorm:"type:varchar(256)" json:"privateKey"`
	PassPhrase       string `gorm:"type:varchar(256)" json:"passPhrase"`
	RememberPassword bool   `json:"rememberPassword"`

	Description string `gorm:"type:varchar(256)" json:"description"`
}
