package entity

import "github.com/1Panel-dev/1Panel/app/entity/common"

type User struct {
	common.BaseModel
	Name     string `gorm:"primarykey"`
	Email    string
	Tel      string
	NickName string
	Password string
}
