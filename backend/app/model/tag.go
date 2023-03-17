package model

type Tag struct {
	BaseModel
	Key  string `json:"key" gorm:"type:varchar(64);not null"`
	Name string `json:"name" gorm:"type:varchar(64);not null"`
}
