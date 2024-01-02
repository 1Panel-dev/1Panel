package model

type PHPExtensions struct {
	BaseModel
	Name       string ` json:"name" gorm:"not null"`
	Extensions string `json:"extensions" gorm:"not null"`
}
