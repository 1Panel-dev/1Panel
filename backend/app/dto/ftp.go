package dto

import (
	"time"
)

type FtpInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	User        string `json:"user"`
	Password    string `json:"password"`
	Path        string `json:"path"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

type FtpBaseInfo struct {
	IsActive bool `json:"isActive"`
	IsExist  bool `json:"isExist"`
}

type FtpLogSearch struct {
	PageInfo
	User      string `json:"user"`
	Operation string `json:"operation"`
}

type FtpCreate struct {
	User        string `json:"user" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Path        string `json:"path" validate:"required"`
	Description string `json:"description"`
}

type FtpUpdate struct {
	ID uint `json:"id"`

	Password    string `json:"password" validate:"required"`
	Path        string `json:"path" validate:"required"`
	Status      string `json:"status"`
	Description string `json:"description"`
}
