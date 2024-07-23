package model

type Command struct {
	BaseModel
	Name    string `gorm:"type:varchar(64);unique;not null" json:"name"`
	GroupID uint   `gorm:"type:decimal" json:"groupID"`
	Command string `gorm:"type:varchar(256);not null" json:"command"`
}

type RedisCommand struct {
	BaseModel
	Name    string `gorm:"type:varchar(64);unique;not null" json:"name"`
	Command string `gorm:"type:varchar(256);not null" json:"command"`
}
