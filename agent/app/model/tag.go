package model

type Tag struct {
	BaseModel
	Key          string `json:"key" gorm:"not null"`
	Name         string `json:"name" gorm:"not null"`
	Sort         int    `json:"sort" gorm:"not null;default:1"`
	Translations string `json:"translations"`
}
