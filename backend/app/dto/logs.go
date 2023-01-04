package dto

import (
	"time"
)

type OperationLog struct {
	ID    uint   `json:"id"`
	Group string `json:"group"`

	IP        string `json:"ip"`
	Path      string `json:"path"`
	Method    string `json:"method"`
	UserAgent string `json:"userAgent"`

	Latency time.Duration `json:"latency"`
	Status  string        `json:"status"`
	Message string        `json:"message"`

	DetailZH  string    `json:"detailZH"`
	DetailEN  string    `json:"detailEN"`
	CreatedAt time.Time `json:"createdAt"`
}

type LoginLog struct {
	ID        uint      `json:"id"`
	IP        string    `json:"ip"`
	Address   string    `json:"address"`
	Agent     string    `json:"agent"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

type CleanLog struct {
	LogType string `json:"logType" validate:"required,oneof=login operation"`
}
