package entity

import "github.com/1Panel-dev/1Panel/app/entity/common"

type OperateLog struct {
	common.BaseModel
	Name      string `gorm:"type:varchar(64)"`
	Type      string `gorm:"type:varchar(64)"`
	User      string `gorm:"type:varchar(64)"`
	Path      string `gorm:"type:varchar(64)"`
	IP        string `gorm:"type:varchar(64)"`
	UserAgent string `gorm:"type:varchar(64)"`
	Source    string `gorm:"type:varchar(64)"`
	Detail    string `gorm:"type:longText"`
}
