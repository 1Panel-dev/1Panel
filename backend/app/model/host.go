package model

import "gorm.io/gorm"

type Host struct {
	gorm.Model
	Group      string `gorm:"type:varchar(64);not null" json:"group"`
	Name       string `gorm:"type:varchar(64);unique;not null" json:"name"`
	Addr       string `gorm:"type:varchar(16);unique;not null" json:"addr"`
	Port       int    `gorm:"type:varchar(5);not null" json:"port"`
	User       string `gorm:"type:varchar(64);not null" json:"user"`
	AuthMode   string `gorm:"type:varchar(16);not null" json:"authMode"`
	Password   string `gorm:"type:varchar(64)" json:"password"`
	PrivateKey string `gorm:"type:varchar(256)" json:"privateKey"`

	Description string `gorm:"type:varchar(256)" json:"description"`
}
