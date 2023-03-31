package request

import "github.com/1Panel-dev/1Panel/backend/app/dto"

type RuntimeSearch struct {
	dto.PageInfo
	Type string `json:"type"`
	Name string `json:"name"`
}

type RuntimeCreate struct {
	AppDetailID uint                   `json:"appDetailId"`
	Name        string                 `json:"name"`
	Params      map[string]interface{} `json:"params"`
	Resource    string                 `json:"resource"`
	Image       string                 `json:"image"`
	Type        string                 `json:"type"`
	Version     string                 `json:"version"`
}

type RuntimeDelete struct {
	ID uint `json:"Id"`
}
