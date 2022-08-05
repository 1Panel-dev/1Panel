package dto

type PageInfo struct {
	Page     int `json:"page" validate:"required,number"`
	PageSize int `json:"pageSize" validate:"required,number"`
}

type OperationWithName struct {
	Name string `json:"name" validate:"required"`
}

type OperationWithNameAndType struct {
	Name string `json:"name" validate:"required"`
	Type string `json:"type" validate:"required"`
}
