package dto

import "github.com/1Panel-dev/1Panel/backend/app/model"

type WebsiteSSLDTO struct {
	model.WebSiteSSL
}

type WebsiteSSLCreate struct {
	Name          string            `json:"name" validate:"required"`
	Type          string            `json:"type" validate:"required"`
	Authorization map[string]string `json:"authorization" validate:"required"`
}
