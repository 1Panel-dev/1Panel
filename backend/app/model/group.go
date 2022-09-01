package model

type Group struct {
	BaseModel
	Name string `gorm:"type:varchar(64);not null" json:"name"`
	Type string `gorm:"type:varchar(16);not null" json:"type"`
}
