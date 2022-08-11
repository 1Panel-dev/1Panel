package model

import (
	"time"

	"gorm.io/gorm"
)

type OperationLog struct {
	gorm.Model
	Group  string `gorm:"type:varchar(64)" json:"type"`
	Source string `gorm:"type:varchar(64)" json:"source"`
	Action string `gorm:"type:varchar(64)" json:"action"`

	IP        string `gorm:"type:varchar(64)" json:"ip"`
	Path      string `gorm:"type:varchar(64)" json:"path"`
	Method    string `gorm:"type:varchar(64)" json:"method"`
	UserAgent string `gorm:"type:text(65535)" json:"userAgent"`
	Body      string `gorm:"type:text(65535)" json:"body"`
	Resp      string `gorm:"type:text(65535)" json:"resp"`

	Status       int           `gorm:"type:varchar(64)" json:"status"`
	Latency      time.Duration `gorm:"type:varchar(64)" json:"latency"`
	ErrorMessage string        `gorm:"type:varchar(256)" json:"errorMessage"`

	Detail string `gorm:"type:longText" json:"detail"`
}
