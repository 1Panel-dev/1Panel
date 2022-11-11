package dto

import "github.com/1Panel-dev/1Panel/backend/app/model"

type WebsiteAcmeAccountDTO struct {
	model.WebsiteAcmeAccount
}

type WebsiteAcmeAccountCreate struct {
	Email string `json:"email" validate:"required"`
}
