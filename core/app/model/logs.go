package model

import (
	"time"
)

type OperationLog struct {
	BaseModel
	Source string `gorm:"type:varchar(64)" json:"source"`

	IP        string `gorm:"type:varchar(64)" json:"ip"`
	Path      string `gorm:"type:varchar(64)" json:"path"`
	Method    string `gorm:"type:varchar(64)" json:"method"`
	UserAgent string `gorm:"type:varchar(256)" json:"userAgent"`

	Latency time.Duration `gorm:"type:varchar(64)" json:"latency"`
	Status  string        `gorm:"type:varchar(64)" json:"status"`
	Message string        `gorm:"type:varchar(256)" json:"message"`

	DetailZH string `gorm:"type:varchar(256)" json:"detailZH"`
	DetailEN string `gorm:"type:varchar(256)" json:"detailEN"`
}

type LoginLog struct {
	BaseModel
	IP      string `gorm:"type:varchar(64)" json:"ip"`
	Address string `gorm:"type:varchar(64)" json:"address"`
	Agent   string `gorm:"type:varchar(256)" json:"agent"`
	Status  string `gorm:"type:varchar(64)" json:"status"`
	Message string `gorm:"type:longText" json:"message"`
}
