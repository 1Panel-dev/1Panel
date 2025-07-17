package model

type AlertConfig struct {
	BaseModel
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Config string `json:"config"`
}
