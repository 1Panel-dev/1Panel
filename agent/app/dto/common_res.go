package dto

type PageResult struct {
	Total int64       `json:"total"`
	Items interface{} `json:"items"`
}

type Response struct {
	Code      int         `json:"code"`
	ErrorCode string      `json:"errorCode,omitempty"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
}

type Options struct {
	Option string `json:"option"`
}
