package model

import (
	"time"
)

type OperationLog struct {
	BaseModel
	Source    string `json:"source"`
	IP        string `json:"ip"`
	Node      string `json:"node"`
	Path      string `json:"path"`
	Method    string `json:"method"`
	UserAgent string `json:"userAgent"`

	Latency time.Duration `json:"latency"`
	Status  string        `json:"status"`
	Message string        `json:"message"`

	DetailZH string `json:"detailZH"`
	DetailEN string `json:"detailEN"`
}

type LoginLog struct {
	BaseModel
	IP      string `json:"ip"`
	Address string `json:"address"`
	Agent   string `json:"agent"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
