package dto

type SearchWithPage struct {
	PageInfo
	Info string `json:"info"`
}

type PageInfo struct {
	Page     int `json:"page" validate:"required,number"`
	PageSize int `json:"pageSize" validate:"required,number"`
}
