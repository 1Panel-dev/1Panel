package model

import (
	"time"
)

type OperationLog struct {
	BaseModel
	Group  string `gorm:"type:varchar(64)" json:"group"`
	Source string `gorm:"type:varchar(64)" json:"source"`
	Action string `gorm:"type:varchar(64)" json:"action"`

	IP        string `gorm:"type:varchar(64)" json:"ip"`
	Path      string `gorm:"type:varchar(64)" json:"path"`
	Method    string `gorm:"type:varchar(64)" json:"method"`
	UserAgent string `gorm:"type:varchar(256)" json:"userAgent"`
	Body      string `gorm:"type:longText" json:"body"`
	Resp      string `gorm:"type:longText" json:"resp"`

	Latency time.Duration `gorm:"type:varchar(64)" json:"latency"`

	Detail string `gorm:"type:longText" json:"detail"`
}
