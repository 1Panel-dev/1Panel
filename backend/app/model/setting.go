package model

import "gorm.io/gorm"

type Setting struct {
	gorm.Model
	Key   string `json:"key" gorm:"type:varchar(256);not null;"`
	Value string `json:"value" gorm:"type:varchar(256)"`
	About string `json:"about" gorm:"type:longText"`
}
