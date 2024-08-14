package model

type Command struct {
	BaseModel
	Name    string `gorm:"unique;not null" json:"name"`
	GroupID uint   `json:"groupID"`
	Command string `gorm:"not null" json:"command"`
}

type RedisCommand struct {
	BaseModel
	Name    string `gorm:"unique;not null" json:"name"`
	Command string `gorm:"not null" json:"command"`
}
