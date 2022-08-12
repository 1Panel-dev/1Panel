package dto

import (
	"time"
)

type OperationLogBack struct {
	ID     uint   `json:"id"`
	Group  string `json:"group"`
	Source string `json:"source"`
	Action string `json:"action"`

	IP        string `json:"ip"`
	Path      string `json:"path"`
	Method    string `json:"method"`
	UserAgent string `json:"userAgent"`
	Body      string `json:"body"`
	Resp      string `json:"resp"`

	Status       int           `json:"status"`
	Latency      time.Duration `json:"latency"`
	ErrorMessage string        `json:"errorMessage"`

	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"createdAt"`
}
