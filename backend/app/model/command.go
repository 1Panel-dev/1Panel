package model

import "gorm.io/gorm"

type Command struct {
	gorm.Model
	Name    string `gorm:"type:varchar(64));unique;not null" json:"name"`
	Command string `gorm:"type:varchar(256);unique;not null" json:"command"`
}
