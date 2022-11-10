package dto

import "github.com/1Panel-dev/1Panel/backend/app/model"

type WebsiteDnsAccountDTO struct {
	model.WebsiteDnsAccount
	Authorization map[string]string `json:"authorization"`
}

type WebsiteDnsAccountCreate struct {
	Name          string            `json:"name" validate:"required"`
	Type          string            `json:"type" validate:"required"`
	Authorization map[string]string `json:"authorization" validate:"required"`
}

type WebsiteDnsAccountUpdate struct {
	ID            uint              `json:"id" validate:"required"`
	Name          string            `json:"name" validate:"required"`
	Type          string            `json:"type" validate:"required"`
	Authorization map[string]string `json:"authorization" validate:"required"`
}
