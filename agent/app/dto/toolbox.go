package dto

type NetToolReq struct {
	Tool  string `json:"tool" validate:"required,oneof=ping traceroute dns http"`
	Host  string `json:"host" validate:"required"`
	Count int    `json:"count"`
}

type NetToolRes struct {
	Output string `json:"output"`
}
