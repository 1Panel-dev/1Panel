package dto

type SearchWithPage struct {
	PageInfo
	Info string `json:"info"`
}

type SearchPageWithType struct {
	PageInfo
	Type string `json:"type"`
	Info string `json:"info"`
}

type SearchPageWithGroup struct {
	PageInfo
	GroupID uint   `json:"groupID"`
	Info    string `json:"info"`
}

type PageInfo struct {
	Page     int `json:"page" validate:"required,number"`
	PageSize int `json:"pageSize" validate:"required,number"`
}

type PageResult struct {
	Total int64       `json:"total"`
	Items interface{} `json:"items"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type Options struct {
	Option string `json:"option"`
}

type OperateByType struct {
	Type string `json:"type"`
}

type OperateByName struct {
	Name string `json:"name"`
}

type OperateByID struct {
	ID uint `json:"id"`
}
type OperateByIDs struct {
	IDs []uint `json:"ids"`
}
