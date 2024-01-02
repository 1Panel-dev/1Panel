package request

import "github.com/1Panel-dev/1Panel/backend/app/dto"

type PHPExtensionsSearch struct {
	dto.PageInfo
	All bool `json:"all"`
}

type PHPExtensionsCreate struct {
	Name       string `json:"name" validate:"required"`
	Extensions string `json:"extensions" validate:"required"`
}

type PHPExtensionsUpdate struct {
	ID         uint   `json:"id" validate:"required"`
	Extensions string `json:"extensions" validate:"required"`
}

type PHPExtensionsDelete struct {
	ID uint `json:"id" validate:"required"`
}
