package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `json:"name" gorm:"type:varchar(64);not null;"`
	Password string `json:"password" gorm:"type:varchar(64)"`
	Email    string `json:"email" gorm:"type:varchar(64);not null;"`
}
