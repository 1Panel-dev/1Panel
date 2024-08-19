package model

type Command struct {
	BaseModel
	Type    string `gorm:"not null" json:"type"`
	Name    string `gorm:"not null" json:"name"`
	GroupID uint   `gorm:"not null" json:"groupID"`
	Command string `gorm:"not null" json:"command"`
}
