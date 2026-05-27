package dto

import (
	"time"
)

type OperationLog struct {
	ID        uint   `json:"id"`
	Source    string `json:"source"`
	User      string `json:"user"`
	Node      string `json:"node"`
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

type SearchOpLogWithPage struct {
	PageInfo
	Source    string `json:"source"`
	Status    string `json:"status"`
	Node      string `json:"node"`
	Operation string `json:"operation"`
}

type SearchLgLogWithPage struct {
	PageInfo
	Info      string    `json:"info"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type LoginLog struct {
	ID        uint      `json:"id"`
	IP        string    `json:"ip"`
	User      string    `json:"user"`
	Address   string    `json:"address"`
	Agent     string    `json:"agent"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

type CleanLog struct {
	LogType string `json:"logType" validate:"required,oneof=login operation"`
}
