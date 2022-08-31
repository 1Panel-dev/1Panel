package dto

type GroupOperate struct {
	Name string `json:"name" validate:"required"`
	Type string `json:"type" validate:"required"`
}

type GroupSearch struct {
	Type string `json:"type" validate:"required"`
}

type GroupInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}
