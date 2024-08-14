package model

type Group struct {
	BaseModel
	IsDefault bool   `json:"isDefault"`
	Name      string `gorm:"not null" json:"name"`
	Type      string `gorm:"not null" json:"type"`
}
