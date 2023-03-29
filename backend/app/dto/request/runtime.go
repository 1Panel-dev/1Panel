package request

import "github.com/1Panel-dev/1Panel/backend/app/dto"

type RuntimeSearch struct {
	dto.PageInfo
	Type string `json:"type"`
	Name string `json:"name"`
}
