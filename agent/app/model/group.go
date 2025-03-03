package model

type Group struct {
	BaseModel
	IsDefault bool   `json:"isDefault"`
	Name      string `json:"name"`
	Type      string `json:"type"`
}
